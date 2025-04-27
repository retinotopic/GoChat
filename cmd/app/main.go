package main

import (
	"bufio"
	"bytes"
	"log"

	// "log"
	"os"

	"github.com/retinotopic/GoChat/app"
)

func main() {
	f, err := os.Create("logs")
	if err != nil {
		panic(err)
	}
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
	wsUrl := wsstr + "://" + apphost + ":" + appport + "/connect"
	chat, errch := app.NewChat(bufstr.String(), wsUrl, 30, true, log.New(f, "app log: ", log.LstdFlags))
	go chat.Run()
	for err := range errch {
		panic(err)
	}
}
