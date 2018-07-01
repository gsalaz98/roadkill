package poloniex

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/CuteQ/roadkill/orderbook/tectonic"

	"github.com/gorilla/websocket"
	"github.com/pquerna/ffjson/ffjson"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

// ExchangeName : Exchange name as an exportable constant
const ExchangeName string = "poloniex"

// Settings : Structure is used to load settings into the application.
type Settings struct {
	connURL  string
	headers  http.Header
	messages []map[string]string
	conn     *websocket.Conn
	symbols  []string

	assetTable map[float64]string // Poloniex-specific implementation, used to identify asset from tick data
}

// DefaultSettings : Setup a simple skeleton of the Settings struct for ease of use
var DefaultSettings = Settings{
	connURL:    "wss://api2.poloniex.com",
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
func (s *Settings) SendMessages(messages []map[string]string) {
	for _, message := range messages {
		s.conn.WriteJSON(message)
	}
}

// SubscribeWizard : Creates the subscription messages and stores it in the
// Settings struct based off a list of assets passed as variadic string parameters
func (s *Settings) SubscribeWizard(symbols ...string) {
	s.messages = make([]map[string]string, len(symbols))
	s.symbols = symbols[:]

	for _, asset := range symbols {
		s.messages = append(s.messages, map[string]string{
			"command": "subscribe",
			"channel": asset,
		})
	}
	s.SendMessages(s.messages)
	s.parseOrderbookSnapshots(symbols...)
}

// Subscribe : Sends a subscribe message to the server. We let the SubscribeWizard take care of that for us
func (s *Settings) Subscribe(symbol string) {
	s.SubscribeWizard([]string{symbol}...)
}

// UnsubscribeWizard : Accepts variadic paramters allowing you to unsubscribe from multiple assets at the same time
func (s *Settings) UnsubscribeWizard(symbols ...string) {
	s.messages = make([]map[string]string, len(symbols))
	for _, asset := range symbols {
		s.messages = append(s.messages, map[string]string{
			"command": "unsubscribe",
			"channel": asset,
		})
	}
	s.SendMessages(s.messages)
}

// Unsubscribe : Stops receiving data for the asset `symbol`
func (s *Settings) Unsubscribe(symbol string) {
	s.UnsubscribeWizard([]string{symbol}...)
}

// getAssetCodes : Retrieves integer representations of symbols on the exchange.
func (s *Settings) getAssetCodes() {
	var jsonData interface{}

	// HTTP Responses may be prone to failure. Make sure to error handle here.
	resp, err := http.Get("https://poloniex.com/public?command=returnTicker")
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	ffjson.Unmarshal(body, &jsonData)

	for assetPair, data := range jsonData.(map[string]interface{}) {
		assetID := data.(map[string]interface{})["id"].(float64)
		s.assetTable[float64(assetID)] = assetPair
	}
}

// Initialize : Initializes the entire structure automatically and calls methods required to run perfectly.
// Requires symbols you want to subscribe to to be passed as variadic parameters for the `Settings.SubscribeWizard` method.
// Stops short of calling the ReceiveMessageLoop method.
func (s *Settings) Initialize(symbols ...string) {
	s.getAssetCodes()
	s.CreateConnection()
	s.SubscribeWizard(symbols...)
}

// parseOrderbookSnapshots: Handles the initial ticks where the orderbook snapshots are sent.
// Only to be called from the SubscribeWizard method.
func (s *Settings) parseOrderbookSnapshots(symbols ...string) {
	var (
		snapshot    = make([]orderbook.Snapshot, len(symbols))
		jsonMessage orderbook.IPoloniexOrderbookSnapshot
		byteMessage []byte
	)
	// Poloniex for whatever reason sends `n` empty ticks before sending the actual data, where `n` = len(symbols)
	// Read those ticks and then get to the actual parsing
	for i := 0; i < len(symbols); i++ {
		_, _, _ = s.conn.ReadMessage()
	}
	for i := 0; i < len(symbols); i++ {
		_, byteMessage, _ = s.conn.ReadMessage()

		// Scan the byte array for the first occurance of the left curly brace.
		for char := 0; char < len(byteMessage); char++ {
			if byteMessage[char] == '{' {
				byteMessage = byteMessage[char:]
				byteMessage = byteMessage[:len(byteMessage)-3] // Orderbook ticks on Poloniex have 3 right square brackets before ending
				break
			}
		}
		jsonMessage.UnmarshalJSON(byteMessage)

		snapshot[i].Symbol = jsonMessage.CurrencyPair // TODO: Make it so that this string gets converted to a standardized format
		snapshot[i].Timestamp = uint64(time.Now().UnixNano() / 1000)

		// Run for loop to convert orderbook data into the desired format
		// TODO: Work on snapshots
		var (
		//orderbook = make(map[uint8]map[float64]float64, 2)
		//askSide   = make(map[float64]float64, len(jsonMessage.Orderbook[0]))
		//bidSide   = make(map[float64]float64, len(jsonMessage.Orderbook[1]))
		)
		//for side, sideEntry := range jsonMessage.Orderbook {
		//	switch side {
		//	case 0: // Ask side
		//		for _, askLevel := range sideEntry {
		//			fmt.Println(askLevel, sideEntry[askLevel])
		//		}
		//	case 1: // Bid side
		//		//for index, askEntry := range jsonMessage.Orderbook[0] {
		//		//	for _, askLevel := range askEntry {
		//		//		//levelSize, _ := strconv.ParseFloat(askEntry[askLevel], 64)
		//		//	}
		//		//}
		//	}
		//}
		//snapshot[i].BidSide = jsonMessage.Orderbook[1]
	}
}

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the user or server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output *chan orderbook.DeltaBatch) {
	var (
		seqCount = make(map[int64]uint32, len(s.symbols)) // Key seq by assetCode
		tectConn = tectonic.DefaultTectonic
	)
	err := tectConn.Connect()

	if err != nil {
		panic(err)
	}

	for _, symbol := range s.symbols {
		dbName := fmt.Sprintf("poloniex:%s", symbol)
		if tectConn.Exists(dbName) {
			continue
		}
		tectConn.Create("poloniex:" + symbol)
	}

	for {
		var (
			_, tickBytes, _ = s.conn.ReadMessage() // We store websocket frames here temporarily to parse the information contained within
			assetCode       int64                  // Asset code is a string that corresponds to a certain asset-pair on the poloniex exchange. Use this to access that information.
		)

		// Get the current asset code.
		for index, char := range tickBytes {
			if char == ',' {
				assetCode, _ = strconv.ParseInt(string(tickBytes[1:index]), 10, 64)
				break
			}
		}

		// If we receive a heartbeat code, we can just restart the loop and avoid allocating a massive slice
		if assetCode == 0 {
			continue
		}

		// Create an array to index the start of every single piece of data we get
		var (
			tickPositions        = make([]int, int(len(tickBytes)/20)) // This is a reasonable estimate of the amount of ticks encapsulated within the byte array
			firstTickEncountered bool                                  // Use this to signal whether we've encountered the first tick or not. For use in the loops below
			tickIndex            int                                   // Stores the *true* length of the ticks we've detected in our data
		)

		// Five pieces of data is the most we can push it before we start making assumptions
		for char := 5; char < len(tickBytes); char++ {
			// Get all of the left bracket indicies
			if firstTickEncountered && tickBytes[char] == '[' {
				tickPositions[tickIndex] = char
				tickIndex++

				// To Make sure we don't go out of bounds
				if len(tickBytes) < char+32 {
					break
				}
				char += 32 // Minimum length of a piece of update data. Updates are the tiniest piece of data that gets sent
			} else if !firstTickEncountered && tickBytes[char] == '[' {
				firstTickEncountered = true
			}
		}

		var (
			blockTimestamp = float64(time.Now().UnixNano()/1000) * 1e-6 // Format the timestamp to be inline with what TectonicDB wants
			deltas         = make([]*orderbook.Delta, tickIndex)        // We will return this data to the `output` channel as type `DeltaBatch`
			deltaCount     int
		)
		// Loops over every bracket. For each bracket, parse all of the data.
		// All of the parsing of data from the exchange happens here.
		// We construct deltas of each event and send them back to the `output` channel.
		for char := 0; char < tickIndex; char++ {
			var (
				dataIndex = &tickPositions[char] // Gets starting index of left square bracket `[`
				side      uint8
				price     float64
				size      float64
			)

			switch tickBytes[*dataIndex+2] { // Update type. "o" is an update, and "t" is a trade event
			case 'o': // Book updates and removes
				// Orderbook information
				var (
					// Looping specific control flow variables
					priceEncountered bool
					sizeIters        int
				)
				for dotIndex := *dataIndex + 8; ; dotIndex++ {
					if tickBytes[dotIndex] == '.' {
						if priceEncountered { // Get everything from the size data iteration
							// Parses `size` byte slice to a floating point number
							size, _ = strconv.ParseFloat(string(tickBytes[dotIndex-sizeIters-1:dotIndex+9]), 64)

							deltas[deltaCount] = &orderbook.Delta{
								Timestamp: blockTimestamp,
								Seq:       seqCount[assetCode],
								IsTrade:   false,
								IsBid:     ((1 + tickBytes[*dataIndex+5]) << 4) == orderbook.IsBid, // Increment side and make it equal to either orderbook.(IsBid || IsAsk)
								Price:     price,
								Size:      size,
							}

							seqCount[assetCode]++
							deltaCount++

							break // Passes control back to the bracket iterator

						} else {
							// First entry we hit will contain price data
							// This monster converts the price float enclosed within into a useable float64 value
							priceEncountered = true
							price, _ = strconv.ParseFloat(string(tickBytes[*dataIndex+8:dotIndex+9]), 64)
							dotIndex += 12 // Adds the absolute minimum distance from the next '.'
						}
					} else if priceEncountered {
						sizeIters++ // Count how many times we've iterated searching for the size decimal point
					}
				}

			case 't':
				var (
					// Declare looping flow control variables
					sideGathered  bool
					priceGathered bool
					commaIndex    int
				)
				for dotIndex := *dataIndex + 5; ; dotIndex++ {
					if !sideGathered && tickBytes[dotIndex] == ',' {
						side = (1 + tickBytes[dotIndex+1]) << 4 // Increment side and make it equal to either orderbook.(IsBid || IsAsk)

						commaIndex = dotIndex + 4 // Set first comma equal to the start of the price
						sideGathered = true       //
						dotIndex += 3             // Skips to the first possible decimal point in the price field

					} else if sideGathered && tickBytes[dotIndex] == '.' {
						if !priceGathered { // We should be getting to the price field by now
							price, _ = strconv.ParseFloat(string(tickBytes[commaIndex:dotIndex+9]), 64)
							commaIndex = dotIndex + 13 // commaIndex gets set to the first possible number in the set
							dotIndex += 12             // dotIndex becomes earliest possible '.' -- Set to one before '.' index because variable increments

							priceGathered = true
						} else {
							size, _ = strconv.ParseFloat(string(tickBytes[commaIndex:dotIndex+9]), 64)

							deltas[deltaCount] = &orderbook.Delta{
								Timestamp: blockTimestamp,
								Seq:       seqCount[assetCode],
								IsTrade:   true,
								IsBid:     side == orderbook.IsBid,
								Price:     price,
								Size:      size,
							}

							seqCount[assetCode]++
							deltaCount++

							break // Returns control to left bracket/tick iterator
						}
					}
				}
			}
		}
		// After we've finished parsing the tick, let's return a `DeltaBatch` to the `output` channel.
		if tickIndex != 0 {
			*output <- orderbook.DeltaBatch{
				Exchange: "poloniex",
				Symbol:   s.assetTable[float64(assetCode)],
				Deltas:   deltas,
			}
		}
	}
}
