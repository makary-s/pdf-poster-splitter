package main

import "strings"

// watermarkStrategyKey is the canonical value stored in preferences and passed to watermark rendering.
type watermarkStrategyKey string

const (
	watermarkStrategyVectorGrid watermarkStrategyKey = "vector_grid"
	watermarkStrategyRasterGrid watermarkStrategyKey = "raster_grid"
)

var watermarkStrategyLabels = map[watermarkStrategyKey]string{
	watermarkStrategyVectorGrid: "Векторная сетка",
	watermarkStrategyRasterGrid: "Один растр",
}

// Order of items in the settings select (Fyne Select has no separate value column).
var watermarkStrategiesInSelectOrder = []watermarkStrategyKey{
	watermarkStrategyVectorGrid,
	watermarkStrategyRasterGrid,
}

// Older app versions had no watermark strategy setting, so empty value means default.
var legacyWatermarkStrategyPref = map[string]watermarkStrategyKey{
	"Векторная сетка": watermarkStrategyVectorGrid,
	"Один растр":      watermarkStrategyRasterGrid,
}

func watermarkStrategySelectOptionStrings() []string {
	out := make([]string, len(watermarkStrategiesInSelectOrder))
	for i, k := range watermarkStrategiesInSelectOrder {
		out[i] = watermarkStrategyLabels[k]
	}
	return out
}

func parseWatermarkStrategyKey(s string) (watermarkStrategyKey, bool) {
	switch watermarkStrategyKey(s) {
	case watermarkStrategyVectorGrid, watermarkStrategyRasterGrid:
		return watermarkStrategyKey(s), true
	default:
		return "", false
	}
}

func defaultWatermarkStrategyKey() watermarkStrategyKey {
	if k, ok := parseWatermarkStrategyKey(defaultWatermarkStrategy); ok {
		return k
	}
	return watermarkStrategyVectorGrid
}

// watermarkStrategyFromStoredPref interprets a preference string: canonical key, legacy label, or default.
// The second return is true when prefs should be rewritten to the canonical key.
func watermarkStrategyFromStoredPref(raw string) (watermarkStrategyKey, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultWatermarkStrategyKey(), false
	}
	if k, ok := parseWatermarkStrategyKey(raw); ok {
		return k, false
	}
	if k, ok := legacyWatermarkStrategyPref[raw]; ok {
		return k, true
	}
	return defaultWatermarkStrategyKey(), true
}

func watermarkStrategyKeyFromLabel(label string) watermarkStrategyKey {
	for _, k := range watermarkStrategiesInSelectOrder {
		if watermarkStrategyLabels[k] == label {
			return k
		}
	}
	return defaultWatermarkStrategyKey()
}
