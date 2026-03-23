package main

import (
	"fmt"
	"strings"
)

type TileSize struct {
	Name       string
	Display    string
	WidthPt    float64
	HeightPt   float64
	FileSuffix string
}

func newTileSize(name string, widthPt, heightPt float64) TileSize {
	return TileSize{
		Name:       name,
		Display:    fmt.Sprintf("%s (%sx%s)", name, prettyCm(ptToCm(widthPt)), prettyCm(ptToCm(heightPt))),
		WidthPt:    widthPt,
		HeightPt:   heightPt,
		FileSuffix: name,
	}
}

var tileSizes = []TileSize{
	newTileSize("A0", cmToPt(84.1), cmToPt(118.9)),
	newTileSize("A1", cmToPt(59.4), cmToPt(84.1)),
	newTileSize("A2", cmToPt(42.0), cmToPt(59.4)),
	newTileSize("A3", cmToPt(29.7), cmToPt(42.0)),
	newTileSize("A4", cmToPt(21.0), cmToPt(29.7)),
	newTileSize("A5", cmToPt(14.8), cmToPt(21.0)),
	newTileSize("A6", cmToPt(10.5), cmToPt(14.8)),
	newTileSize("B0", cmToPt(100.0), cmToPt(141.4)),
	newTileSize("B1", cmToPt(70.7), cmToPt(100.0)),
	newTileSize("B2", cmToPt(50.0), cmToPt(70.7)),
	newTileSize("B3", cmToPt(35.3), cmToPt(50.0)),
	newTileSize("B4", cmToPt(25.0), cmToPt(35.3)),
	newTileSize("B5", cmToPt(17.6), cmToPt(25.0)),
	newTileSize("B6", cmToPt(12.5), cmToPt(17.6)),
	newTileSize("LETTER", 612.0, 792.0),
	newTileSize("LEGAL", 612.0, 1008.0),
	newTileSize("TABLOID", 792.0, 1224.0),
	newTileSize("LEDGER", 1224.0, 792.0),
}

var tileSizeByDisplay = func() map[string]TileSize {
	res := make(map[string]TileSize, len(tileSizes))
	for _, size := range tileSizes {
		res[size.Display] = size
	}
	return res
}()

var tileSizeByName = func() map[string]TileSize {
	res := make(map[string]TileSize, len(tileSizes))
	for _, size := range tileSizes {
		res[size.Name] = size
	}
	return res
}()

func tileSizeDisplays() []string {
	items := make([]string, 0, len(tileSizes))
	for _, size := range tileSizes {
		items = append(items, size.Display)
	}
	return items
}

func defaultTileSize() TileSize {
	for _, size := range tileSizes {
		if size.Name == "A4" {
			return size
		}
	}
	return tileSizes[0]
}

func getTileSizeByName(key string) TileSize {
	key = strings.TrimSpace(key)
	if size, ok := tileSizeByName[key]; ok {
		return size
	}
	return defaultTileSize()
}

// tileSizeFromStoredPref resolves a preferences value: canonical Name, legacy Display, or default.
// The second return is true when prefs should be rewritten to the canonical name.
func tileSizeFromStoredPref(raw string) (TileSize, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultTileSize(), false
	}
	if size, ok := tileSizeByName[raw]; ok {
		return size, false
	}
	if size, ok := tileSizeByDisplay[raw]; ok {
		return size, true
	}
	return defaultTileSize(), true
}

// tileSizeFromSelectLabel maps the Fyne select option (Display) back to TileSize.
func tileSizeFromSelectLabel(label string) TileSize {
	for _, size := range tileSizes {
		if size.Display == label {
			return size
		}
	}
	return defaultTileSize()
}
