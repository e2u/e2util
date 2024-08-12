package e2math

import (
	"math"
)

func RoundToDecimal(value float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Round(value*shift) / shift
}
