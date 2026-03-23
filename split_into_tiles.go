package main

import (
	"fmt"
	"math"

	"github.com/phpdave11/gofpdi"
	"github.com/signintech/gopdf"
)

type tileViewport struct {
	x1 float64
	y1 float64
	x2 float64
	y2 float64
}

type tileProgressCallback func(currentTile int, totalTiles int) (shouldStop bool)

// Tolerance for floating-point comparison of page dimensions (~0.18 mm).
// PDF generators round A4 to 595.28 pt while cmToPt(21) gives 595.2756 pt;
// without tolerance that 0.004 pt gap triggers extra tile rows/columns.
const tileTolerance = 0.5

// getTileCount returns how many tiles fit along one axis.
// Spec algorithm: each tile after the first starts at (step = tileDim - margin) from the previous,
// so the last tile always covers to the end of the poster.
func getTileCount(posterDim, tileDim, margin float64) int {
	if posterDim <= tileDim+tileTolerance {
		return 1
	}
	step := tileDim - margin
	if step <= 0 {
		return 1
	}
	ratio := (posterDim - margin) / step
	n := math.Ceil(ratio)
	if n-ratio < 1e-9 {
		n = math.Floor(ratio)
	}
	return int(n) + 1
}

// getTileViewport returns the poster-coordinate region a tile at (ix, iy) should show.
// Origin is the top-left corner of the poster; axes grow right and down.
func getTileViewport(ix, iy int, tileSize TileSize, glueMarginPt float64) tileViewport {
	stepX := tileSize.WidthPt - glueMarginPt
	stepY := tileSize.HeightPt - glueMarginPt
	x1 := float64(ix) * stepX
	y1 := float64(iy) * stepY
	return tileViewport{x1: x1, y1: y1, x2: x1 + tileSize.WidthPt, y2: y1 + tileSize.HeightPt}
}

func getFirstPageSize(sourcePath string) (float64, float64, error) {
	importer := gofpdi.NewImporter()
	importer.SetSourceFile(sourcePath)
	sizes := importer.GetPageSizes()
	page, ok := sizes[1]
	if !ok {
		return 0, 0, fmt.Errorf("failed to read first page size for %q", sourcePath)
	}

	for _, boxName := range []string{"/MediaBox", "/CropBox", "MediaBox", "CropBox"} {
		box, ok := page[boxName]
		if !ok {
			continue
		}

		width, wOK := box["w"]
		height, hOK := box["h"]
		if wOK && hOK && width > 0 && height > 0 {
			return width, height, nil
		}
	}

	return 0, 0, fmt.Errorf("no supported page box for %q", sourcePath)
}

// glueLineSides returns which sides of the tile page get a glue line for the given strategy.
// trailing: only right and bottom toward the next tile. all: every side that borders another tile.
// full: all four sides of every tile, including the outer perimeter of the poster.
func glueLineSides(strategy string, ix, iy, tileColumns, tileRows int) (hasLeft, hasTop, hasRight, hasBottom bool) {
	switch strategy {
	case "full":
		return true, true, true, true
	case "all":
		return ix > 0, iy > 0, ix < tileColumns-1, iy < tileRows-1
	default: // "trailing" and unknown
		return false, false, ix < tileColumns-1, iy < tileRows-1
	}
}

// drawCutLines draws dashed glue lines on the selected sides of the tile page.
// Lines terminate at their intersections instead of crossing — corner areas with
// double overlap are left empty (spec: "угловые области остаются без линий").
func drawCutLines(pdf *gopdf.GoPdf, tileSize TileSize, glueMarginPt float64, hasLeft, hasTop, hasRight, hasBottom bool) {
	if !hasLeft && !hasTop && !hasRight && !hasBottom {
		return
	}

	W := tileSize.WidthPt
	H := tileSize.HeightPt
	m := glueMarginPt

	// Vertical lines run from topClip to bottomClip; horizontal from leftClip to rightClip.
	topClip := 0.0
	if hasTop {
		topClip = m
	}
	bottomClip := H
	if hasBottom {
		bottomClip = H - m
	}
	leftClip := 0.0
	if hasLeft {
		leftClip = m
	}
	rightClip := W
	if hasRight {
		rightClip = W - m
	}

	pdf.SetStrokeColor(glueLineColorR, glueLineColorG, glueLineColorB)
	pdf.SetLineType(glueLineStyle)
	pdf.SetLineWidth(cutLineWeightPt)

	if hasLeft {
		pdf.Line(m, topClip, m, bottomClip)
	}
	if hasRight {
		pdf.Line(W-m, topClip, W-m, bottomClip)
	}
	if hasTop {
		pdf.Line(leftClip, m, rightClip, m)
	}
	if hasBottom {
		pdf.Line(leftClip, H-m, rightClip, H-m)
	}

	pdf.SetLineType("straight")
}

// drawTileLabel draws the tile grid label (col.row) at the bottom-right of the tile page.
// Insets from outer edges (SPEC @Flow:SplitIntoTiles): bottom glueLine.margin + 0.3× font size;
// right glueLine.margin + 0.5× font size.
func drawTileLabel(pdf *gopdf.GoPdf, tileSize TileSize, glueMarginPt float64, label string) error {
	fontSize := glueMarginPt * 0.7
	if err := pdf.SetFont(appFontFamily, "", fontSize); err != nil {
		return err
	}
	pdf.SetTextColor(glueLineColorR, glueLineColorG, glueLineColorB)

	textWidth, err := pdf.MeasureTextWidth(label)
	if err != nil {
		return err
	}

	W := tileSize.WidthPt
	H := tileSize.HeightPt
	insetBottom := glueMarginPt + 0.3*fontSize
	insetRight := glueMarginPt + 0.5*fontSize
	// gopdf's y is the text baseline; approximate descender so the ink bottom sits at H - insetBottom.
	descentPt := fontSize * 0.22

	x := W - insetRight - textWidth
	y := H - insetBottom - descentPt
	pdf.SetXY(x, y)
	return pdf.Text(label)
}

func appendImportedPage(
	pdf *gopdf.GoPdf,
	templateID int,
	offsetX float64,
	offsetY float64,
	templateW float64,
	templateH float64,
) {
	pdf.AddPage()
	pdf.UseImportedTemplate(templateID, offsetX, offsetY, templateW, templateH)
}

func splitIntoTiles(
	pdf *gopdf.GoPdf,
	sourcePath string,
	tileSize TileSize,
	glueMarginPt float64,
	strategy string,
	onTile tileProgressCallback,
) error {
	if err := ensurePDFFont(pdf); err != nil {
		return err
	}

	posterW, posterH, err := getFirstPageSize(sourcePath)
	if err != nil {
		return err
	}

	tplID := pdf.ImportPage(sourcePath, 1, "/MediaBox")
	tileColumns := getTileCount(posterW, tileSize.WidthPt, glueMarginPt)
	tileRows := getTileCount(posterH, tileSize.HeightPt, glueMarginPt)
	total := tileColumns * tileRows
	current := 0

	for iy := 0; iy < tileRows; iy++ {
		for ix := 0; ix < tileColumns; ix++ {
			hasLeft, hasTop, hasRight, hasBottom := glueLineSides(strategy, ix, iy, tileColumns, tileRows)
			label := fmt.Sprintf("%d.%d", ix+1, iy+1)
			viewport := getTileViewport(ix, iy, tileSize, glueMarginPt)

			offsetX := -viewport.x1
			offsetY := -viewport.y1

			appendImportedPage(pdf, tplID, offsetX, offsetY, posterW, posterH)
			drawCutLines(pdf, tileSize, glueMarginPt, hasLeft, hasTop, hasRight, hasBottom)
			if err := drawTileLabel(pdf, tileSize, glueMarginPt, label); err != nil {
				return err
			}

			current++
			if onTile != nil && onTile(current, total) {
				return nil
			}
		}
	}

	return nil
}
