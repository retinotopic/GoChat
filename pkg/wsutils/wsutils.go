package wsutils

import (
	"errors"
	"time"

	"github.com/fasthttp/websocket"
)

func KeepAlive(c *websocket.Conn, timeout time.Duration, ch chan<- error) {
	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	for {
		err := c.WriteMessage(websocket.PingMessage, []byte("keepalive"))
		if err != nil {
			ch <- errors.New("write message error")
		}
		time.Sleep(timeout / 2)
		if time.Since(lastResponse) > timeout {
			c.Close()
			ch <- errors.New("connection timeout")
		}
	}

}
