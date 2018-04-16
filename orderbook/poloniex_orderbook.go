package orderbook

// IPoloniexOrderbookSnapshot implements a custom snapshot type for performance reasons. Same as OrderbookDelta; should be converted to OrderbookSnapshot
// in a goroutine to prevent performance degredation
const (
	PoloniexBid = 1
	PoloniexAsk = 0
)

type IPoloniexOrderbookSnapshot struct {
	CurrencyPair string              `json:"currencyPair"`
	Orderbook    []map[string]string `json:"orderBook"`
}

// IPoloniexDelta : Custom Delta type for the efficient parsing of the poloniex orderbook data
type IPoloniexDelta struct {
	MarketID uint16
	Tick     uint64
	Data     []interface{}
}
