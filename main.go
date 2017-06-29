package main

import (
	"container/heap"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var Config struct {
	Port           int
	WWWDir         string
	SessionSecret  string
	CookieName     string
	SessionSeconds int
}

type State struct {
	Areas  []*Area
	Rooms  []*Room
	Events Queue
}

func main() {
	// load config
	switch len(os.Args) {
	case 1:
		//loadConfig("/etc/gruffles/config.json")
		Config.Port = 9001
	case 2:
		//loadConfig(os.Args[1])
	default:
		log.Fatalf("Usage: %s [<config file>]", os.Args[0])
	}
	rand.Seed(time.Now().UnixNano())

	// load worlds
	paths, err := filepath.Glob("areas/*.json")
	if err != nil {
		log.Fatalf("glob error: %v", err)
	}

	state := new(State)
	state.Areas, state.Rooms = LoadAreas(paths)
	q := make(chan Event)
	state.Events = q

	SetupCommands()

	// start listening for connections
	go listenForPlayerConnections(Config.Port, q)

	// start the main loop
	mainEventLoop(state, q)
}

var directions = []string{"north", "east", "south", "west", "up", "down"}
var deltaX = []int{0, 1, 0, -1, 0, 0}
var deltaY = []int{1, 0, -1, 0, 0, 0}

type Event struct {
	When time.Time
	What func(*State)
}

type Queue chan<- Event

func (q Queue) Schedule(f func(*State), delay time.Duration) {
	e := Event{When: time.Now().Add(delay), What: f}
	select {
	case q <- e:
	default:
		go func() { q <- e }()
	}
}

type EventSlice []Event

func (e EventSlice) Len() int            { return len(e) }
func (e EventSlice) Swap(i, j int)       { e[i], e[j] = e[j], e[i] }
func (e EventSlice) Less(i, j int) bool  { return e[i].When.Before(e[j].When) }
func (e *EventSlice) Push(x interface{}) { *e = append(*e, x.(Event)) }
func (e *EventSlice) Pop() interface{} {
	item := (*e)[len(*e)-1]
	*e = (*e)[:len(*e)-1]
	return item
}

func startEventQueue() Queue {
	incoming := make(chan Event, 100)
	q := EventSlice{}
	state := &State{}

	// create a timer that is not running
	timer := time.NewTimer(time.Hour)
	timerTarget := time.Now()
	if !timer.Stop() {
		<-timer.C
	}
	timerRunning := false

	// run events in a single worker goroutine
	run := make(chan func(*State))
	finished := make(chan struct{})
	eventRunning := false
	go func() {
		for f := range run {
			f(state)
			finished <- struct{}{}
		}
	}()

	// main event loop
	go func() {
		for {
			// sleep until something happens
			select {
			case e := <-incoming:
				heap.Push(&q, e)
			case <-finished:
				eventRunning = false
			case <-timer.C:
				timerRunning = false
			}

			// do we just need to keep waiting?
			if eventRunning || len(q) == 0 {
				continue
			}

			now := time.Now()

			if !q[0].When.After(now) {
				// run an event now
				eventRunning = true
				e := heap.Pop(&q).(Event)
				run <- e.What

				// no timer when an event is running
				if timerRunning {
					timerRunning = false
					if !timer.Stop() {
						<-timer.C
					}
				}

				continue
			}

			if !timerRunning {
				// start the timer
				timerRunning = true
				timerTarget = q[0].When
				timer.Reset(timerTarget.Sub(now))
			} else if q[0].When.Before(timerTarget) {
				// move the timer up
				if !timer.Stop() {
					<-timer.C
				}
				timerTarget = q[0].When
				timer.Reset(timerTarget.Sub(now))
			}
		}
	}()

	return incoming
}

func mainEventLoop(state *State, incoming <-chan Event) {
	q := EventSlice{}

	// create a timer that is not running
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	timerRunning := false

	// main event loop
	for {
		// sleep until something happens
		select {
		case e := <-incoming:
			heap.Push(&q, e)
		case <-timer.C:
			timerRunning = false
		}

		// nothing to do?
		if len(q) == 0 {
			continue
		}

		// is there an event ready to run?
		now := time.Now()
		if !q[0].When.After(now) {
			// cancel pending timer if any
			if timerRunning {
				if !timer.Stop() {
					<-timer.C
				}
				timerRunning = false
			}

			// run an event now
			e := heap.Pop(&q).(Event)
			e.What(state)
		}

		if !timerRunning && len(q) > 0 {
			timer.Reset(time.Until(q[0].When))
			timerRunning = true
		}

	}
}
