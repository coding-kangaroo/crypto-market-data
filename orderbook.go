package main

import "strconv"

type Orderbook struct {
	entries map[int][]OrderbookValue
}

type OrderbookEntry struct {
  key int
  val OrderbookValue
}

type OrderbookValue struct {
  Market Market
  //make price float
  Price string
}

type Market string
const(
  Binance Market = "Binance"
  Bithumb = "Bithumb"
)

func (orderbook Orderbook) start() chan<-OrderbookEntry {
	c := make(chan OrderbookEntry)
  go func() {
      for v := range c {
          orderbook.notify(v)
      }
  }()
  return c
}

func (orderbook Orderbook) notify(v OrderbookEntry) {
    orderbook.entries[v.key] = append(orderbook.entries[v.key], v.val)
}

func normalizeBithumbTime(time string) int {
    normalizedTime, _ := strconv.Atoi(time)
    return normalizedTime / 1000000
}
