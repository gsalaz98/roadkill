package orderbook

// Orderbook event types. This is used to encode two boolean values
// into one byte of data instead of two. This saves lots of space in the long run.
const (
	IsBid uint8 = 1 << 5
	IsAsk uint8 = 1 << 4

	IsTrade  uint8 = 1 << 3
	IsUpdate uint8 = 1 << 2
	IsRemove uint8 = 1 << 1
	IsInsert uint8 = 1 << 0 // Some exchanges choose to transmit an insert event. Let's have it here for good measure.
)

// Delta : Same format as `tectonic.Tick` - should make our lives much easier in the long run
type Delta struct {
	Timestamp float64 `json:"ts"`
	Seq       uint64  `json:"seq"`
	IsTrade   bool    `json:"is_trade"`
	IsBid     bool    `json:"is_bid"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
}

// LegacyDelta : This stores orderbook tick deltas used to reconstruct the orderbook.
// It is constructed in a way to save as much space in storage per tick.
type LegacyDelta struct {
	Timestamp uint64  `json:"d"`
	Seq       uint64  `json:"s"`
	Event     uint8   `json:"e"`
	Price     float64 `json:"p"`
	Size      float64 `json:"z"`
}

// DeltaBatch : This is the representation of multiple Delta struct objects, mainly used for batch insertion or sending
type DeltaBatch struct {
	Exchange string `json:"e"`
	Symbol   string `json:"s"`
	Deltas   []Delta
}

// Snapshot : Here we store the orderbook data before we serialize it and
// send it to the ZMQ socket and Redis.
// Orderbook is keyed by type `uint8`, where `0` == bid, and `1` == ask.
type Snapshot struct {
	Symbol    string                        `json:"s"`
	Timestamp uint64                        `json:"t"`
	Orderbook map[uint8]map[float64]float64 `json:"o"` // Here, we will make `0` the bid side, and `1` the ask side. N.B.!! <<<<
}

// IOrderbookSnapshot : We must capture a snapshot of the orderbook
// when we first start capturing data. This type is to be used as a container
// for the ordrebook, then loaded into the predefined structure `OrderbookSnapshot`
type IOrderbookSnapshot []interface{}

// ITickMessage : For JSON objects returned from sockets, save them in this interface.
type ITickMessage []interface{}

// NormalizedSymbol : Lookup in the database the equivalent name for a given market and asset.
// e.g., In BitMEX, Bitcoin is represented as `XBT`, whereas in most other exchanges,
// it's abbreviated `BTC`. Changes like these must be accounted for
func NormalizedSymbol(exchange, symbol string) {} // TODO: work on this

// TODO LIST:
//
// Futures date generator/parser
// Options data capturing from Deribit
