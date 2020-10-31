package main

import (
	"log"
	"os"
	"os/signal"
	"github.com/gorilla/websocket"
  "encoding/json"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c, _, err := websocket.DefaultDialer.Dial("wss://pubwss.bithumb.com/pub/ws", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

  m := map[string]interface{}{"type":"orderbookdepth", "symbols":[]string{"BTC_KRW" , "ETH_KRW"}}
  b, _ := json.Marshal(m)

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
    default:
      err := c.WriteMessage(websocket.TextMessage, []byte(b))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}
