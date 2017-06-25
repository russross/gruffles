package main

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// A Command is a function:
// func (state *State, m *Mob, args string) time.Duration
// args is what the user typed with the command removed from the beginning
// the command should return the duration of the delay before another
// command can execute. zero means minimum delay.

type Command struct {
	Command string
	Execute func(state *State, mob *Mob, cmd string) time.Duration
	Fast    bool
}

var Commands map[string]*Command

func SetupCommands() {
	Commands = make(map[string]*Command)
	addCommand(&Command{Command: "look", Execute: CmdLook, Fast: true}, []string{"l"})
	addCommand(&Command{Command: "north", Execute: CmdNorth, Fast: false}, []string{"n"})
	addCommand(&Command{Command: "east", Execute: CmdEast, Fast: false}, []string{"e"})
	addCommand(&Command{Command: "south", Execute: CmdSouth, Fast: false}, []string{"s"})
	addCommand(&Command{Command: "west", Execute: CmdWest, Fast: false}, []string{"w"})
	addCommand(&Command{Command: "up", Execute: CmdUp, Fast: false}, []string{"u"})
	addCommand(&Command{Command: "down", Execute: CmdDown, Fast: false}, []string{"d"})
	addCommand(&Command{Command: "recall", Execute: CmdRecall, Fast: false}, nil)
}

func ParseCommand(input string) (*Command, string) {
	if !utf8.ValidString(input) {
		return nil, ""
	}

	// parse the command word from the rest of the string
	word, rest := strings.TrimSpace(input), ""
	if space := strings.IndexFunc(input, unicode.IsSpace); space >= 0 {
		rest = strings.TrimSpace(input[space:])
		word = input[:space]
	}
	if len(word) == 0 {
		return nil, ""
	}

	cmd, exists := Commands[word]
	if !exists {
		return nil, ""
	}

	return cmd, rest
}

func addCommand(cmd *Command, aliases []string) {
	Commands[cmd.Command] = cmd
	for _, alias := range aliases {
		Commands[alias] = cmd
	}
}
