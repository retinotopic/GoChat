package main

import (
	"flag"
	"fmt"
	"log"

	server "github.com/retinotopic/GoChat/internal/routes"
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
