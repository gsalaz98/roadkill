package main

import (
	"gitlab.com/CuteQ/roadkill/exchanges/poloniex"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

func main() {
	var receiver = make(chan orderbook.Delta)
	polo := poloniex.DefaultSettings
	polo.Initialize("BTC_ETH", "BTC_XMR", "BTC_XRP", "BTC_LTC", "BTC_STR")
	polo.ReceiveMessageLoop(receiver)

	//bitm := bitmex.DefaultSettings
	//bitm.Initialize("XBTUSD", "ETHM18")
}
