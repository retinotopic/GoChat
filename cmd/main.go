package main

import (
	server "GoChat/internal/server"
	"log"
)

func main() {
	srv := server.NewServer("localhost:8080")
	err := srv.Run()
	if err != nil {
		log.Println(err)
	}
}
