package server

import (
	"github.com/gorilla/websocket"
	"github.com/qaisjp/studenthackv-go-gameserver/game"
	"log"
	"net/http"
)

type Server struct {
	Options *Options
	Games   []*game.Game

	newGameChan (chan *game.Game)
}

func NewServer(options *Options) *Server {

	return &Server{
		Options: options,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) servePlayer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	game.NewPlayer(s.Games[0], conn)
}

func (s *Server) Run() {
	// Prepare the server game list
	s.Games = []*game.Game{}
	s.newGameChan = make(chan *game.Game)

	// Start processing games
	go processGames(s)

	// Test game
	s.newGameChan <- game.NewGame()

	http.HandleFunc("/game/0/ws", s.servePlayer)

	log.Println("Server hosted at " + s.Options.Address)

	err := http.ListenAndServe(s.Options.Address, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *Server) Exit() {

}

func processGames(s *Server) {
	for {

		// Add to the games list and process ticks
		g := <-s.newGameChan
		s.Games = append(s.Games, g)

		go g.Run()
	}
}
