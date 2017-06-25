package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Area struct {
	Name     string
	Filename string `json:"-"`
	Helps    []*Help
	Mobiles  []*Mobile
	Objects  []*Object
	Rooms    []*Room
	Resets   []*Reset
	Shops    []*Shop
	Specials []*Special
}

type Help struct {
	Level    int
	Keywords string
	Text     string
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

type input struct {
	data []byte
	rest []byte
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <areafile1> ...", os.Args[0])
	}

	var areas []*Area
	for _, filename := range os.Args[1:] {
		basename := filename
		if strings.HasSuffix(basename, ".are") {
			basename = basename[:len(basename)-len(".are")]
		}

		raw, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		if !utf8.Valid(raw) {
			log.Fatalf("file is not valid utf8")
		}
		in := &input{
			data: raw,
			rest: raw,
		}

		log.Printf("parsing file %s", filename)
		var area *Area

		for len(in.rest) > 0 {
			section := in.parseHeader()

			switch section {
			case "$":
				log.Printf("end of file marker found, quitting")
				in.rest = nil

			case "AREA":
				name := in.parseString()
				log.Printf("starting AREA %s", name)
				area = &Area{Name: name, Filename: basename}
				areas = append(areas, area)

			case "HELPS":
				log.Printf("starting HELPS section")
				if area == nil {
					in.Failf("HELPS found outside an area")
				}
				area.Helps = parseHelps(in)
				log.Printf("found %d HELPS", len(area.Helps))

			case "MOBILES":
				log.Printf("starting MOBILES section")
				if area == nil {
					in.Failf("MOBILES found outside an area")
				}
				area.Mobiles = parseMobiles(in)
				log.Printf("found %d MOBILES", len(area.Mobiles))

			case "OBJECTS":
				log.Printf("starting OBJECTS section")
				if area == nil {
					in.Failf("OBJECTS found outside an area")
				}
				area.Objects = parseObjects(in)
				log.Printf("found %d OBJECTS", len(area.Objects))

			case "ROOMS":
				log.Printf("starting ROOMS section")
				if area == nil {
					in.Failf("ROOMS found outside an area")
				}
				area.Rooms = parseRooms(in)
				log.Printf("found %d ROOMS", len(area.Rooms))

			case "RESETS":
				log.Printf("starting RESETS section")
				if area == nil {
					in.Failf("RESETS found outside an area")
				}
				area.Resets = parseResets(in)
				log.Printf("found %d RESETS", len(area.Resets))

			case "SHOPS":
				log.Printf("starting SHOPS section")
				if area == nil {
					in.Failf("SHOPS found outside an area")
				}
				area.Shops = parseShops(in)
				log.Printf("found %d SHOPS", len(area.Shops))

			case "SPECIALS":
				log.Printf("starting SPECIALS section")
				if area == nil {
					in.Failf("SPECIALS found outside an area")
				}
				area.Specials = parseSpecials(in)
				log.Printf("found %d SPECIALS", len(area.Specials))

			default:
				in.Failf("unimplemented section: %s", section)
			}
		}
	}

	count := make(map[string]int)
	for _, elt := range areas {
		filename := elt.Filename
		count[filename]++
		if n := count[filename]; n > 1 {
			filename += "." + strconv.Itoa(n)
		}
		filename += ".json"
		raw, err := json.MarshalIndent(elt, "", "    ")
		if err != nil {
			log.Fatalf("json error: %v", err)
		}
		if err = ioutil.WriteFile(filename, raw, 0644); err != nil {
			log.Fatalf("error writing %s: %v", filename, err)
		}
	}
}

func (in *input) parseHeader() string {
	in.rest = bytes.TrimLeftFunc(in.rest, unicode.IsSpace)
	in.expectLetter("#")
	word := in.parseWord()
	return word
}

func (in *input) parseLetter() string {
	in.rest = bytes.TrimLeftFunc(in.rest, unicode.IsSpace)
	if len(in.rest) == 0 {
		in.Failf("unexpected end-of-file while parsing a letter")
	}
	r, size := utf8.DecodeRune(in.rest)
	in.rest = in.rest[size:]
	return string(r)
}

func (in *input) hasLetter(ch string) bool {
	rest := bytes.TrimLeftFunc(in.rest, unicode.IsSpace)
	if len(rest) == 0 {
		in.Failf("unexpected end-of-file while peeking at letter")
	}
	r, _ := utf8.DecodeRune(rest)
	return string(r) == ch
}

func (in *input) expectLetter(expected string) {
	found := in.parseLetter()
	if found != expected {
		in.Failf("expected %q but found %q", expected, found)
	}
}

func (in *input) parseWord() string {
	in.rest = bytes.TrimLeftFunc(in.rest, unicode.IsSpace)
	start := in.rest
	length := 0
	for {
		if len(in.rest) == 0 {
			break
		}
		r, size := utf8.DecodeRune(in.rest)
		if unicode.IsSpace(r) {
			break
		}
		length += size
		in.rest = in.rest[size:]
	}
	if length == 0 {
		in.Failf("found EOF while parsing a word")
	}
	return string(start[:length])
}

func (in *input) parseString() string {
	in.rest = bytes.TrimLeftFunc(in.rest, unicode.IsSpace)
	start := in.rest
	length := 0
	for {
		if len(in.rest) == 0 {
			in.Failf("unexpected end-of-file while parsing a string")
		}
		r, size := utf8.DecodeRune(in.rest)
		in.rest = in.rest[size:]
		if r == '~' {
			break
		}
		length += size
	}
	return string(start[:length])
}

func (in *input) parseNumber() int {
	in.rest = bytes.TrimLeftFunc(in.rest, unicode.IsSpace)
	start := in.rest
	length := 0
	for {
		if len(in.rest) == 0 {
			in.Failf("unexpected end-of-file while parsing a number")
		}
		r, size := utf8.DecodeRune(in.rest)
		if !unicode.IsDigit(r) && r != '|' && r != '+' && r != '-' {
			break
		}
		if length > 0 && (r == '+' || r == '-') {
			break
		}
		in.rest = in.rest[size:]
		length += size
	}
	if length == 0 {
		in.Failf("found zero-length while parsing a number")
	}
	if length == 1 && (start[0] == '+' || start[1] == '-') {
		in.Failf("found only a + or - while parsing a number")
	}

	s := string(start[:length])
	n, err := interpretNumber(s)
	if err != nil {
		in.Failf("error interpretting number %q: %v", s, err)
	}
	return n
}

func interpretNumber(s string) (int, error) {
	n := 0
	for _, elt := range strings.Split(s, "|") {
		if elt == "" {
			continue
		}
		val, err := strconv.Atoi(elt)
		if err != nil {
			return 0, err
		}
		n += val
	}
	return n, nil
}

func (in *input) parseToEOL() string {
	in.rest = bytes.TrimLeft(in.rest, " \t")
	start := in.rest
	length := 0
	for {
		if len(in.rest) == 0 {
			in.Failf("unexpected end-of-file while parsing to EOL")
		}
		r, size := utf8.DecodeRune(in.rest)
		in.rest = in.rest[size:]
		if r == '\n' {
			break
		}
		length += size
	}
	return string(start[:length])
}

func (in *input) Failf(fmt string, params ...interface{}) {
	parsed := in.data[:len(in.data)-len(in.rest)]
	newlines := bytes.Count(parsed, []byte{'\n'})
	lastNewline := bytes.LastIndex(parsed, []byte{'\n'})
	if lastNewline == -1 {
		lastNewline = 0
	}
	lastLine := parsed[lastNewline:]
	log.Printf("at line %d column %d", newlines+1, len(lastLine))
	log.Fatalf(fmt, params...)
}

func parseHelps(in *input) []*Help {
	helps := []*Help{}
	for {
		level := in.parseNumber()
		keywords := in.parseString()
		if level == 0 && keywords == "$" {
			break
		}
		text := in.parseString()
		help := &Help{
			Level:    level,
			Keywords: keywords,
			Text:     text,
		}
		helps = append(helps, help)
	}
	return helps
}

func parseMobiles(in *input) []*Mobile {
	mobiles := []*Mobile{}
	for {
		in.expectLetter("#")
		id := in.parseNumber()
		if id == 0 {
			break
		}
		keywords := in.parseString()
		short := in.parseString()
		long := in.parseString()
		desc := in.parseString()
		actFlags := in.parseNumber()
		affFlags := in.parseNumber()
		alignment := in.parseNumber()
		in.expectLetter("S")
		level := in.parseNumber()
		hitroll := in.parseNumber()
		armor := in.parseNumber()
		hitNumberDice := in.parseNumber()
		if in.hasLetter("D") {
			in.expectLetter("D")
		} else {
			in.expectLetter("d")
		}
		hitSizeDice := in.parseNumber()
		in.expectLetter("+")
		hitPlus := in.parseNumber()
		damNumberDice := in.parseNumber()
		if in.hasLetter("D") {
			in.expectLetter("D")
		} else {
			in.expectLetter("d")
		}
		damSizeDice := in.parseNumber()
		in.expectLetter("+")
		damPlus := in.parseNumber()
		gold := in.parseNumber()
		exp := in.parseNumber()
		position1 := in.parseNumber()
		position2 := in.parseNumber()
		sex := in.parseNumber()

		mob := &Mobile{
			ID:               id,
			Keywords:         keywords,
			ShortDescription: short,
			LongDescription:  long,
			Description:      desc,
			ActFlags:         actFlags,
			AffectedFlags:    affFlags,
			Alignment:        alignment,
			Level:            level,
			Hitroll:          hitroll,
			Armor:            armor,
			HitNumberDice:    hitNumberDice,
			HitSizeDice:      hitSizeDice,
			HitPlus:          hitPlus,
			DamNumberDice:    damNumberDice,
			DamSizeDice:      damSizeDice,
			DamPlus:          damPlus,
			Gold:             gold,
			Experience:       exp,
			Position1:        position1,
			Position2:        position2,
			Sex:              sex,
		}
		mobiles = append(mobiles, mob)
	}
	return mobiles
}

func parseObjects(in *input) []*Object {
	objects := []*Object{}
	for {
		in.expectLetter("#")
		id := in.parseNumber()
		if id == 0 {
			break
		}
		obj := &Object{
			ID:                id,
			Keywords:          in.parseString(),
			ShortDescription:  in.parseString(),
			LongDescription:   in.parseString(),
			ActionDescription: in.parseString(),
			ItemType:          in.parseNumber(),
			ExtraFlags:        in.parseNumber(),
			WearFlags:         in.parseNumber(),
			Value0:            in.parseNumber(),
			Value1:            in.parseNumber(),
			Value2:            in.parseNumber(),
			Value3:            in.parseNumber(),
			Weight:            in.parseNumber(),
			Cost:              in.parseNumber(),
			CostPerDay:        in.parseNumber(),
			Extras:            []Extra{},
			Applies:           []Apply{},
		}
		for {
			if in.hasLetter("E") {
				in.expectLetter("E")
				extra := Extra{
					Keywords:    in.parseString(),
					Description: in.parseString(),
				}
				obj.Extras = append(obj.Extras, extra)
			} else if in.hasLetter("A") {
				in.expectLetter("A")
				apply := Apply{
					Type:  in.parseNumber(),
					Value: in.parseNumber(),
				}
				obj.Applies = append(obj.Applies, apply)
			} else {
				break
			}
		}

		objects = append(objects, obj)
	}
	return objects
}

func parseRooms(in *input) []*Room {
	rooms := []*Room{}
	for {
		in.expectLetter("#")
		id := in.parseNumber()
		if id == 0 {
			break
		}
		room := &Room{
			ID:          id,
			Name:        in.parseString(),
			Description: in.parseString(),
			Area:        in.parseNumber(),
			RoomFlags:   in.parseNumber(),
			SectorType:  in.parseNumber(),
			Doors:       []Door{},
			Extras:      []Extra{},
		}
	optionals:
		for {
			kind := in.parseLetter()
			switch kind {
			case "D":
				door := Door{
					Door:        in.parseNumber(),
					Description: in.parseString(),
					Keywords:    in.parseString(),
					Locks:       in.parseNumber(),
					Key:         in.parseNumber(),
					ToRoom:      in.parseNumber(),
				}
				room.Doors = append(room.Doors, door)
			case "E":
				extra := Extra{
					Keywords:    in.parseString(),
					Description: in.parseString(),
				}
				room.Extras = append(room.Extras, extra)
			case "S":
				break optionals
			default:
				in.Failf("unknown optional field %q in room", kind)
			}
		}

		rooms = append(rooms, room)
	}
	return rooms
}

func parseResets(in *input) []*Reset {
	resets := []*Reset{}
loop:
	for {
		kind := in.parseLetter()
		reset := &Reset{}
		switch kind {
		case "*":
			in.parseToEOL()
		case "M", "O", "P", "E", "D":
			reset.Type = kind
			reset.IfFlag = in.parseNumber()
			reset.Num1 = in.parseNumber()
			reset.Num2 = in.parseNumber()
			reset.Num3 = in.parseNumber()
			reset.Comment = in.parseToEOL()
		case "G", "R":
			reset.Type = kind
			reset.IfFlag = in.parseNumber()
			reset.Num1 = in.parseNumber()
			reset.Num2 = in.parseNumber()
			reset.Comment = in.parseToEOL()
		case "S":
			break loop
		default:
			in.Failf("unexpected RESET type: %q", kind)
		}
		resets = append(resets, reset)
	}
	return resets
}

func parseShops(in *input) []*Shop {
	shops := []*Shop{}
	for {
		keeper := in.parseNumber()
		if keeper == 0 {
			break
		}
		shop := &Shop{
			Keeper:     keeper,
			Trade0:     in.parseNumber(),
			Trade1:     in.parseNumber(),
			Trade2:     in.parseNumber(),
			Trade3:     in.parseNumber(),
			Trade4:     in.parseNumber(),
			ProfitBuy:  in.parseNumber(),
			ProfitSell: in.parseNumber(),
			OpenHour:   in.parseNumber(),
			CloseHour:  in.parseNumber(),
			Comment:    in.parseToEOL(),
		}
		shops = append(shops, shop)
	}
	return shops
}

func parseSpecials(in *input) []*Special {
	specials := []*Special{}
loop:
	for {
		kind := in.parseLetter()
		switch kind {
		case "*":
			in.parseToEOL()
		case "M":
			special := &Special{
				MobID:    in.parseNumber(),
				Function: in.parseWord(),
				Comment:  in.parseToEOL(),
			}
			specials = append(specials, special)
		case "S":
			break loop
		default:
			in.Failf("unknown special type %q", kind)
		}
	}
	return specials
}
