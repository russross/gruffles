package main

import (
	"bytes"
	"time"
)

const RecallLocation = 3001

func CmdLook(state *State, mob *Mob, cmd string) time.Duration {
	if cmd != "" {
		mob.Send(MsgEnvironment, "I don't know how to look at that\n")
		return 0
	}

	mob.Send(MsgEnvironment, mob.Location.GetDescription())
	mob.Send(MsgMap, GetMap(state, mob.Location, mob.Visited))

	return 0
}

const (
	DirNorth int = iota
	DirEast
	DirSouth
	DirWest
	DirUp
	DirDown
)

func CmdNorth(state *State, mob *Mob, cmd string) time.Duration {
	return cmdDirection(state, mob, cmd, DirNorth)
}

func CmdEast(state *State, mob *Mob, cmd string) time.Duration {
	return cmdDirection(state, mob, cmd, DirEast)
}

func CmdSouth(state *State, mob *Mob, cmd string) time.Duration {
	return cmdDirection(state, mob, cmd, DirSouth)
}

func CmdWest(state *State, mob *Mob, cmd string) time.Duration {
	return cmdDirection(state, mob, cmd, DirWest)
}

func CmdUp(state *State, mob *Mob, cmd string) time.Duration {
	return cmdDirection(state, mob, cmd, DirUp)
}

func CmdDown(state *State, mob *Mob, cmd string) time.Duration {
	return cmdDirection(state, mob, cmd, DirDown)
}

func cmdDirection(state *State, mob *Mob, cmd string, dir int) time.Duration {
	if cmd != "" {
		mob.Send(MsgEnvironment, "I don't know what to do with the extra information.\n")
		return TimeToMove
	}

	// see if there is a door in that direction
	for _, door := range mob.Location.Doors {
		if door.Door == dir {
			id := door.ToRoom
			if id < 0 || id >= len(state.Rooms) || state.Rooms[id] == nil {
				mob.Send(MsgEnvironment, "Error trying to move in that direction\n")
				return TimeToMove
			}

			mob.Location = state.Rooms[door.ToRoom]
			var msg string
			if mob.Visited[mob.Location.ID] {
				msg = mob.Location.GetShortDescription()
			} else {
				msg = mob.Location.GetDescription()
			}
			mob.Visited[mob.Location.ID] = true
			mob.Send(MsgEnvironment, msg)
			mob.Send(MsgMap, GetMap(state, mob.Location, mob.Visited))
			return TimeToMove
		}
	}

	mob.Send(MsgEnvironment, "You cannot go that way.\n")
	return TimeToMove
}

func CmdRecall(state *State, mob *Mob, cmd string) time.Duration {
	if cmd != "" {
		mob.Send(MsgEnvironment, "I don't know what to do with the extra information.\n")
		return TimeToMove
	}

	mob.Location = state.Rooms[RecallLocation]
	mob.Send(MsgEnvironment, mob.Location.GetShortDescription())
	mob.Send(MsgMap, GetMap(state, mob.Location, mob.Visited))
	return TimeToMove
}

func (r *Room) Zone() int {
	return r.ID / 100
}

func (r *Room) Exit(state *State, dir rune) *Room {
	for _, door := range r.Doors {
		exit := rune(directions[door.Door][0])
		if exit == dir && door.ToRoom >= 0 && door.ToRoom < len(state.Rooms) {
			return state.Rooms[door.ToRoom]
		}
	}
	return nil
}

func (r *Room) GetShortDescription() string {
	var buf bytes.Buffer
	buf.WriteString(r.Name)
	buf.WriteString("\nExits [")
	for i, door := range r.Doors {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(directions[door.Door][0:1])
	}
	buf.WriteString("]\n")
	return buf.String()
}

func (r *Room) GetDescription() string {
	var buf bytes.Buffer
	buf.WriteString(r.Description)
	buf.WriteString("\nExits [")
	for i, door := range r.Doors {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(directions[door.Door][0:1])
	}
	buf.WriteString("]\n")
	return buf.String()
}
