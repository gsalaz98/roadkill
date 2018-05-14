package gdaxslow

import (
	"encoding/json"
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
	symbols  []string

	// GDAX Specific settings and implementations
	assetInfo     map[string]map[string]float64 // Used to store index of symbols to reverse engineer price from ID
	ChannelType   []string                      // Valid values are: "level2", "heartbeat", "ticker"
	SingleChannel []string                      // For values that don't require a symbol/value to subscribe to
}

// DefaultSettings : Setup a simple skeleton of the Settings struct for ease of use
var DefaultSettings = Settings{
	connURL: "wss://ws-feed.gdax.com",
	headers: http.Header{},

	ChannelType: []string{"full"},
}

// CreateConnection : Creates a websocket connection and returns the connection object
func (s *Settings) CreateConnection() {
	var c websocket.Dialer
	conn, _, err := c.Dial(s.connURL, s.headers)

	if err != nil {
		fmt.Println("Error in establishing connection to GDAX. Verbose error output: ", err)
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
	msg := make(map[string]interface{}, 3)

	msg["type"] = "subscribe"
	msg["product_ids"] = symbols
	msg["channels"] = s.ChannelType

	s.conn.WriteJSON(msg)
}

// UnsubscribeWizard : Accepts variadic paramters allowing you to unsubscribe from multiple assets at the same time.
// This is identical to the SubscribeWizard method, only that the argument passed to "op" is "unsubscribe"
func (s *Settings) UnsubscribeWizard(symbols ...string) {
	msg := make(map[string]interface{}, 3)

	msg["type"] = "unsubscribe"
	msg["product_ids"] = symbols
	msg["channels"] = s.ChannelType

	s.conn.WriteJSON(msg)
}

// Unsubscribe : Stops receiving data for the asset `symbol`
func (s *Settings) Unsubscribe(symbol string) {
	// Offload the work to the UnsubscribeWizard by passing in the symbol as a single variadic parameter
	s.UnsubscribeWizard(symbol)
}

// Initialize : Initializes the entire structure automatically and calls methods required to run perfectly.
// Requires symbols you want to subscribe to to be passed as variadic parameters for the `Settings.SubscribeWizard` method.
// Stops short of calling the ReceiveMessageLoop method.
func (s *Settings) Initialize(symbols ...string) {
	// Begin boilerplate websocket connection initialization
	s.CreateConnection()
	s.SubscribeWizard(symbols...) // We pass in arguments here instead of using `Settings.symbols` to keep in line what we originally designed it for
}

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the user or server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output *chan orderbook.DeltaBatch) {
	for {
		var tickJSON interface{}

		_, tickBytes, _ := s.conn.ReadMessage()
		json.Unmarshal(tickBytes, &tickJSON)

		if tickBytes[9] == 'r' {
			continue
		}

		switch tickJSON["type"] {
		case "open": // LOB update

		case "done":
		case "match":
		}
	}
}
