package main

import (
	"net/http"

	"gitlab.com/CuteQ/roadkill/exchanges"
)

func main() {
	var (
		header   http.Header
		messages []map[string]string
	)
	messages = append(messages, map[string]string{
		"command": "subscribe",
		"channel": "BTC_ETH",
	})
	messages = append(messages, map[string]string{
		"command": "subscribe",
		"channel": "BTC_XMR",
	})
	poloConnection := poloniex.CreateConnection(header)
	poloniex.SendMessage(poloConnection, messages)
	poloniex.ReceiveMessageLoop(poloConnection)
}
