package mathematics

func european_call(S_t, t float64) float64 {
	0.00
}

func european_put(S_t, t float64) {}

func _euro_d1(S_t, t, K, sigma, r float64) float64 {
	return (1 / (sigma * sqrt(t)) * (ln(S_t/K) + (r + ((sigma^2)/2)*t)))
}

func _euro_d2(d1, sigma, t float64) float64 {
	return d1 - (sigma * sqrt(t))
}

func cdf()                 {}
func normal_distribution() {}
