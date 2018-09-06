package orderbook

// TODO: finish filling this out so we can use it for the black-scholes equations
type Option struct {
	isCall bool `json:"is_call"`

	K  float32 `json:"k"`
	S0 float32 `json:"s0"`

	sigma float32 `json:"sigma"`
	r     float32 `json:"r"`

	expiration float64 `json:"exp_ts"`
	current_ts float64 `json:"curr_ts"`
}
