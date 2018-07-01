package main

import (
	"github.com/gsalaz98/roadkill/exchanges/websockets/fast/poloniex"
	"github.com/gsalaz98/roadkill/exchanges/websockets/slow/bitmex"
	"github.com/gsalaz98/roadkill/exchanges/websockets/slow/gdax/clean"

	"github.com/gsalaz98/roadkill/orderbook"
	"github.com/gsalaz98/roadkill/orderbook/tectonic"
)

func main() {
	var (
		receiver        = make(chan orderbook.DeltaBatch, 1<<16)
		tConn           = tectonic.DefaultTectonic
		exchangeSymbols = make(map[string][]string, 64)
	)
	exchangeSymbols["gdax"] = []string{"BTC-USD", "ETH-USD", "BTC-ETH"}
	exchangeSymbols["bitmex"] = []string{"XBTUSD", "ETHM18", "XBT7D_U110", "XBT7D_D90"}
	exchangeSymbols["poloniex"] = []string{"BTC_ETH", "BTC_XMR", "BTC_ETC", "USDT_BTC", "USDT_ETH"}

	tErr := tConn.Connect()

	if tErr != nil {
		panic(tErr)
	}

	for exchange, symbols := range exchangeSymbols {
		for _, symbol := range symbols {
			dbName := exchange + ":" + symbol
			if !tConn.Exists(dbName) {
				tConn.Create(dbName)
			}
		}
	}

	// TODO: Automate this process
	// TODO: Initialize web server
	polo := poloniex.DefaultSettings
	polo.Initialize(exchangeSymbols["poloniex"]...)

	bitm := bitmexslow.DefaultSettings
	bitm.ChannelType = []string{"orderBookL2", "trade"}
	bitm.Initialize(exchangeSymbols["bitmex"]...)

	gdax := gdaxslow.DefaultSettings
	gdax.Initialize(exchangeSymbols["gdax"]...)

	go bitm.ReceiveMessageLoop(&receiver)
	go polo.ReceiveMessageLoop(&receiver)
	go gdax.ReceiveMessageLoop(&receiver)

	for {
		var (
			tickBatch = <-receiver
			dbName    = tickBatch.Exchange + ":" + tickBatch.Symbol
		)
		insErr := tConn.BulkAddInto(dbName, tickBatch.Deltas)
		// Catch any insertion errors here
		// TODO: Implement some logging here
		if insErr != nil {
			panic(insErr)
		}
	}
}
