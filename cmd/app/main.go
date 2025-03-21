package main

import (
	"bufio"
	"bytes"
	"io"
	"os"

	// "github.com/jackc/pgx/v5/stdlib"
	"github.com/retinotopic/GoChat/app"
	// "github.com/rivo/tview"
)

func main() {
	bufstr := bytes.NewBuffer(make([]byte, 0, 50))
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		token := scanner.Text()
		io.WriteString(os.Stdout, token)
		io.WriteString(bufstr, token)
	}
	_, err := app.NewChat(bufstr.String(), os.Getenv("CHAT_CONNECT_ADDRESS"))
	if err != nil {
		panic(err)
	}

}
