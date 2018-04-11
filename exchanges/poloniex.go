package poloniex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

// AssetTable : enum with various helper parameters useful for identifying the market-asset pair
var AssetTable = make(map[float64]string)

func init() {
	var jsonData interface{}

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
		AssetTable[assetID] = assetPair
	}
}

// CreateConnection : Creates a websocket connection and returns the connection object
func CreateConnection(headers http.Header) *websocket.Conn {
	const ConnURL string = "wss://api2.poloniex.com"
	var c websocket.Dialer

	conn, _, err := c.Dial(ConnURL, headers)

	if err != nil {
		fmt.Println("Error in connection: ", err)
	}

	return conn
}

// SendMessage : Sends a websocket message to the socket object `c`.
func SendMessage(c *websocket.Conn, messages []map[string]string) {
	for _, message := range messages {
		c.WriteJSON(message)
	}
}

// ReceiveMessageLoop : This runs infinitely until the connection is closed by the server.
// For concurrent use only.
func ReceiveMessageLoop(c *websocket.Conn) {
	// TODO: Get poloniex current tick count by asking the database itself
	var tickMessage orderbook.ITickMessage

	for {
		c.ReadJSON(&tickMessage)
		msgLength := len(tickMessage)

		if msgLength < 3 {
			continue
		}

		msgData := tickMessage[2].([]interface{})
		dataLength := len(msgData)
		deltas := make([]orderbook.Delta, dataLength, dataLength)

		st := time.Now()
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
					Timestamp: uint32(time.Now().UnixNano() / 1000),
					StartSeq:  0,
					AskSide:   snapshotTick[0],
					BidSide:   snapshotTick[1],
				}
				fmt.Println(snapshot)
				continue dataIter
			}

			deltas[i] = orderbook.Delta{
				Timestamp: uint64(time.Now().UnixNano() / 1000),
				Tick:      0,
				Event:     eventType,
				Price:     float32(price),
				Size:      float32(size),
			}
			//fmt.Println(pairCode, AssetTable[pairCode], deltas[i])
		}
		en := time.Now().Sub(st)
		fmt.Println("Time elapsed: ", en)
	}
}

// NormalizedMarketName : Using standard market and asset
// names, returns an exchange specific asset pair.
// TODO: Work on making a database lookup table
func NormalizedMarketName(market, asset string) string {
	var marketBuffer bytes.Buffer

	marketBuffer.WriteString(market)
	marketBuffer.WriteString("_")
	marketBuffer.WriteString(asset)

	return marketBuffer.String()
}
