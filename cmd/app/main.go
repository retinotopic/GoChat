package main

import (
	"bufio"
	"bytes"

	// "log"
	"os"

	"github.com/retinotopic/GoChat/app"
)

func main() {
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
	wsUrl := wsstr + "://" + os.Getenv("APP_HOST") + ":" + os.Getenv("APP_PORT") + "/connect"
	chat, err := app.NewChat(bufstr.String(), wsUrl, 30, true)
	if err != nil {
		panic(err)
	}
	err = chat.Run()
	if err != nil {
		panic(err)
	}
}
