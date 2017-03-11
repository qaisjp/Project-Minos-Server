package game

import (
	"encoding/json"
)

type RawMessageIn struct {
	Type    string
	Payload json.RawMessage
}

type MessageIn struct {
	Player  *Player
	Type    string
	Payload json.RawMessage
}

type MessageOut struct {
	Type    string
	Payload interface{}
}
