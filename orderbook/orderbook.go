package orderbook

// Orderbook event types. This is used to encode two boolean values
// into one byte of data instead of two. This saves lots of space in the long run.
const (
	IsBidTrade       uint8 = 0
	IsBidUpdate      uint8 = 1
	IsBidRemove      uint8 = 2
	IsBidTradeRemove uint8 = 11 // For support purposes only

	IsAskTrade       uint8 = 3
	IsAskUpdate      uint8 = 4
	IsAskRemove      uint8 = 5
	IsAskTradeRemove uint8 = 10 // For support purposes only

	// DEPRECATED. Use events defined above instead
	IsBid    uint8 = 1 << 0
	IsAsk    uint8 = 1 << 1
	IsTrade  uint8 = 1 << 2
	IsUpdate uint8 = 1 << 3
)

// Delta : This stores orderbook tick deltas used to reconstruct the orderbook.
// It is constructed in a way to save as much space in storage per tick.
type Delta struct {
	TimeDelta uint64  `json:"d"`
	Seq       uint64  `json:"s"`
	Event     uint8   `json:"e"`
	Price     float64 `json:"p"`
	Size      float64 `json:"z"`
}

// Deltas : This is the representation of multiple Delta struct objects
type Deltas []Delta

// Snapshot : Here we store the orderbook data before we serialize it and
// send it to the ZMQ socket and Redis.
type Snapshot struct {
	Timestamp uint64      `json:"t"`
	StartSeq  uint64      `json:"s"`
	AskSide   interface{} `json:"a"` // TODO: MAKE THIS THE DEFAULT STYLE map[float64]float64
	BidSide   interface{} `json:"b"` // TODO: MAKE THIS THE DEFAULT STYLE map[float64]float64
}

// IOrderbookSnapshot : We must capture a snapshot of the orderbook
// when we first start capturing data. This type is to be used as a container
// for the ordrebook, then loaded into the predefined structure `OrderbookSnapshot`
type IOrderbookSnapshot []interface{}

// ITickMessage : For JSON objects returned from sockets, save them in this interface.
type ITickMessage []interface{}

// GetNormalizedMarketName : Lookup in the database the equivalent name for a given market and asset.
// e.g., In BitMEX, Bitcoin is represented as `XBT`, whereas in most other exchanges,
// it's abbreviated `BTC`. Changes like these must be accounted for
func GetNormalizedMarketName(exchange, market, asset string) string {
	// TODO: FINISH THIS FUNCTION
	return "NOTYETFINISHED"
}

// InsertSnapshot :
func (snapshot *Snapshot) InsertSnapshot() {}

// InsertTick : Insert a single tick to some database
func (delta *Delta) InsertTick() {}

// InsertTicks : Insert multiple ticks into a database
func InsertTicks(ticks []Delta) {}
