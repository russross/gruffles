package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxPlayerOutgoingQueueLength = 1000
	fastCommandDelay             = 500 * time.Millisecond
)

type Player struct {
	outgoingQueue    []Msg
	outgoingNotEmpty sync.Cond
}

type Request struct {
	Command string `json:"cmd"`
}

type MsgType string

const (
	MsgSocial      MsgType = "social"
	MsgCombat              = "combat"
	MsgEnvironment         = "environment"
	MsgError               = "error"
	MsgMap                 = "map"
)

type Msg struct {
	Type    MsgType `json:"type"`
	Message string  `json:"msg"`
}

func (p *Player) Send(msg Msg) {
	p.outgoingNotEmpty.L.Lock()
	defer p.outgoingNotEmpty.L.Unlock()

	// nil queue means the connection is closed/closing
	if p.outgoingQueue == nil {
		return
	}

	// add the response to the queue
	p.outgoingQueue = append(p.outgoingQueue, msg)

	// if the queue is overflowing, truncate it, keeping the most recent
	if len(p.outgoingQueue) > maxPlayerOutgoingQueueLength {
		log.Printf("truncating outgoing queue for player ???")
		p.outgoingQueue = p.outgoingQueue[len(p.outgoingQueue)-maxPlayerOutgoingQueueLength:]
	}

	// wake up the goroutine that delivers messages
	p.outgoingNotEmpty.Signal()
}

func listenForPlayerConnections(port int, q Queue) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/play/", http.StripPrefix("/play/", http.FileServer(http.Dir(filepath.Join(wd, "client")))))

	log.Printf("listening for incoming connections on port %d", port)
	http.HandleFunc("/server", func(w http.ResponseWriter, r *http.Request) {
		HandleIncommingConnection(w, r, q)
	})
	if err = http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil); err != nil {
		log.Printf("error listening for connections: %v", err)
	}
}

func HandleIncommingConnection(w http.ResponseWriter, r *http.Request, q Queue) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		//EnableCompression: true,
	}
	socket, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("websocket error: %v", err)
		return
	}

	// TODO: get user, character, and auth token from headers

	player := &Player{
		outgoingQueue: []Msg{},
	}
	player.outgoingNotEmpty.L = new(sync.Mutex)
	var mob *Mob
	ready := make(chan struct{})

	q.Schedule(func(state *State) {
		now := time.Now()
		mob = &Mob{
			Name:             "Gnoric",
			Location:         state.Rooms[3001],
			StartLocation:    state.Rooms[3001],
			Visited:          make([]bool, len(state.Rooms)),
			State:            StateStanding,
			SlowBlockedUntil: now,
			FastBlockedUntil: now,
			Player:           player,
		}
		for i := 0; i < len(mob.Visited); i++ {
			if state.Rooms[i] != nil {
				mob.Visited[i] = true
			}
		}
		ready <- struct{}{}
	}, 0)
	<-ready
	q.Schedule(func(state *State) {
		CmdLook(state, mob, "")
	}, 0)

	// a goroutine that sends messages to the player
	go func() {
		for {
			player.outgoingNotEmpty.L.Lock()

			// wait for data to transmit
			for len(player.outgoingQueue) == 0 && player.outgoingQueue != nil {
				player.outgoingNotEmpty.Wait()
			}

			if player.outgoingQueue == nil {
				// our signal to quit
				player.outgoingNotEmpty.L.Unlock()
				socket.Close()
				break
			}

			// get the item from the queue
			elt := player.outgoingQueue[0]
			player.outgoingQueue = player.outgoingQueue[1:]
			player.outgoingNotEmpty.L.Unlock()

			if err := socket.WriteJSON(&elt); err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") ||
					strings.Contains(err.Error(), "close 1005") {
					// websocket closed
				} else {
					log.Printf("websocket write error: %v", err)
					socket.WriteControl(websocket.CloseMessage, nil, time.Now().Add(5*time.Second))
					socket.Close()
				}

				// stop accumulating messages in the queue
				player.outgoingNotEmpty.L.Lock()
				player.outgoingQueue = nil
				player.outgoingNotEmpty.L.Unlock()
				break
			}
		}
	}()

	// the main goroutine reads commands from the player
	for {
		req := new(Request)
		if err := socket.ReadJSON(req); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") ||
				strings.Contains(err.Error(), "close 1005") {
				// websocket closed
			} else {
				log.Printf("websocket read error: %v", err)
				socket.WriteControl(websocket.CloseMessage, nil, time.Now().Add(5*time.Second))
				socket.Close()
			}

			// close the queue of outgoing messages
			player.outgoingNotEmpty.L.Lock()
			player.outgoingQueue = nil
			player.outgoingNotEmpty.Signal()
			player.outgoingNotEmpty.L.Unlock()
			break
		}

		// parse and enqueue the command
		cmd, rest := ParseCommand(req.Command)
		if cmd == nil {
			player.Send(Msg{Type: MsgError, Message: "Huh?"})
			continue
		}

		q.Schedule(func(state *State) {
			cmd.Execute(state, mob, rest)
		}, 0)
	}

	// just to be sure
	socket.Close()
}
