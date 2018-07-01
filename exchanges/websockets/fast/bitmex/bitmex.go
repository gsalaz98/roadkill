package bitmex

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/gsalaz98/roadkill/orderbook"
)

// Settings : Structure is used to load settings into the application.
type Settings struct {
	connURL  string
	headers  http.Header
	messages []map[string]string
	conn     *websocket.Conn
	symbols  []string
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
	s.conn.WriteJSON(map[string]interface{}{
		"op":   "unsubscribe",
		"args": symArgs,
	})
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

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the user or server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output chan orderbook.Delta) {
	var (
		tickBytes []byte
	)

	// Read the first couple of messages; They don't have any useful information in them
	// N.B. - This doesn't parse any orderbook data, just meaningless ticks
	for i := 0; i <= len(s.symbols); i++ {
		_, _, _ = s.conn.ReadMessage()
	}

	// We will be parsing all of our data from the byte array for performance purposes
	for {
		_, tickBytes, _ = s.conn.ReadMessage()

		// Put the more used path first before to optimize runtime

		// This compares the first character in the second key. In this instance, we want to check
		// if this is a partial (i.e. orderbook snapshot), which transmits a "key" field. We check
		// for equality to the first character, 'k'. Orderbook updates, deletes, and trades

		if tickBytes[24] != 'k' {

		} else {

		}
	}
}
