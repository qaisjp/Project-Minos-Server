package game

type RawMessage struct {
	player  *Player
	payload []byte
}

type EncodedMessage struct {
	Type    string
	Payload interface{}
}
