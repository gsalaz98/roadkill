package poloniex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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
	unsubMsg := map[string]string{
		"command": "unsubscribe",
		"channel": symbol,
	}
	s.conn.WriteJSON(unsubMsg)
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
	json.Unmarshal(body, &jsonData)

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

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the server.
// It is recommended that you call this method concurrently.
func (s *Settings) ReceiveMessageLoop(output chan orderbook.Delta) {
	// TODO: Get poloniex current tick count by asking the database itself
	var tickMessage orderbook.ITickMessage

	for {
		// TODO: Consider replacing this method with a `conn.ReadMessage(&tickMessage)` call instead.
		// Might yield better performance in the longer run if we handle all errors and the data ourselves.
		s.conn.ReadJSON(&tickMessage)
		msgLength := len(tickMessage)

		if msgLength < 3 {
			continue
		}

		blockTimestamp := uint64(time.Now().UnixNano())
		msgData := tickMessage[2].([]interface{})
		dataLength := len(msgData)
		deltas := make([]orderbook.Delta, dataLength, dataLength)

	dataIter: // Define a block to escape the orderbook parsing logic
		for i := 0; i < dataLength; i++ {
			var (
				_         = tickMessage[0].(float64)
				eventType uint8
				price     float64
				size      float64
			)

			tickData := msgData[i].([]interface{})

			switch tickData[0] {
			case "o": // Orderbook updates
				// Poloniex update format:
				//	[<MARKET_ID>, <MARKET_TICK>, [
				//		[<TICK_TYPE>, <BOOK_SIDE>, <PRICE>, <NEW_PRICE>],
				//		...
				//	]]
				price, _ = strconv.ParseFloat(tickData[2].(string), 32)
				size, _ = strconv.ParseFloat(tickData[3].(string), 32)

				switch tickData[1] { // Book side
				case 0: // Ask
					eventType = orderbook.IsUpdate | orderbook.IsAsk
				case 1: // Bid
					eventType = orderbook.IsUpdate | orderbook.IsBid
				}

			case "t": // Trade event
				price, _ = strconv.ParseFloat(tickData[3].(string), 32)
				size, _ = strconv.ParseFloat(tickData[4].(string), 32)

				switch tickData[2] { // Book side
				case 0: // Ask
					eventType = orderbook.IsTrade | orderbook.IsAsk
				case 1: // Bid
					eventType = orderbook.IsTrade | orderbook.IsBid
				}

			case "i": // Base orderbook event
				// The Poloniex orderbook tick is formatted as follows:
				//	[<MARKET_ID>, <MARKET_TICK>, {
				//		currencyPair: <MARKET>_<ASSET>,
				//		orderBook: [
				//			<ASK>{<ASK_PRICE>: <AMOUNT_ASSET>, ...},
				//			<BID>{<BID_PRICE>: <AMOUNT_ASSET>, ...}
				//		]
				//	}]

				// snapshotTick converts the orderbook data into a parsable format
				snapshotTick := tickData[1].(map[string]interface{})["orderBook"].([]interface{})
				snapshot := orderbook.Snapshot{
					Timestamp: blockTimestamp,
					StartSeq:  0,
					AskSide:   snapshotTick[0],
					BidSide:   snapshotTick[1],
				}
				fmt.Println(snapshot)
				continue dataIter
			}

			deltas[i] = orderbook.Delta{
				Timestamp: blockTimestamp,
				Event:     eventType,
				Price:     price,
				Size:      size,
			}
			fmt.Println(deltas[i])
			output <- deltas[i]
		}
	}
}
