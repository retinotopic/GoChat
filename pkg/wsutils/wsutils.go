package wsutils

import (
	"time"

	"github.com/fasthttp/websocket"
)

func KeepAlive(c *websocket.Conn, timeout time.Duration, ch chan<- bool) {
	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	for {
		err := c.WriteMessage(websocket.PingMessage, []byte("keepalive"))
		if err != nil {
			ch <- true
			return
		}
		time.Sleep(timeout / 2)
		if time.Since(lastResponse) > timeout {
			c.Close()
			ch <- true
			return
		}
	}

}
