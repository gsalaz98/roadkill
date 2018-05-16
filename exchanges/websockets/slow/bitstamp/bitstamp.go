package bitstampslow

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	ChannelType: []string{"level2", "matches"},
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
	var seqCount uint32
recvLoop:
	for {
		var (
			blockTimestamp        = float64(time.Now().UnixNano()/1000) * 1e-6
			_, tickBytes, readErr = s.conn.ReadMessage()
		)

		// Restarts the connection in case of a disconnection or unforseen error
		if readErr != nil {
			// I personally have problems running this quick hack for production systems.
			// I don't know how go handles functions running on the stack, especially for prolonged periods.
			// Perhaps we can fix this by sending a special message to the channel? TODO
			fmt.Println("Connection was closed")
			s.conn.Close()
			s.conn = nil
			s.Initialize(s.symbols...)
			s.ReceiveMessageLoop(output)
		}

		switch tickBytes[9] {
		case 's': // Snapshot event
			// TODO: Implement this
			continue recvLoop
		case 't': // Trade/match event
			var matchMessage orderbook.SlowGDAXMatches
			matchMessage.UnmarshalJSON(tickBytes)

			var (
				price, _ = strconv.ParseFloat(matchMessage.Price, 64)
				size, _  = strconv.ParseFloat(matchMessage.Size, 64)
				side     = orderbook.IsBid
			)
			if matchMessage.Side == "sell" {
				side = orderbook.IsAsk
			}

			*output <- orderbook.DeltaBatch{
				Exchange: "gdax",
				Symbol:   matchMessage.ProductID,
				Deltas: []*orderbook.Delta{&orderbook.Delta{
					Timestamp: blockTimestamp,
					Price:     price,
					Size:      size,
					Seq:       uint32(matchMessage.Sequence),
					IsTrade:   true,
					IsBid:     side == orderbook.IsBid,
				}},
			}

		case 'l': // Orderbook update event
			var updateMessage orderbook.SlowGDAXOrderbookUpdates
			updateMessage.UnmarshalJSON(tickBytes)

			deltaBatch := make([]*orderbook.Delta, len(updateMessage.Changes))

			for updateSeq, updateTick := range updateMessage.Changes {
				var (
					side     = orderbook.IsBid
					price, _ = strconv.ParseFloat(updateTick[1], 64)
					size, _  = strconv.ParseFloat(updateTick[2], 64)
				)
				if updateTick[0] == "sell" {
					side = orderbook.IsAsk
				}

				deltaBatch[updateSeq] = &orderbook.Delta{
					Timestamp: blockTimestamp,
					Price:     price,
					Size:      size,
					Seq:       seqCount, // TODO: Fix this
					IsTrade:   false,
					IsBid:     side == orderbook.IsBid,
				}
				seqCount++
			}

			*output <- orderbook.DeltaBatch{
				Exchange: "gdax",
				Symbol:   updateMessage.ProductID,
				Deltas:   deltaBatch,
			}
		}
	}
}
