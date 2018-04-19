package poloniex

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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
	s.symbols = symbols

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
		s.assetTable[assetID] = assetPair
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

		snapshot[i].Timestamp = uint64(time.Now().UnixNano() / 1000)
		snapshot[i].StartSeq = 0
		snapshot[i].AskSide = jsonMessage.Orderbook[0]
		snapshot[i].BidSide = jsonMessage.Orderbook[1]
	}
}

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the user or server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output chan orderbook.Delta) {
	var (
		totalLoops            int
		smallestBracketLength = 1000
		minimumLoops          = 50
	)
	for {
		var (
			tickBytes []byte
			//assetCode uint16
		)
		_, tickBytes, _ = s.conn.ReadMessage()

		//start := time.Now()

		tickPositions := make([]int, int(len(tickBytes)/20)) // This is a reasonable estimate of the amount of ticks encapsulated within the byte array
		firstTickEncountered := false                        // Use this to signal whether we've encountered the first tick or not. For use in the loops below
		tickIndex := 0                                       // To keep an index of the last location we've inserted

		if minimumLoops < totalLoops { // We might have more than 100 symbols on this connection. Let's make sure we all opinions before we continue ;)
			// For efficiency purposes, we place the most access branch before the second one. This runs second!
			// *****************
			// We copy the code here to use the smallest distance between the start of the data and avoid wasting cycles iterating on nothing
			for char := smallestBracketLength; char < len(tickBytes); char++ {
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
					if char < smallestBracketLength {
						smallestBracketLength = char + 1 // Add one extra to get the index of the first *real* bracket
					}
				}
			}
		} else { // This loops a fixed amount before switching to the faster one. We do this to get the minimum shortest length of the first bracket.
			for char := 3; char < len(tickBytes); char++ {
				if firstTickEncountered && tickBytes[char] == '[' { // Ensures that we've encountered the first left bracket
					// Comments in the block above should explain the importance of all these elements
					tickPositions[tickIndex] = char
					tickIndex++

					if len(tickBytes) < char+32 {
						break
					}
					char += 32
				} else if !firstTickEncountered && tickBytes[char] == '[' {
					firstTickEncountered = true
					if char < smallestBracketLength {
						smallestBracketLength = char + 1
					}
				}
			}
		}
		// Loops over every bracket. For each bracket, parse all of the data
		for char := 0; char < tickIndex; char++ {
			var (
				dataIndex = tickPositions[char] // Gets starting index of left square bracket `[`
				side      uint8
				price     float64
				size      float64
				action    uint8
			)

			switch tickBytes[dataIndex+2] { // Update type. "o" is an update, and "t" is a trade event
			case 'o': // Book updates and removes
				// Orderbook information
				var (
					// Looping specific control flow variables
					priceEncountered = false
					sizeIters        int
				)
				// Orderbook information
				side = (1 + tickBytes[dataIndex+5]) << 4 // Increment side and make it equal to either orderbook.(IsBid || IsAsk)

				for dotIndex := dataIndex + 8; ; dotIndex++ {
					if tickBytes[dotIndex] == '.' {
						if priceEncountered { // Get everything from the size data iteration
							// Parses `size` byte slice to a floating point number
							size, _ = strconv.ParseFloat(string(tickBytes[dotIndex-sizeIters-1:dotIndex+9]), 64)
							action = side ^ orderbook.IsUpdate

							if size == 0 {
								action = side ^ orderbook.IsRemove
							}

							//fmt.Println(string(tickBytes))
							_ = orderbook.Delta{
								TimeDelta: 0,
								Seq:       0,
								Event:     action,
								Price:     price,
								Size:      size,
							}

							break // Passes control back to the bracket iterator

						} else {
							// First entry we hit will contain price data
							// This monster converts the price float enclosed within into a useable float64 value
							priceEncountered = true
							price, _ = strconv.ParseFloat(string(tickBytes[dataIndex+8:dotIndex+9]), 64)
							dotIndex += 12 // Adds the absolute minimum distance from the next '.'
						}
					} else if priceEncountered {
						sizeIters++ // Count how many times we've iterated searching for the size decimal point
					}
				}

			case 't':
				var ( // Declare looping flow control variables
					commasIterated int
					commaIndex     int
				)
				for dotIndex := dataIndex + 5; ; dotIndex++ {
					fmt.Println("Index: ", dotIndex, " ; len(tickBytes): ", len(tickBytes))
					if commasIterated == 0 && tickBytes[dotIndex] == ',' {
						side = (1 + tickBytes[dotIndex+1]) << 4 // Increment side and make it equal to either orderbook.(IsBid || IsAsk)
						action = side ^ orderbook.IsTrade

						commaIndex = dotIndex + 4 // Set first comma equal to the start of the price
						//dotIndex++                // Skips to the first possible decimal point in the price field
						commasIterated++
						fmt.Println(string(tickBytes))
						fmt.Println(string(tickBytes[dotIndex:]))

					}
					if tickBytes[dotIndex] == '.' {
						fmt.Println("Something happend :O", commasIterated)
						if commasIterated == 1 { // We should be getting to the price field by now
							price, _ = strconv.ParseFloat(string(tickBytes[commaIndex:dotIndex+9]), 64)
							commaIndex = dotIndex + 13 // commaIndex gets set to the first possible number in the set
							dotIndex = dotIndex + 13   // dotIndex becomes earliest possible '.'

							commasIterated++
						} else {
							size, _ = strconv.ParseFloat(string(tickBytes[commaIndex:dotIndex+9]), 64)

							fmt.Println(string(tickBytes))
							fmt.Println(string(tickBytes[commaIndex:]))

							//fmt.Println("TRADE: ", orderbook.Delta{
							//TimeDelta: 0,
							//Seq:       0,
							//Event:     action,
							//Price:     price,
							//Size:      size,
							//})
							break
						}
					}
				}
			}
		}
	}
}
