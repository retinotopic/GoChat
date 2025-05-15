package main

import (
	"bufio"
	"bytes"
	"log"
	"time"

	// "log"
	// "flag"
	"github.com/retinotopic/GoChat/app"
	"os"
)

func main() {
	// idk := flag.String("identKey", "user1493key3051example", "identity key")
	// flag.Parse()
	bufstr := bytes.NewBuffer(make([]byte, 0, 50))
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		token := scanner.Text()
		bufstr.WriteString(token)
		break
	}
	wsstr := "ws"
	if os.Getenv("SSL_ENABLE") == "true" {
		wsstr = "wss"
	}
	apphost := os.Getenv("APP_HOST")
	if len(apphost) == 0 {
		apphost = "localhost"
	}
	appport := os.Getenv("APP_PORT")
	if len(appport) == 0 {
		appport = "8080"
	}
	f, err := os.OpenFile(bufstr.String(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	wsUrl := wsstr + "://" + apphost + ":" + appport + "/connect"
	chat := app.NewChat(bufstr.String(), wsUrl, 20, true, log.New(f, bufstr.String(), 0))
	errch := chat.TryConnect()
	recnct := 0
	go chat.ProcessEvents()
	for {
		select {
		case _, ok := <-errch:
			if !ok && recnct < 10 {
				chat.App.Stop()
				errch = chat.TryConnect()
				log.Println("trying to reconnect...")
				time.Sleep(time.Second * 3)
				go chat.ProcessEvents()
			} else {
				log.Fatalln("Max Reconnect tries exceeded")
			}
		}
	}
}
