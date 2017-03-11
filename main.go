package main

import (
	"flag"
	"github.com/qaisjp/studenthackv-go-gameserver/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var addr = flag.String("address", ":8080", "the address to ")

func main() {
	flag.Parse()

	// Create a new signal receiver
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	// Create a new API
	server := server.NewServer(&server.Options{
		Address: *addr,
	})

	log.Println("Server starting...")

	// Spawn it in a new goroutine
	go server.Run()

	// Watch for a signal
	<-sc

	log.Println("Server shutting down...")

	// Exit the API
	server.Exit()
}
