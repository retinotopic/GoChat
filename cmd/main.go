package main

import (
	server "GoChat/internal/server"
	"flag"
	"fmt"
	"log"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address to listen on")
	fromsmtp := flag.String("fromsmtp", "example@mail.com", "smtp sender")
	passwordsmtp := flag.String("passwordsmtp", "admin123", "password smtp")
	smtpHost := flag.String("smtpHost", "smtp.gmail.com", "host of smtp server")
	smtpPort := flag.String("smtpPort", "587", "port of smtp server")
	flag.Parse()
	fmt.Println(*addr, *fromsmtp, *passwordsmtp, *smtpHost, *smtpPort)
	srv := server.NewServer(*addr, *fromsmtp, *passwordsmtp, *smtpHost, *smtpPort)
	err := srv.Run()
	if err != nil {
		log.Println(err)
	}
}
