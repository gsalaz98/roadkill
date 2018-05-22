package orderbook

// Here, let's have all the fee schedules for various exchanges

// FeeTier : We will store a list of these inside a FeeSchedule structure
type FeeTier struct {
	FeePercentage  float32
	RequiredVolume float32
	LessThan       bool
}

// FeeSchedule : Describes our current fee schedule on a certain exchange
type FeeSchedule struct {
	Tiers           []FeeTier
	VolumeUntilNext float32
	CurrentFees     float64
}
