package main

import (
	"log"
	"os"
	"os/signal"
  "encoding/json"
)

func main() {
  interrupt := make(chan os.Signal, 1)
  signal.Notify(interrupt, os.Interrupt)

  binanceAddr := "wss://stream.binance.com/stream"
  binanceParams := map[string]interface{}{"method":"SUBSCRIBE", "params":[]string{"btcusdt@aggTrade"}, "id":1}
  binanceC := make(chan []byte)
  go getMessages(binanceAddr, binanceParams, binanceC)

  bithumbAddr := "wss://pubwss.bithumb.com/pub/ws"
  bithumbParams := map[string]interface{}{"type":"orderbookdepth", "symbols":[]string{"BTC_KRW"}}
  bithumbC := make(chan []byte)
  go getMessages(bithumbAddr, bithumbParams, bithumbC)

  orderbook := Orderbook{
		entries: map[int][]OrderbookValue{},
	}
  orderbookChannel := orderbook.start()

  go storeMessages(binanceC, Binance, orderbookChannel)
  go storeMessages(bithumbC, Bithumb, orderbookChannel)

  <-interrupt
    log.Println("interrupt")
    log.Println(orderbook)
    return
}

func storeMessages(channel chan []byte, market Market, orderbookChannel chan<-OrderbookEntry) {
  for {
    select {
    case message := <- channel:
      switch market {
      case Binance:
        binanceMsg := BinanceMessage{}
        json.Unmarshal(message, &binanceMsg)
        time := binanceMsg.Data.E
        //log.Printf("recv: %s", binanceMsg)
				orderbookChannel <- OrderbookEntry{time / 1000, OrderbookValue{market, binanceMsg.Data.P}}
      case Bithumb:
        bithumbMsg := BithumbMessage{}
        json.Unmarshal(message, &bithumbMsg)
        //log.Printf("recv: %s", bithumbMsg)
        if (!bithumbMsg.IsEmpty()) {
            //price, _ := strconv.ParseFloat(bithumbMsg.Content.List[0].Price, 64)
            time := normalizeBithumbTime(bithumbMsg.Content.Datetime)
						orderbookChannel <- OrderbookEntry{time, OrderbookValue{market, bithumbMsg.Content.List[0].Price}}
        }
      }
      log.Printf("recv: %s", message)
    }
  }
}

type BinanceMessage struct {
  Data struct {
    E int
    S string
    P string
  }
}

type BithumbMessage struct {
  Content struct {
    List []BithumbMessageDetail
    Datetime string
  }
}

type BithumbMessageDetail struct {
  Symbol string
  OrderType string
  Price string
}

func (bithumbMsg BithumbMessage) IsEmpty() bool {
  return bithumbMsg.Content.List == nil || len(bithumbMsg.Content.List) <= 0
}
