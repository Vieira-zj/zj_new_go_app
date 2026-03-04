package utils

import "math"

func NewInt(x int) *int {
	return &x
}

func RoundWithPrecision(num float64, precision uint) float64 {
	base := math.Pow(10, float64(precision))
	return math.Round(num*base) / base
}

func FloorWithPrecision(num float64, precision uint) float64 {
	base := math.Pow(10, float64(precision))
	return math.Floor(num*base) / base
}
