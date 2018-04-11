package orderbook

// Orderbook event bit flags
const (
	IsBid    uint8 = 1 << 0
	IsAsk    uint8 = 1 << 1
	IsTrade  uint8 = 1 << 2
	IsUpdate uint8 = 1 << 3
)

// Delta : This stores orderbook tick deltas used to reconstruct the orderbook.
// It is constructed in a way to save as much space in storage per tick.
type Delta struct {
	Timestamp uint64
	Tick      uint32
	Event     uint8
	Price     float32
	Size      float32
}

// Snapshot : Here we store the orderbook data before we serialize it and
// send it to the ZMQ socket and Redis.
type Snapshot struct {
	Timestamp uint32
	StartSeq  uint64
	AskSide   interface{} // TODO: MAKE THIS THE DEFAULT STYLE map[float64]float64
	BidSide   interface{} // TODO: MAKE THIS THE DEFAULT STYLE map[float64]float64
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
func InsertSnapshot(snapshot Snapshot) {

}

// InsertTick : Inserts ticks as a batch data
func InsertTick(tick []Delta) {

}
