package orderbook

// Orderbook event types. This is used to encode two boolean values
// into one byte of data instead of two. This saves lots of space in the long run.
const (
	IsInsert uint8 = 1 << iota
	IsRemove uint8 = 1 << iota
	IsUpdate uint8 = 1 << iota
	IsTrade  uint8 = 1 << iota
	IsAsk    uint8 = 1 << iota
	IsBid    uint8 = 1 << iota
)

// Delta : Indicates a change in the orderbook state. This delta struct also works for
// options and other derivative products. The field `IsBid` is overloaded to equal a call
// option if it is true.
type Delta struct {
	Timestamp float64 `json:"ts"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Seq       uint32  `json:"seq"`
	IsTrade   bool    `json:"is_trade"`
	IsBid     bool    `json:"is_bid"`
}

// DeltaBatch : This is the representation of multiple Delta struct objects, mainly used for batch insertion
type DeltaBatch struct {
	Exchange string `json:"e"`
	Symbol   string `json:"s"`
	Deltas   *[]Delta
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

// ITickMessage : For JSON objects returned from sockets, save them in this interface if a specific type doesn't exist.
type ITickMessage []interface{}

// TODO LIST:
//
// Futures date generator/parser
// Options data capturing from Deribit

// -----------------------------------------------
// |    BEGIN ORDERBOOK STATE IMPLEMENTATION     |
// -----------------------------------------------

type BidEntry struct {
}

// Orderbook : Maintains the state of the orderbook for any given symbol
type Orderbook struct {
	TickSize float32
	LotSize  float32

	BidSide []float32
	AskSide []float32

	BestBid uint64
	BestAsk uint64

	InsertBid func(priceLevel uint64)
	InsertAsk func(priceLevel uint64)

	RemoveBid func(priceLevel uint64)
	RemoveAsk func(priceLevel uint64)
}
