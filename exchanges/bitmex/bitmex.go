package bitmex

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

// Settings : Structure is used to load settings into the application.
type Settings struct {
	connURL  string
	headers  http.Header
	messages []map[string]string
	conn     *websocket.Conn
}

// DefaultSettings : Setup a simple skeleton of the Settings struct for ease of use
var DefaultSettings = Settings{
	connURL: "wss://www.bitmex.com/realtime",
	headers: http.Header{},
}

// CreateConnection : Creates a websocket connection and returns the connection object
func (s *Settings) CreateConnection() {
	var c websocket.Dialer
	conn, _, err := c.Dial(s.connURL, s.headers)

	if err != nil {
		fmt.Println("Error in connection: ", err)
	}
	s.conn = conn
}

// SendMessages : Sends multiple messages to the client. Not appropriate for BitMEX, but will
// be left in for compatability and modularity purposes.
func (s *Settings) SendMessages(messages []map[string]interface{}) {
	for _, message := range messages {
		s.conn.WriteJSON(message)
	}
}

// SubscribeWizard : Creates the subscription messages and stores it in the
// Settings struct based off a list of assets passed as variadic string parameters
func (s *Settings) SubscribeWizard(symbols ...string) {
	// TODO: Work on getting available futures contracts and work on rollover
	symArgs := make([]string, len(symbols))
	for i, symbol := range symbols {
		// Use a bytes Buffer to do fast string concatenation
		argBuf := bytes.Buffer{}
		argBuf.WriteString("orderBookL2:")
		argBuf.WriteString(symbol)

		symArgs[i] = argBuf.String()
	}
	message := map[string]interface{}{
		"op":   "subscribe",
		"args": symArgs,
	}
	s.conn.WriteJSON(message)
}

// UnsubscribeWizard : Accepts variadic paramters allowing you to unsubscribe from multiple assets at the same time.
// This is identical to the SubscribeWizard method, only that the argument passed to "op" is "unsubscribe"
func (s *Settings) UnsubscribeWizard(symbols ...string) {
	symArgs := make([]string, len(symbols))
	for i, symbol := range symbols {
		// Use a bytes Buffer to do fast string concatenation
		argBuf := bytes.Buffer{}
		argBuf.WriteString("orderBookL2:")
		argBuf.WriteString(symbol)

		symArgs[i] = argBuf.String()
	}
	message := map[string]interface{}{
		"op":   "unsubscribe",
		"args": symArgs,
	}
	s.conn.WriteJSON(message)
}

// Unsubscribe : Stops receiving data for the asset `symbol`
func (s *Settings) Unsubscribe(symbol string) {
	// Offload the work to the UnsubscribeWizard by passing in the symbol as a single variadic parameter
	s.UnsubscribeWizard([]string{symbol}...)
}

// Initialize : Initializes the entire structure automatically and calls methods required to run perfectly.
// Requires symbols you want to subscribe to to be passed as variadic parameters for the `Settings.SubscribeWizard` method.
// Stops short of calling the ReceiveMessageLoop method.
func (s *Settings) Initialize(symbols ...string) {
	s.CreateConnection()
	s.SubscribeWizard(symbols...)
}

// parseOrderbookTick : Private method, only to be called from the method `SubscribeWizard`
func (s *Settings) parseOrderbookTick(symbols ...string) {

}

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the user or server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output chan orderbook.Delta) {
	var (
		bitmexTick orderbook.IBitMexTick
		bitmexSnapshot ordrebook.Snapshot
		tickBytes  []byte
	)

	for i := 0; i <= len(symbols); i++ {
		_, _, _ = s.conn.ReadMessage()
	}

	for {
		_, tickBytes, _ = s.conn.ReadMessage()
		bitmexTick.UnmarshalJSON(tickBytes)

		switch bitmexTick.Action {
		case "partial":
			for entryIndex, level := range bitmexTick.Data {
				switch level.Side {
				case "Buy":
				case "Sell"
				}
			}
		case "update":
			for entryIndex, level := range bitmexTick.Data {
			
			}
		case "insert":
			for entryIndex, level := range bitmexTick.Data {

			}
		case "delete":
			for entryIndex, level := range bitmexTick.Data {

			}
		}
	}
}
