package main

import (
	"gitlab.com/CuteQ/roadkill/exchanges/websockets/slow/gdax"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

func main() {
	a := make(chan orderbook.DeltaBatch, 10)
	g := gdaxslow.DefaultSettings

	g.Initialize("BTC-USD", "ETH-USD")
	g.ReceiveMessageLoop(&a)
}
