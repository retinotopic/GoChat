package main

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/retinotopic/GoChat/app"
	"github.com/rivo/tview"
)

func main() {
	bufstr := bytes.NewBuffer(make([]byte, 0, 50))
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		token := scanner.Text()
		io.WriteString(os.Stdout, token)
	}
	chat := chat.NewChat()
	chat.App = tview.NewApplication()
	err := chat.TryConnect()
	if err != nil {
		panic(err)
	}
	chat.MainFlex.AddItem(chat.Lists[0], 0, 1, true)
	go chat.StartEventUILoop()
	if err := chat.App.SetRoot(chat.MainFlex, true).Run(); err != nil {
		panic(err)
	}
}
