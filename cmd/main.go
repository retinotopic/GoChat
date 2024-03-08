package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/retinotopic/GoChat/internal/router"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address to listen on")

	flag.Parse()
	fmt.Println(*addr)
	srv := router.NewRouter(*addr)
	err := srv.Run()
	if err != nil {
		log.Println(err)
	}
}
