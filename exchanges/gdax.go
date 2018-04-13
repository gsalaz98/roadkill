package gdax

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

// Settings : Structure is used to load settings into the application.
type Settings struct {
	connURL    string
	headers    http.Header
	messages   []map[string]string
	conn       *websocket.Conn
	assetTable map[float64]string
}

// DefaultSettings : Setup a simple skeleton of the Settings struct for ease of use
var DefaultSettings = Settings{
	connURL:    "wss://ws-feed.gdax.com",
	headers:    http.Header{},
	assetTable: make(map[float64]string),
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

// SendMessages : Sends multiple messages to the client
// TODO: The parameter type for `messages` might not be appropriate for GDAX.
func (s *Settings) SendMessages(messages []map[string]string) {
	for _, message := range messages {
		s.conn.WriteJSON(message)
	}
}

// SubscribeWizard : Creates the subscription messages and stores it in the
// Settings struct based off a list of assets passed as variadic string parameters
func (s *Settings) SubscribeWizard(symbols ...string) {
	// TODO: Implement SubscribeWizard for exchange GDAX
}

// UnsubscribeWizard : Accepts variadic paramters allowing you to unsubscribe from multiple assets at the same time
func (s *Settings) UnsubscribeWizard(symbols ...string) {
	// TODO: Work on GDAX implementation
}

// Unsubscribe : Stops receiving data for the asset `symbol`
func (s *Settings) Unsubscribe(symbol string) {
	// TODO: Work on GDAX implementation
}

// getAssetCodes : Retrieves integer representations of symbols on the exchange.
func (s *Settings) getAssetCodes() {
	var jsonData interface{}
	// TODO: we might not need this for GDAX.
}

// Initialize : Initializes the entire structure automatically and calls methods required to run perfectly.
// Requires symbols you want to subscribe to to be passed as variadic parameters for the `Settings.SubscribeWizard` method.
// Stops short of calling the ReceiveMessageLoop method.

// TODO: GDAX implementation
func (s *Settings) Initialize(symbols ...string) {
	// s.getAssetCodes()
	// s.CreateConnection()
	// s.SubscribeWizard(symbols...)
}

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output chan orderbook.Delta) {
	// TODO: Get poloniex current tick count by asking the database itself
	var tickMessage orderbook.ITickMessage
}
