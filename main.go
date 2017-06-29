package main

import (
	"container/heap"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/russross/meddler"
)

var Config struct {
	Hostname         string `json:"hostname"`
	Database         string `json:"database"`
	LetsEncryptCache string `json:"letsEncryptCache"`
	LetsEncryptEmail string `json:"letsEncryptEmail"`
	ClientDir        string `json:"clientDir"`
	SessionSecret    string `json:"sessionSecret"`
	SessionSeconds   int    `json:"sessionSeconds"`
	CookieName       string `json:"cookieName"`
}

type State struct {
	Areas  []*Area
	Rooms  []*Room
	Events Queue
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// load config
	var configFile string
	switch len(os.Args) {
	case 1:
		configFile = "/etc/gruffles/config.json"
	case 2:
		configFile = os.Args[1]
	default:
		log.Fatalf("Usage: %s [<config file>]", os.Args[0])
	}
	Config.LetsEncryptCache = "/etc/gruffles"
	if raw, err := ioutil.ReadFile(configFile); err != nil {
		log.Fatalf("loading config file: %v", err)
	} else if err := json.Unmarshal(raw, &Config); err != nil {
		log.Fatalf("parsing config file: %v", err)
	}
	Config.SessionSecret = unBase64(Config.SessionSecret)
	Config.CookieName = "gruffles"
	Config.SessionSeconds = 90*24*60*60 - 3*60*60
	if Config.Hostname == "" {
		log.Fatalf("cannot run with no hostname in the config file")
	}
	if Config.Database == "" {
		log.Fatalf("cannot run with no database path in the config file")
	}
	if Config.LetsEncryptEmail == "" {
		log.Fatalf("cannot run with no letsEncryptEmail in the config file")
	}
	if Config.ClientDir == "" {
		log.Fatalf("cannot run with no clientDir in the config file")
	}
	if Config.SessionSecret == "" {
		log.Fatalf("cannot run with no sessionSeconds in the config file")
	}

	meddler.Default = meddler.SQLite
	dbPath := Config.Database + "?_busy_timeout=10000&_loc=auto&_foreign_keys=1"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}

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

	// listen for player connections
	http.HandleFunc("/server", func(w http.ResponseWriter, r *http.Request) {
		HandleIncommingConnection(w, r, q)
	})

	// listen for API requests
	server := setupAPI(db)

	// start the server
	go func() {
		log.Printf("accepting https connections")
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("ListenAndServeTLS: %v", err)
		}
	}()

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
