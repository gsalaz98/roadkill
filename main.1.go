package main

import (
	"gitlab.com/CuteQ/roadkill/exchanges/websockets/slow/gdax/clean"
	"gitlab.com/CuteQ/roadkill/orderbook"
	"gitlab.com/CuteQ/roadkill/orderbook/tectonic"
)

func main() {
	var (
		receiver        = make(chan orderbook.DeltaBatch, 1<<16)
		tConn           = tectonic.DefaultTectonic
		exchangeSymbols = make(map[string][]string, 64)
	)
	exchangeSymbols["gdax"] = []string{"BTC-USD", "ETH-USD"}

	tErr := tConn.Connect()

	if tErr != nil {
		panic(tErr)
	}

	for exchange, symbols := range exchangeSymbols {
		for _, symbol := range symbols {
			var dbName = exchange + ":" + symbol

			if !tConn.Exists(dbName) {
				tConn.Create(dbName)
			}
		}
	}

	gdax := gdaxslow.DefaultSettings
	gdax.Initialize(exchangeSymbols["gdax"]...)

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
