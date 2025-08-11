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
	apphost := os.Getenv("APP_HOST")
	if len(apphost) == 0 {
		apphost = "localhost"
	}
	appport := os.Getenv("APP_PORT")
	if len(appport) == 0 {
		appport = "8080"
	}
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	logs, err := os.OpenFile(dir+"/logs/logs", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	testlogs, err := os.OpenFile(dir+"/logs/testlogs", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	// wsUrl := "ws://" + apphost + ":" + appport + "/connect"
	wsUrl := "ws://" + "localhost" + ":" + "80" + "/connect"
	chat := app.NewChat(bufstr.String(), wsUrl, 20, true, false, log.New(logs, bufstr.String()+" ", 0), log.New(testlogs, bufstr.String()+" ", 0))
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
