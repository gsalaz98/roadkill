package poloniex

import (
	"bytes"
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
	for {
		var (
			tickBytes []byte
			buf       bytes.Buffer
		)
		_, tickBytes, _ = s.conn.ReadMessage()

		buf.Write([]byte(`{"assetCode":`))

		// This is painful to type out, but we're doing this for the sake of efficiency
		if tickBytes[2] == ',' {
			buf.Write(tickBytes[1:2])
		} else if tickBytes[3] == ',' {
			buf.Write(tickBytes[1:3])
		} else if tickBytes[4] == ',' {
			buf.Write(tickBytes[1:4])
		}
		buf.Write([]byte(`,"data":`))

		for char := 13; ; char++ {
			// Once we have the first occurence, check for others
			if tickBytes[char] == '[' {
				var localBuf bytes.Buffer
				localBuf.Write(tickBytes[0:char])

				for dataChar := char + 1; ; dataChar++ {
					// Check for various conditions that indicate a non-string entry
					if tickBytes[dataChar] == '[' {
						switch tickBytes[dataChar+2] { // Update type. "o" is an update, and "t" is a trade event
						case 'o':
							var (
								event uint8
								side  = tickBytes[dataChar+5] // without fail, this one will always be 5 chars away
								size  float64
								price float64

								priceIndex int
								sizeIndex  int
							)
							for pricePoint := dataChar + 8; pricePoint < 16; pricePoint++ {
								if tickBytes[pricePoint] == '.' {
									// This monster converts the price float enclosed within into a useable float64 value
									price = math.Float64frombits(binary.LittleEndian.Uint64(tickBytes[dataChar+8 : pricePoint+8]))
									priceIndex = pricePoint + 8
									break
								}
							}
							for sizePoint := priceIndex + 4; sizePoint < 16; sizePoint++ {
								if tickBytes[sizePoint] == '.' {
									// Parses `size` byte slice to a floating point number
									size = math.Float64frombits(binary.LittleEndian.Uint64(tickBytes[priceIndex+4 : sizePoint+8]))
									char = sizePoint + 11
									break
								}
							}

						case 't':
						}
					}
				}
			}
		}
	}
}
