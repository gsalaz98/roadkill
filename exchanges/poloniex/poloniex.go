package poloniex

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pquerna/ffjson/ffjson"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

// ExchangeName : Exchange name as an exportable constant
const ExchangeName string = "poloniex"

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

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output chan orderbook.Delta) {
	var (
		totalLoops            int
		smallestBracketLength = 1000
	)
	for {
		var (
			tickBytes []byte
			//assetCode uint16
			//deltas    []orderbook.Delta
		)
		_, tickBytes, _ = s.conn.ReadMessage()

		tickPositions := make([]int, len(tickBytes)/20)
		tickCount := -1

		start := time.Now()

		if totalLoops < 100 { // We might have more than 100 symbols on this connection. Let's make sure we all opinions before we continue ;)
			for char := 3; char < len(tickBytes); char++ {
				// Get all of the left bracket indicies
				if tickBytes[char] == '[' {
					if tickCount == -1 { // There's 2 left brackets before the real data
						tickCount++
						continue
					}
					tickPositions[tickCount] = char
					tickCount++

					if tickCount == 0 && char < smallestBracketLength {
						smallestBracketLength = char
					}
					if len(tickBytes) < char+33 {
						break
					}
					char += 33
				}
			}
		} else { // We copy the code here to use the smallest distance between the start of the data and avoid wasting cycles iterating on nothing
			tickCount = 0
			for char := smallestBracketLength; char < len(tickBytes); char++ {
				// Get all of the left bracket indicies
				if tickBytes[char] == '[' {
					tickPositions[tickCount] = char
					tickCount++

					if len(tickBytes) < char+33 {
						break
					}
					char += 32 // Minimum length of a piece of update data. Updates are the tiniest piece of data that gets sent
				}
			}
		}
	tickIter:
		for char := 0; char <= tickCount; char++ {
			dataChar := tickPositions[char]
			// Check for various conditions that indicate a non-string entry
			var (
				event uint8
				size  float64
				price float64
			)
			switch tickBytes[dataChar+2] { // Update type. "o" is an update, and "t" is a trade event
			case 'o': // Updates
				var (
					side       = tickBytes[dataChar+5] // without fail, this one will always be 5 chars away
					priceIndex int
				)
				for pricePoint := dataChar + 8; pricePoint < 32; pricePoint++ {
					if tickBytes[pricePoint] == '.' {
						// This monster converts the price float enclosed within into a useable float64 value
						price = math.Float64frombits(binary.LittleEndian.Uint64(tickBytes[dataChar+8 : pricePoint+8]))
						priceIndex = pricePoint + 8
						break
					}
				}
				for sizePoint := priceIndex + 4; sizePoint < 64; sizePoint++ {
					if tickBytes[sizePoint] == '.' {
						// Parses `size` byte slice to a floating point number
						size = math.Float64frombits(binary.LittleEndian.Uint64(tickBytes[priceIndex+4 : sizePoint+8]))
						char = sizePoint + 11 // Let's set the char cursor ready for the next entry in the message

						if side == orderbook.PoloniexBid {
							if size == 0.0 {
								event = orderbook.IsBidRemove
							} else {
								event = orderbook.IsBidUpdate
							}
						} else if side == orderbook.PoloniexAsk {
							if size == 0.0 {
								event = orderbook.IsAskRemove
							} else {
								event = orderbook.IsAskUpdate
							}
						}
						// Add a terminating clause here to prevent accessing an index that doesn't exist.
						if len(tickBytes) < sizePoint+16 {
							break tickIter
						}
						//fmt.Println(event, side, size, price, sizePoint, len(tickBytes))
						_ = orderbook.Delta{
							TimeDelta: 0,
							Seq:       0,
							Event:     event,
							Price:     price,
							Size:      size,
						}
					}
				}

			case 't':
				///var (
				///	side  uint8
				///	event uint8
				///	price float64
				///	size  float64
				///)
				///// Find the first decimal point, then get the order side (bid/ask) by using decimal index
				///for pricePoint := char + 5; pricePoint < 64; pricePoint++ {

				///}
			}
		}
		end := time.Now()
		fmt.Println(end.Sub(start))
	}
}
