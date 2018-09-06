package mathematics

import "math"

func europeanCall(S_t, t float64) float64 {
	return 0.00
}

func europeanPut(S_t, t float64) {
}

func _euroD1(S_t, t, K, sigma, r float64) float64 {
	return 1 / (sigma * math.Sqrt(t)) * (math.Log(S_t/K) + (r + ((math.Pow(sigma, 2) / 2) * t)))
}

func _euroD2(d1, sigma, t float64) float64 {
	return d1 - (sigma * sqrt(t))
}

func cdf()                 {}
func pdf()                 {}
func normal_distribution() {}
