package main

import (
	"fmt"
	"math"
	"strings"
)

const (
	pointsPerInch = 72.0
	cmPerInch     = 2.54
)

func cmToPt(cm float64) float64 {
	return cm * pointsPerInch / cmPerInch
}

func ptToCm(pt float64) float64 {
	return pt * cmPerInch / pointsPerInch
}

func mmToPt(mm float64) float64 {
	return mm * pointsPerInch / (cmPerInch * 10)
}

func prettyCm(value float64) string {
	s := fmt.Sprintf("%.1f", value)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

func clampFloat64(value, minValue, maxValue float64) float64 {
	return math.Max(minValue, math.Min(maxValue, value))
}
