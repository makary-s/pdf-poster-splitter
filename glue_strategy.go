package main

import "strings"

// glueStrategyKey is the canonical value stored in preferences and passed to splitIntoTiles.
type glueStrategyKey string

const (
	glueStrategyTrailing glueStrategyKey = "trailing"
	glueStrategyAll      glueStrategyKey = "all"
	glueStrategyFull     glueStrategyKey = "full"
)

var glueStrategyLabels = map[glueStrategyKey]string{
	glueStrategyTrailing: "Только справа и снизу",
	glueStrategyAll:      "Все стороны",
	glueStrategyFull:     "Все стороны (включая периметр)",
}

// Order of items in the settings select (Fyne Select has no separate value column).
var glueStrategiesInSelectOrder = []glueStrategyKey{
	glueStrategyTrailing,
	glueStrategyAll,
	glueStrategyFull,
}

// Older app versions stored Russian labels in preferences instead of keys.
var legacyGlueStrategyPref = map[string]glueStrategyKey{
	"Только справа и снизу":          glueStrategyTrailing,
	"Все стороны":                    glueStrategyAll,
	"Все внутренние стороны":         glueStrategyAll,
	"Все стороны (включая периметр)": glueStrategyFull,
}

func glueStrategySelectOptionStrings() []string {
	out := make([]string, len(glueStrategiesInSelectOrder))
	for i, k := range glueStrategiesInSelectOrder {
		out[i] = glueStrategyLabels[k]
	}
	return out
}

func parseGlueStrategyKey(s string) (glueStrategyKey, bool) {
	switch glueStrategyKey(s) {
	case glueStrategyTrailing, glueStrategyAll, glueStrategyFull:
		return glueStrategyKey(s), true
	default:
		return "", false
	}
}

func defaultGlueStrategyKey() glueStrategyKey {
	if k, ok := parseGlueStrategyKey(defaultGlueStrategy); ok {
		return k
	}
	return glueStrategyTrailing
}

// glueStrategyFromStoredPref interprets a preference string: canonical key, legacy label, or default.
// The second return is true when prefs should be rewritten to the canonical key.
func glueStrategyFromStoredPref(raw string) (glueStrategyKey, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultGlueStrategyKey(), false
	}
	if k, ok := parseGlueStrategyKey(raw); ok {
		return k, false
	}
	if k, ok := legacyGlueStrategyPref[raw]; ok {
		return k, true
	}
	return defaultGlueStrategyKey(), true
}

func glueStrategyKeyFromLabel(label string) glueStrategyKey {
	for _, k := range glueStrategiesInSelectOrder {
		if glueStrategyLabels[k] == label {
			return k
		}
	}
	return defaultGlueStrategyKey()
}
