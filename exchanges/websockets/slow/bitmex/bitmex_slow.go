package bitmexslow

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gsalaz98/roadkill/orderbook"
	"github.com/pquerna/ffjson/ffjson"
)

// Settings : Structure is used to load settings into the application.
type Settings struct {
	connURL  string
	headers  http.Header
	messages []map[string]string
	conn     *websocket.Conn
	symbols  []string

	// BitMEX Specific settings and implementations
	assetInfo     map[string]map[string]float64 // Used to store index of symbols to reverse engineer price from ID
	ChannelType   []string                      // Valid values are: "orderBookL2", "orderBookL2_25". Make sure "orderBook*" is first entry in array
	SingleChannel []string                      // For values that don't require a symbol/value to subscribe to
}

// DefaultSettings : Setup a simple skeleton of the Settings struct for ease of use
var DefaultSettings = Settings{
	connURL: "wss://www.bitmex.com/realtime",
	headers: http.Header{},

	ChannelType:   []string{"orderBookL2", "trade"},
	SingleChannel: []string{"liquidation"},
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
	symArgs := make([]string, len(symbols)*len(s.ChannelType))
	argCount := 0

	for _, channel := range s.ChannelType {
		// Loops over every channel we want to subscribe to, and adds it to the args buffer
		for _, symbol := range symbols {
			// Use a bytes Buffer to do fast string concatenation
			argBuf := bytes.Buffer{}
			argBuf.WriteString(channel)
			argBuf.WriteByte(':')
			argBuf.WriteString(symbol)

			symArgs[argCount] = argBuf.String()
			argCount++
		}
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
	// TODO: Work on getting available futures contracts and work on rollover
	symArgs := make([]string, len(symbols)*len(s.ChannelType))
	argCount := 0

	for _, channel := range s.ChannelType {
		// Loops over every channel we want to subscribe to, and adds it to the args buffer
		for _, symbol := range symbols {
			// Use a bytes Buffer to do fast string concatenation
			argBuf := bytes.Buffer{}
			argBuf.WriteString(channel)
			argBuf.WriteByte(':')
			argBuf.WriteString(symbol)

			symArgs[argCount] = argBuf.String()
			argCount++
		}
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

// local function getAssetIndexes: loads asset indexes into a map variable defined in the settings structure
func (s *Settings) getAssetIndexes() {
	var (
		bodyJSON    []interface{}
		legacyTicks = map[string]float64{
			"XBTUSD": 0.01,
			"XBTZ17": 0.1,
			"XBJZ17": 1.0,
		}
	)

	// URL allows us to get the index of the symbol used to calculate the ID
	response, _ := http.Get("https://www.bitmex.com/api/v1/instrument?columns=symbol,tickSize&start=0&count=500")
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	ffjson.Unmarshal(body, &bodyJSON)

	for _, symbol := range s.symbols {
		// Initialize inner map to index data
		s.assetInfo[symbol] = make(map[string]float64, 2)
		for index, symbolData := range bodyJSON {
			// Use type assertion to access the data contained within
			if symbolData.(map[string]interface{})["symbol"].(string) == symbol {
				s.assetInfo[symbol]["index"] = float64(index)
				// Implement legacy tick support. `XBTUSD` perpetual contract requires this to get proper price data.
				if legacyTicks[symbol] == 0.00 {
					s.assetInfo[symbol]["tickSize"] = symbolData.(map[string]interface{})["tickSize"].(float64)
				} else {
					s.assetInfo[symbol]["tickSize"] = legacyTicks[symbol]
				}
			}
		}
	}
}

// Initialize : Initializes the entire structure automatically and calls methods required to run perfectly.
// Requires symbols you want to subscribe to to be passed as variadic parameters for the `Settings.SubscribeWizard` method.
// Stops short of calling the ReceiveMessageLoop method.
func (s *Settings) Initialize(symbols ...string) {
	// BitMEX specific implementations
	// *********
	// Check that our first channel entry in `ChannelType`
	if s.ChannelType[0][0:11] != "orderBookL2" {
		panic("First argument in `ChannelType` must be `orderBookL2` or `orderBookL2_25` - Exiting...")
	}
	// Initialize symbols list in `Settings` structure and create map
	s.symbols = symbols
	s.assetInfo = make(map[string]map[string]float64, len(symbols))
	s.getAssetIndexes()

	// Begin boilerplate websocket connection initialization
	s.CreateConnection()
	s.SubscribeWizard(symbols...) // We pass in arguments here instead of using `Settings.symbols` to keep in line what we originally designed it for

}

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the user or server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output *chan orderbook.DeltaBatch) {
	var (
		orderbookTableName = s.ChannelType[0]
		tradeTableName     = "trade"

		partialAction = "partial"

		snapshots = make(map[string][][]orderbook.Delta, len(s.symbols))

		noPartialTicks   int
		incrementPartial = true

		sideMap = map[string]uint8{
			"Buy":  orderbook.IsBid,
			"Sell": orderbook.IsAsk,
		}
		seqCount = make(map[string]uint32, len(s.symbols))
	)

	// Let's construct the slices to be contained within snapshots
	for _, symbol := range s.symbols {
		snapshots[symbol] = make([][]orderbook.Delta, 50)
	}

	// Read the first couple of messages; They don't have any useful information in them
	// N.B. - This doesn't parse any orderbook data, just meaningless ticks
	for i := 0; i <= len(s.symbols); i++ {
		_, _, _ = s.conn.ReadMessage()
	}

	// BEGIN MAIN LOOP
	for {
		// Initialize variables we need. Some of these return errors, but we won't do anything with them, we just need to call the method.
		var (
			_, tickBytes, connErr = s.conn.ReadMessage()
			tick                  = &orderbook.IBitMexTick{}
			_                     = tick.UnmarshalJSON(tickBytes)
			deltas                = make(map[string][]orderbook.Delta, len(s.symbols)) // Use this structure instead of list of deltas because sometimes BitMEX returns two different symbol's data in the same message
			deltaCount            = make(map[string]int, len(tick.Data))
		)

		// Let's restart the connection if we get an error with the connection
		if connErr != nil {
			s.conn.Close()
			s.Initialize(s.symbols...)
			s.ReceiveMessageLoop(output)
		}

		// If we get a empty tick or a partial, let's skip it (for the meanwhile)
		if tick.Action == "" || tick.Action == "partial" {
			continue
		}

		blockTimestamp := float64(time.Now().UnixNano()/1000) * 1e-6

		// Orderbook events. Just to make sure we don't pick up a liquidation event or anything of the sorts.
		if tick.Table == orderbookTableName {
			// We check that the frame sent is a partial (fragmented section of orderbook) and that we haven't looped it over 50 times
			for _, update := range tick.Data {

				if deltas[update.Symbol] == nil {
					deltas[update.Symbol] = make([]orderbook.Delta, len(tick.Data))
					deltaCount[update.Symbol] = 0
				}

				deltas[update.Symbol][deltaCount[update.Symbol]] = orderbook.Delta{
					Timestamp: blockTimestamp,
					Seq:       seqCount[update.Symbol],
					IsTrade:   false,
					IsBid:     orderbook.IsBid == sideMap[update.Side],
					Price:     ((1.0e+8 * s.assetInfo[update.Symbol]["index"]) - float64(update.ID)) * s.assetInfo[update.Symbol]["tickSize"],
					Size:      float64(update.Size),
				}
				seqCount[update.Symbol]++
				deltaCount[update.Symbol]++
			}

			// TODO: Is this really the best way to handle the situation? Maybe we can just move it into its own section. Ponder on this...
			// We check that the frame sent is a partial (fragmented section of orderbook) and that we haven't looped it over 50 times
			if incrementPartial {
				if tick.Action == partialAction {
					//symbol := tick.Data[0].Symbol
					noPartialTicks = 0

					// Adds partial data to the list of snapshots belonging to the 'symbol'
					//snapshots[symbol] = append(snapshots[symbol], deltas)

					continue
				}
				if noPartialTicks > 50 {
					// TODO: Implement Orderbook partials converter to standard `orderbook.Snapshot` interface
					incrementPartial = false
				}
				noPartialTicks++ // we have this increment at the end of the loop so that we can keep count and know when to stop checking for partials
			}

			// After we've cleared all of the tests to make sure we're not processing `partial` events, let's send back the data we've gathered into the `output` channel.
			for symbol, tickDeltas := range deltas {
				*output <- orderbook.DeltaBatch{
					Exchange: "bitmex",
					Symbol:   symbol,
					Deltas:   &tickDeltas,
				}
			}

		} else if tick.Table == tradeTableName {
			for _, trade := range tick.Data {
				if deltas[trade.Symbol] == nil {
					deltas[trade.Symbol] = make([]orderbook.Delta, len(tick.Data))
					deltaCount[trade.Symbol] = 0
				}

				deltas[trade.Symbol][deltaCount[trade.Symbol]] = orderbook.Delta{
					Timestamp: blockTimestamp,
					Seq:       seqCount[trade.Symbol],
					IsTrade:   true,
					IsBid:     orderbook.IsBid == sideMap[trade.Side],
					Price:     float64(trade.Price),
					Size:      float64(trade.Size),
				}

				seqCount[trade.Symbol]++
				deltaCount[trade.Symbol]++
			}
			// After we're done looping, send back the trade ticks as a `DeltaBatch` type
			for symbol, tickDeltas := range deltas {
				*output <- orderbook.DeltaBatch{
					Exchange: "bitmex",
					Symbol:   symbol,
					Deltas:   &tickDeltas,
				}
			}
		}
	}
}
