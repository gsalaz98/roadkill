package orderbook

// GDAXOrderbook :
type GDAXOrderbook struct {
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
