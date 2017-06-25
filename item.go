package main

import "time"

type Item struct {
	ID                int
	Name              string
	ShortDescription  string
	LongDescription   string
	ActionDescription string
	Keywords          []string
	Weight            int
	Value             int
	Expires           time.Time
	Effects           []*Effect
}

type Pop struct {
}
