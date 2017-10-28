package main

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
	"github.com/russross/meddler"
)

type input struct {
	data []byte
	rest []byte
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <areafile1> ...", os.Args[0])
	}

	var areas []*Area
	filenames := make(map[*Area]string)
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
				area = &Area{Name: name}
				areas = append(areas, area)
				filenames[area] = basename

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
				shops := parseShops(in)
				log.Printf("found %d SHOPS", len(shops))

			case "SPECIALS":
				log.Printf("starting SPECIALS section")
				if area == nil {
					in.Failf("SPECIALS found outside an area")
				}
				specials := parseSpecials(in)
				log.Printf("found %d SPECIALS", len(specials))

			default:
				in.Failf("unimplemented section: %s", section)
			}
		}
	}

	db, err := sql.Open("sqlite3", "gruffles.db")
	if err != nil {
		log.Fatalf("opening db: %v", err)
	}
	defer db.Close()
	now := time.Now()

	for _, elt := range areas {
		elt.CreatedAt = now
		elt.ModifiedAt = now
		writeAreaSQL(db, elt, filenames[elt])
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
			Keywords: parseKeywords(keywords),
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
		actionFlags := in.parseNumber()
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
		sex := []string{"it", "he", "she"}[in.parseNumber()]

		mob := &Mobile{
			ID:               id,
			Keywords:         parseKeywords(keywords),
			ShortDescription: short,
			LongDescription:  long,
			Description:      desc,
			ActionFlags:      actionFlags,
			AffectedFlags:    affFlags,
			Alignment:        alignment,
			Level:            level,
			HitRoll:          makeRoll(hitNumberDice, hitSizeDice, hitPlus),
			DamageRoll:       makeRoll(damNumberDice, damSizeDice, damPlus),
			DodgeRoll:        makeRoll(0, 0, 0),
			AbsorbRoll:       makeRoll(0, 0, 0),
			FireRoll:         makeRoll(0, 0, 0),
			IceRoll:          makeRoll(0, 0, 0),
			PoisonRoll:       makeRoll(0, 0, 0),
			LightningRoll:    makeRoll(0, 0, 0),
			Gold:             gold,
			Experience:       exp,
			Pronouns:         sex,
		}
		mobiles = append(mobiles, mob)
		_, _, _, _ = hitroll, armor, position1, position2
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
		keywords := parseKeywords(in.parseString())
		shortDescription := in.parseString()
		longDescription := in.parseString()
		// ActionDescription
		_ = in.parseString()
		itemType := in.parseNumber()
		extraFlags := in.parseNumber()
		wearFlags := in.parseNumber()
		value0 := in.parseNumber()
		value1 := in.parseNumber()
		value2 := in.parseNumber()
		value3 := in.parseNumber()
		weight := in.parseNumber()
		cost := in.parseNumber()
		// CostPerDat
		_ = in.parseNumber()
		obj := &Object{
			ID:               id,
			Keywords:         keywords,
			ShortDescription: shortDescription,
			LongDescription:  longDescription,
			ItemType:         itemType,
			ExtraFlags:       extraFlags,
			WearFlags:        wearFlags,
			Value0:           value0,
			Value1:           value1,
			Value2:           value2,
			Value3:           value3,
			Weight:           weight,
			Cost:             cost,
			Extras:           []ObjectExtraDescription{},
			Applies:          []ObjectApply{},
		}
		for {
			if in.hasLetter("E") {
				in.expectLetter("E")
				extra := ObjectExtraDescription{
					Keywords:    parseKeywords(in.parseString()),
					Description: in.parseString(),
				}
				obj.Extras = append(obj.Extras, extra)
			} else if in.hasLetter("A") {
				in.expectLetter("A")
				apply := ObjectApply{
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
			AreaID:      in.parseNumber(),
			Flags:       in.parseNumber(),
			Terrain:     in.parseNumber(),
			Doors:       []RoomDoor{},
			Extras:      []RoomExtraDescription{},
		}
	optionals:
		for {
			kind := in.parseLetter()
			switch kind {
			case "D":
				door := RoomDoor{
					Direction:   in.parseNumber(),
					Description: in.parseString(),
					Keywords:    parseKeywords(in.parseString()),
					Lock:        in.parseNumber(),
					Key:         in.parseNumber(),
					ToRoom:      in.parseNumber(),
				}
				room.Doors = append(room.Doors, door)
			case "E":
				extra := RoomExtraDescription{
					Keywords:    parseKeywords(in.parseString()),
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

type Shop struct {
	Keeper                                 int
	Trade0, Trade1, Trade2, Trade3, Trade4 int
	ProfitBuy, ProfitSell                  int
	OpenHour, CloseHour                    int
	Comment                                string
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

type Special struct {
	MobID    int
	Function string
	Comment  string
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

func parseKeywords(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	if len(s) > 1 && strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		s = strings.TrimSpace(s[1 : len(s)-2])
	}
	out := []string{}
	for _, elt := range strings.Fields(s) {
		out = append(out, strings.ToLower(elt))
	}
	return out
}

func makeRoll(dice, faces, plus int) []int {
	if dice < 0 || faces < 0 || plus < 0 {
		log.Fatalf("makeRoll: invalid inputs: dice=%d, faces=%d, plus=%d",
			dice, faces, plus)
	}
	mean := float64(dice)*float64(faces+1)/2.0 + float64(plus)
	stddev := math.Sqrt(float64(dice*(faces*faces-1)) / 12.0)
	meanInt := int(100.0*mean + 0.5)
	stddevInt := int(100.0*stddev + 0.5)

	return []int{meanInt, stddevInt}
}

func writeAreaSQL(db *sql.DB, area *Area, filename string) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("starting transaction: %v", err)
	}
	defer tx.Commit()

	// save the area
	log.Printf("saving area %s", area.Name)
	if len(area.Rooms) > 0 {
		area.ID = area.Rooms[0].ID / 100
	} else {
		log.Fatalf("area %d (file %q) has no rooms", area.Name, filename)
	}
	if err = meddler.Insert(tx, "areas", area); err != nil {
		log.Fatalf("insert area: %v", err)
	}

	// save the helps
	if len(area.Helps) > 0 {
		log.Printf("saving %d helps", len(area.Helps))
	}
	for _, help := range area.Helps {
		help.AreaID = area.ID
		if err = meddler.Insert(tx, "helps", help); err != nil {
			log.Fatalf("insert help: %v", err)
		}
	}

	// save the mobs
	if len(area.Mobiles) > 0 {
		log.Printf("saving %d mobiles", len(area.Mobiles))
	}
	for _, mob := range area.Mobiles {
		mob.AreaID = area.ID
		if err = meddler.Insert(tx, "mobiles", mob); err != nil {
			log.Fatalf("insert mobile: %v", err)
		}
	}

	// save the objects
	if len(area.Objects) > 0 {
		log.Printf("saving %d objects", len(area.Objects))
	}
}
