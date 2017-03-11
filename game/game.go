package game

import (
	"encoding/json"
	"github.com/qaisjp/studenthackv-go-gameserver/mapgen"
	"log"
)

type Game struct {
	players map[*Player]bool

	Monster  *Player
	King     *Player
	Servants []*Player
	Map      *mapgen.Map

	// Inbound messages from players.
	broadcast chan MessageIn

	// Register requests from players.
	register chan *Player

	// Unregister requests from players.
	unregister chan *Player

	alive bool
}

func NewGame() *Game {
	log.Println("New game created")

	g := &Game{
		alive: true,

		broadcast:  make(chan MessageIn),
		register:   make(chan *Player),
		unregister: make(chan *Player),
		players:    make(map[*Player]bool),
		Map:        mapgen.NewMap(100, 100),
	}
	return g
}

func (g *Game) IsAlive() bool {
	return g.alive
}

func (g *Game) Run() {
	for {
		select {
		case player := <-g.register:
			g.onPlayerConnect(player)
		case player := <-g.unregister:
			if _, ok := g.players[player]; ok {
				log.Printf("Player(%s) left the server", player.ID)
				delete(g.players, player)
				close(player.send)
			}
		case message := <-g.broadcast:
			log.Printf("Received (%s) from (%s): %s\n", message.Type, message.Player.ID, message.Payload)
			g.onMessageReceive(message)
		}
	}
}

func (g *Game) onPlayerConnect(p *Player) {
	// g.onPlayerJoin(p)
	g.players[p] = true

	// Send them the map
	p.SendMap()

	log.Printf("New player(%s) joined the game!\n", p.ID)
}

func (g *Game) onMessageReceive(m MessageIn) {

	switch m.Type {
	case "ident":

		var payload string
		json.Unmarshal(m.Payload, &payload)

		m.Player.onIdentify(payload)
	case "pos":
		json.Unmarshal(m.Payload, &m.Player.Position)

		for p := range g.players {
			if p.ID != m.Player.ID {
				p.Send(MessageOut{
					Type:    "player",
					Payload: m.Player,
				})
			}
		}
	default:
		var payload string
		json.Unmarshal(m.Payload, &payload)
		log.Printf("Payload: %s\n", payload)

		for player := range g.players {
			select {
			case player.send <- append([]byte("Message: "), payload[:]...):
			default:
				log.Println("Connection lost perhaps?")
				close(player.send)
				delete(g.players, player)
			}
		}
	}
}
