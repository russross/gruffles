package main

import (
	"time"
)

type Area struct {
	ID         int       `meddler:"id,pk"`
	Name       string    `meddler:"name"`
	Helps      []*Help   `meddler:"-"`
	Mobiles    []*Mobile `meddler:"-"`
	Objects    []*Object `meddler:"-"`
	Rooms      []*Room   `meddler:"-"`
	Resets     []*Reset  `meddler:"-"`
	CreatedAt  time.Time `meddler:"created_at"`
	ModifiedAt time.Time `meddler:"modified_at"`
}

type Help struct {
	ID       int      `meddler:"id,pk"`
	AreaID   int      `meddler:"area_id"`
	Level    int      `meddler:"level"`
	Keywords []string `meddler:"keywords,json"`
	Text     string   `meddler:"help_text"`
}

type Mobile struct {
	ID               int      `meddler:"id,pk"`
	AreaID           int      `meddler:"area_id"`
	Keywords         []string `meddler:"keywords,json"`
	ShortDescription string   `meddler:"short_description"`
	LongDescription  string   `meddler:"long_description"`
	Description      string   `meddler:"description"`
	ActionFlags      int      `meddler:"action_flags"`
	AffectedFlags    int      `meddler:"affected_flags"`
	Alignment        int      `meddler:"alignment"`
	Level            int      `meddler:"level"`
	HitRoll          []int    `meddler:"hit_roll,json"`
	DamageRoll       []int    `meddler:"damage_roll,json"`
	DodgeRoll        []int    `meddler:"dodge_roll,json"`
	AbsorbRoll       []int    `meddler:"absorb_roll,json"`
	FireRoll         []int    `meddler:"fire_roll,json"`
	IceRoll          []int    `meddler:"ice_roll,json"`
	PoisonRoll       []int    `meddler:"poison_roll,json"`
	LightningRoll    []int    `meddler:"lightning_roll,json"`
	Gold             int      `meddler:"gold"`
	Experience       int      `meddler:"experience"`
	Pronouns         string   `meddler:"pronouns"`
}

type Object struct {
	ID               int                      `meddler:"id"`
	Keywords         []string                 `meddler:"keywords,json"`
	ShortDescription string                   `meddler:"short_description"`
	LongDescription  string                   `meddler:"long_description"`
	ItemType         int                      `meddler:"item_type"`
	ExtraFlags       int                      `meddler:"extra_flags"`
	WearFlags        int                      `meddler:"wear_flags"`
	Value0           int                      `meddler:"value_0"`
	Value1           int                      `meddler:"value_1"`
	Value2           int                      `meddler:"value_2"`
	Value3           int                      `meddler:"value_3"`
	Weight           int                      `meddler:"weight"`
	Cost             int                      `meddler:"cost"`
	Extras           []ObjectExtraDescription `meddler:"extra,json"`
	Applies          []ObjectApply            `meddler:"applies,json"`
}

type ObjectExtraDescription struct {
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
}

type ObjectApply struct {
	Type  int `json:"type"`
	Value int `json:"value"`
}

type Room struct {
	ID          int                    `meddler:"id,pk"`
	AreaID      int                    `meddler:"area_id"`
	Name        string                 `meddler:"name"`
	Description string                 `meddler:"description"`
	Flags       int                    `meddler:"flags"`
	Terrain     int                    `meddler:"terrain"`
	Doors       []RoomDoor             `meddler:"doors,json"`
	Extras      []RoomExtraDescription `meddler:"extras,json"`
}

type RoomDoor struct {
	Direction   int      `json:"direction"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Lock        int      `json:"lock"`
	Key         int      `json:"key"`
	ToRoom      int      `json:"toRoom"`
}

type RoomExtraDescription struct {
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
}

type Reset struct {
	ID            int    `meddler:"id,pk"`
	Type          string `meddler:"reset_type"`
	AreaID        int    `meddler:"area_id"`
	RoomID        int    `meddler:"room_id,zeroisnull"`
	MobileID      int    `meddler:"mobile_id,zeroisnull"`
	ObjectID      int    `meddler:"object_id,zeroisnull"`
	ContainerID   int    `meddler:"container_id,zeroisnull"`
	MaxInstances  int    `meddler:"max_instances,zeroisnull"`
	DoorDirection int    `meddler:"door_direction"`
	DoorState     int    `meddler:"door_state"`
}
