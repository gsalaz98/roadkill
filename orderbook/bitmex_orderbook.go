package orderbook

// IBitMexTick : Every tick passed will have this format.
type IBitMexTick struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	Data   []struct {
		Symbol string `json:"symbol"`
		ID     int64  `json:"id"`
		Side   string `json:"side"`
		Size   int    `json:"size"`
		Price  int    `json:"price"`
	} `json:"data"`
}
