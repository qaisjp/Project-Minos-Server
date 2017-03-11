package game

import (
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
	broadcast chan RawMessage

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

		broadcast:  make(chan RawMessage),
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
			log.Printf("Player(%s) sent: %s\n", message.player.ID, message.payload)
			g.onMessageReceive(message)
		}
	}
}

func (g *Game) onPlayerConnect(p *Player) {
	// g.onPlayerJoin(p)
	g.players[p] = true
	log.Printf("New player(%s) joined the game!\n", p.ID)

	// Send them the map
	p.SendMap()
}

func (g *Game) onMessageReceive(m RawMessage) {
	for player := range g.players {
		select {
		case player.send <- append([]byte("Message: "), m.payload[:]...):
		default:
			log.Println("Connection lost perhaps?")
			close(player.send)
			delete(g.players, player)
		}
	}
}
