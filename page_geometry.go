package main

import (
	"fmt"
	"math"

	pdfreader "rsc.io/pdf"
)

type pageGeometry struct {
	rawWidthPt     float64
	rawHeightPt    float64
	visualWidthPt  float64
	visualHeightPt float64
	rotationDeg    int
}

// #region First-page geometry

func getFirstPageGeometry(sourcePath string) (pageGeometry, error) {
	reader, err := pdfreader.Open(sourcePath)
	if err != nil {
		return pageGeometry{}, err
	}

	page := reader.Page(1)
	if page.V.IsNull() {
		return pageGeometry{}, fmt.Errorf("failed to read first page for %q", sourcePath)
	}

	box, err := getFirstPageBox(page)
	if err != nil {
		return pageGeometry{}, err
	}

	rawWidth := math.Abs(box.Index(2).Float64() - box.Index(0).Float64())
	rawHeight := math.Abs(box.Index(3).Float64() - box.Index(1).Float64())
	if rawWidth <= 0 || rawHeight <= 0 {
		return pageGeometry{}, fmt.Errorf("invalid first page size for %q", sourcePath)
	}

	rotation := normalizePageRotation(getInheritedPageValue(page.V, "Rotate").Int64())
	visualWidth := rawWidth
	visualHeight := rawHeight
	if rotation%180 != 0 {
		visualWidth, visualHeight = visualHeight, visualWidth
	}

	return pageGeometry{
		rawWidthPt:     rawWidth,
		rawHeightPt:    rawHeight,
		visualWidthPt:  visualWidth,
		visualHeightPt: visualHeight,
		rotationDeg:    rotation,
	}, nil
}

func getFirstPageSize(sourcePath string) (float64, float64, error) {
	geometry, err := getFirstPageGeometry(sourcePath)
	if err != nil {
		return 0, 0, err
	}
	return geometry.visualWidthPt, geometry.visualHeightPt, nil
}

func getFirstPageBox(page pdfreader.Page) (pdfreader.Value, error) {
	for _, boxName := range []string{"MediaBox", "CropBox"} {
		box := getInheritedPageValue(page.V, boxName)
		if box.Len() == 4 {
			return box, nil
		}
	}
	return pdfreader.Value{}, fmt.Errorf("no supported page box on first page")
}

func getInheritedPageValue(pageValue pdfreader.Value, key string) pdfreader.Value {
	for value := pageValue; !value.IsNull(); value = value.Key("Parent") {
		if inheritedValue := value.Key(key); !inheritedValue.IsNull() {
			return inheritedValue
		}
	}
	return pdfreader.Value{}
}

func normalizePageRotation(rawRotation int64) int {
	rotation := int(rawRotation % 360)
	if rotation < 0 {
		rotation += 360
	}
	return rotation
}

// #endregion
