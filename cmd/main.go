package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/retinotopic/GoChat/internal/auth"
	"github.com/retinotopic/GoChat/internal/providers/gfirebase"
	"github.com/retinotopic/GoChat/internal/providers/google"
	server "github.com/retinotopic/GoChat/internal/routes"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address to listen on")
	auth.CurrentProviders = auth.Providers{
		"google":    google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), "http://localhost:8080/google/CompleteAuth"),
		"gfirebase": gfirebase.New(os.Getenv("WEB_API_KEY"), os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "http://localhost:8080/gfirebase/CompleteAuth"),
	}
	flag.Parse()
	fmt.Println(*addr)
	srv := server.NewServer(*addr)
	err := srv.Run()
	if err != nil {
		log.Println(err)
	}
}
