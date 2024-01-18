package main

import (
	server "GoChat/internal/server"
	"flag"
	"fmt"
	"log"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address to listen on")
	flag.Parse()
	fmt.Println(*addr)
	srv := server.NewServer(*addr)
	err := srv.Run()
	if err != nil {
		log.Println(err)
	}
}
