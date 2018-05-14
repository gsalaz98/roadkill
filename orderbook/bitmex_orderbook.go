package orderbook

// IBitMexTick : Every tick passed will have this format.
type IBitMexTick struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	Data   []struct {
		Size   int64   `json:"size"`
		Price  float32 `json:"price"`
		ID     int32   `json:"id"`
		Symbol string  `json:"symbol"`
		Side   string  `json:"side"`
	} `json:"data"`
}
