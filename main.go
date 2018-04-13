package main

import (
	"gitlab.com/CuteQ/roadkill/exchanges"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

func main() {
	receiver := make(chan orderbook.Delta)

	polo := poloniex.DefaultSettings
	polo2 := poloniex.DefaultSettings

	polo.Initialize("BTC_ETH", "BTC_XMR")
	polo2.Initialize("USDT_BTC", "BTC_XRP")
	go polo.ReceiveMessageLoop(receiver)
	go polo2.ReceiveMessageLoop(receiver)

	for {
		//<-receiver
	}
}
