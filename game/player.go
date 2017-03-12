package game

import (
	"bytes"
	"encoding/json"
	"github.com/dchest/uniuri"
	"github.com/gorilla/websocket"
	. "github.com/qaisjp/studenthackv-go-gameserver/structs"
	"log"
	"time"
)

type CharacterID int

const (
	UnassignedCharacter CharacterID = iota
	MonsterCharacter
	KingCharacter
	ServantCharacter
)

type PlayerID string

type Player struct {
	ID        PlayerID
	game      *Game           // The game they belong to
	conn      *websocket.Conn // The websocket connection
	send      chan []byte     // Buffer channel of outbound messages
	Character CharacterID
	Position  Position
	Dead      bool
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 500 * time.Millisecond

	// Time allowed to read the next pong message from the peer.
	pongWait = 1 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func NewPlayer(g *Game, conn *websocket.Conn) *Player {
	client := &Player{
		ID:   PlayerID(string(uniuri.NewLen(uniuri.UUIDLen))),
		game: g,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.Position.X, client.Position.Z = g.Map.GetRandomSpace()

	log.Printf("New player(%s) connected...\n", client.ID)
	g.register <- client

	go client.writePump()
	client.readPump()

	return client
}

// readPump pumps messages from the websocket connection to the game.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (p *Player) readPump() {
	defer func() {
		p.game.unregister <- p
		p.conn.Close()
		log.Println("ReadPump ended")
	}()
	// p.conn.SetReadLimit(maxMessageSize)
	// p.conn.SetReadDeadline(time.Now().Add(pongWait))
	// p.conn.SetPongHandler(func(string) error { p.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			// if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			log.Printf("error: %v", err)
			// }
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		m := MessageIn{}
		err = json.Unmarshal(message, &m)
		if err != nil {
			log.Println(err)
			log.Printf("Invalid JSON received from %s: %s\n", p.ID, message)
			continue
		}

		m.Player = p

		p.game.broadcast <- m
	}
}

// writePump pumps messages from the game to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (p *Player) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()
	for {
		select {
		case message, ok := <-p.send:
			// p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(p.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-p.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (p *Player) Send(m MessageOut) {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	p.send <- b
}

func (p *Player) SendMap() {
	p.Send(MessageOut{
		Type:    "map",
		Payload: p.game.Map,
	})
}

func (p *Player) onIdentify(name string) {
	if string(name) == "oculus" {
		p.game.Monster = p
		p.Character = MonsterCharacter
	} else if string(name) == "player" {
		if p.game.King == nil {
			p.Character = KingCharacter
			p.game.King = p
		} else {
			p.Character = ServantCharacter
			p.game.Servants = append(p.game.Servants, p)
		}
	} else {
		panic("Invalid name " + name)
	}

	players := []*Player{}
	for p := range p.game.players {
		players = append(players, p)
	}

	p.Send(MessageOut{
		Type: "registered",
		Payload: struct {
			You     *Player
			Players []*Player
		}{
			You:     p,
			Players: players,
		},
	})

	for player := range p.game.players {
		if p.ID != player.ID {
			player.Send(MessageOut{
				Type:    "joined",
				Payload: p,
			})
		}
	}
}

func (p *Player) onLeave() {
	log.Printf("Player(%s) left the server", p.ID)
	for player := range p.game.players {
		if player.ID != p.ID {
			player.Send(MessageOut{
				Type:    "quit",
				Payload: p.ID,
			})
		}
	}
}
