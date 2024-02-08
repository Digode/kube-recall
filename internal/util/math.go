package util

import "math"

func Round(val float64, precision int) float64 {
	return math.Round(val*(math.Pow10(precision))) / math.Pow10(precision)
}

func RoundUpToBase10(num float64) float64 {
	return math.Ceil(num/10) * 10
}

func RoundUpToBase50(num float64) float64 {
	return math.Ceil(num/50) * 50
}

func RoundUpToBase100(num float64) float64 {
	return math.Ceil(num/100) * 100
}
