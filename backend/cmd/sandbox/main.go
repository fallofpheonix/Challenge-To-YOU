package main

import (
	"log"

	"challenge-to-you/backend/internal/server"
)

func main() {
	log.Println("Starting Challenge-To-YOU server...")
	server.NewServer().Start()
}
