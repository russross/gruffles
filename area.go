package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"time"
)

type Area struct {
	ID         int        `meddler:"id,pk"`
	Name       string     `meddler:"name"`
	Helps      []*Help    `meddler:""`
	Mobiles    []*Mobile  `meddler:""`
	Objects    []*Object  `meddler:""`
	Rooms      []*Room    `meddler:""`
	Resets     []*Reset   `meddler:""`
	Shops      []*Shop    `meddler:""`
	Specials   []*Special `meddler:""`
	CreatedAt  time.Time  `meddler:"created_at"`
	ModifiedAt time.Time  `meddler:"modified_at"`
}

type Help struct {
	ID       int      `meddler:"id,pk"`
	AreaID   int      `meddler:"area_id"`
	Level    int      `meddler:"level"`
	Keywords []string `meddler:"keywords,json"`
	Text     string   `meddler:"text"`
}

type Mobile struct {
	ID                                  int
	Keywords                            string
	ShortDescription                    string
	LongDescription                     string
	Description                         string
	ActFlags, AffectedFlags, Alignment  int
	Level, Hitroll, Armor               int
	HitNumberDice, HitSizeDice, HitPlus int
	DamNumberDice, DamSizeDice, DamPlus int
	Gold, Experience                    int
	Position1, Position2, Sex           int
}

type Extra struct {
	Keywords    string
	Description string
}

type Apply struct {
	Type  int
	Value int
}

type Object struct {
	ID                              int
	Keywords                        string
	ShortDescription                string
	LongDescription                 string
	ActionDescription               string
	ItemType, ExtraFlags, WearFlags int
	Value0, Value1, Value2, Value3  int
	Weight, Cost, CostPerDay        int
	Extras                          []Extra
	Applies                         []Apply
}

type Door struct {
	Door               int
	Description        string
	Keywords           string
	Locks, Key, ToRoom int
}

type Room struct {
	ID                          int
	Name                        string
	Description                 string
	Area, RoomFlags, SectorType int
	Doors                       []Door
	Extras                      []Extra
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

type Reset struct {
	Type    string
	IfFlag  int
	Num1    int
	Num2    int
	Num3    int
	Comment string
}

type Shop struct {
	Keeper                                 int
	Trade0, Trade1, Trade2, Trade3, Trade4 int
	ProfitBuy, ProfitSell                  int
	OpenHour, CloseHour                    int
	Comment                                string
}

type Special struct {
	MobID    int
	Function string
	Comment  string
}

func LoadAreas(paths []string) ([]*Area, []*Room) {
	var areas []*Area
	var rooms []*Room

	// load the raw area files
	for _, path := range paths {
		fp, err := os.Open(path)
		if err != nil {
			log.Fatalf("opening %s: %v", path, err)
		}
		j := json.NewDecoder(fp)
		area := new(Area)
		if err = j.Decode(area); err != nil {
			log.Fatalf("decoding %s: %v", path, err)
		}
		fp.Close()
		areas = append(areas, area)
	}
	log.Printf("loaded %d areas", len(areas))

	// create a sparse slice of rooms mapping ID -> Room
	max := 0
	for _, area := range areas {
		for _, room := range area.Rooms {
			if room.ID > max {
				max = room.ID
			}
		}
	}
	rooms = make([]*Room, max+1)
	for _, area := range areas {
		for _, room := range area.Rooms {
			if rooms[room.ID] != nil {
				log.Fatalf("duplicate room id %d found in area %q", room.ID, area.Name)
			}
			rooms[room.ID] = room
		}
	}
	return areas, rooms
}
