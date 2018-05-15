package orderbook

// FastGDAXOrderbook : This type is only to be used for the "full" channel. We will model the orderbook events as they pass them through, and attempt to
// calculate the book state faster than GDAX themselves can. However, that may prove to be *actually* impossible due to network latencies and such.
type FastGDAXOrderbook struct {
	Type      string
	Time      string
	ProductID string
	Sequence  uint64
	OrderID   string
	Price     string
	Side      string

	// Market order specific elements
	OrderType string
	Funds     string

	// Orderbook update elements
	RemainingSize string

	// Removal update elements

	// Done event
	Reason string
}

// SlowGDAXOrderbookUpdates : For channel "level2"
type SlowGDAXOrderbookUpdates struct {
	Type      string     `json:"type"`       // The two strings that are sent here are "snapshot" and "l2update"
	ProductID string     `json:"product_id"` // Alias for the asset-pair being traded. GDAX Follows the format `<ASSET>-<MARKET>`
	Changes   [][]string `json:"changes"`    // Data is sent as follows: [["buy", "<price>", "<size>"], ...] - Removals have size=0
}

// SlowGDAXMatches : Otherwise known as the "ticker" channel, this tracks any trades made on the GDAX exchange
type SlowGDAXMatches struct {
	Type      string `json:"type"`       // Channel name
	TradeID   uint64 `json:"trade_id"`   // Trade seq count
	Sequence  uint64 `json:"sequence"`   // Orderbook seq count
	Time      string `json:"time"`       // timestamp in ISO8601 format
	ProductID string `json:"product_id"` // asset-pair
	Price     string `json:"price"`      // Order @price
	Side      string `json:"side"`       // Either "buy" or "sell"
	LastSize  string `json:"last_size"`  // Order size
	BestBid   string `json:"best_bid"`   // Best bid when trade occured
	BestAsk   string `json:"best_ask"`   // Best ask when trade occured
}
