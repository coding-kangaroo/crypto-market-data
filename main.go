package main

import (
	"log"
	"os"
	"os/signal"
  "sync"
  "encoding/json"
  "strconv"
)

type Market string
const(
  Binance Market = "Binance"
  Bithumb = "Bithumb"
)

type OrderbookEntry struct {
  Market Market
  //make price float
  Price string
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

  orderbook := make(map[int][]OrderbookEntry)
  lock := sync.RWMutex{}

  go storeMessages(binanceC, Binance, orderbook, lock)
  go storeMessages(bithumbC, Bithumb, orderbook, lock)

  <-interrupt
    log.Println("interrupt")
    log.Println(orderbook)
    return
}

func storeMessages(channel chan []byte, market Market, orderbook map[int][]OrderbookEntry, lock sync.RWMutex) {
  for {
    select {
    case message := <- channel:
      lock.Lock()
      switch market {
      case Binance:
        binanceMsg := BinanceMessage{}
        json.Unmarshal(message, &binanceMsg)
        time := binanceMsg.Data.E
        //log.Printf("recv: %s", binanceMsg)
        tmp := append(orderbook[time], OrderbookEntry{market, binanceMsg.Data.P})
        log.Println(tmp)
      case Bithumb:
        bithumbMsg := BithumbMessage{}
        json.Unmarshal(message, &bithumbMsg)
        //log.Printf("recv: %s", bithumbMsg)
        if (!bithumbMsg.IsEmpty()) {
            //price, _ := strconv.ParseFloat(bithumbMsg.Content.List[0].Price, 64)
            time := normalizeBithumbTime(bithumbMsg.Content.Datetime)
            tmp := append(orderbook[time], OrderbookEntry{market, bithumbMsg.Content.List[0].Price})
            log.Println(tmp)
        }
      }
      //log.Printf("recv: %s", message)
      lock.Unlock()
    }
  }
}

func normalizeBithumbTime(time string) int {
    normalizedTime, _ := strconv.Atoi(time)
    return normalizedTime / 1000
}
