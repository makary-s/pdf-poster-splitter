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

// drawCutLines draws dashed glue lines only where a neighboring tile exists.
// When both lines are present they terminate at their intersection instead of
// crossing (spec: "при пересечении линии обрываются").
func drawCutLines(pdf *gopdf.GoPdf, tileSize TileSize, glueMarginPt float64, hasRightGlue, hasBottomGlue bool) {
	if !hasRightGlue && !hasBottomGlue {
		return
	}

	pdf.SetStrokeColor(glueLineColorR, glueLineColorG, glueLineColorB)
	pdf.SetLineType(glueLineStyle)
	pdf.SetLineWidth(cutLineWeightPt)

	if hasRightGlue {
		x := tileSize.WidthPt - glueMarginPt
		if hasBottomGlue {
			pdf.Line(x, 0, x, tileSize.HeightPt-glueMarginPt)
		} else {
			pdf.Line(x, 0, x, tileSize.HeightPt)
		}
	}
	if hasBottomGlue {
		y := tileSize.HeightPt - glueMarginPt
		if hasRightGlue {
			pdf.Line(0, y, tileSize.WidthPt-glueMarginPt, y)
		} else {
			pdf.Line(0, y, tileSize.WidthPt, y)
		}
	}

	pdf.SetLineType("straight")
}

// drawTileLabel places the tile coordinate label.
// Orientation rules (spec): only right glueLine -> vertical inside right margin strip;
// bottom glueLine present or no glueLine at all -> horizontal in bottom-right corner.
func drawTileLabel(pdf *gopdf.GoPdf, tileSize TileSize, glueMarginPt float64, label string, hasRightGlue, hasBottomGlue bool) error {
	fontSize := glueMarginPt * 0.7
	if err := pdf.SetFont(appFontFamily, "", fontSize); err != nil {
		return err
	}
	pdf.SetTextColor(glueLineColorR, glueLineColorG, glueLineColorB)

	labelWidthEstimate := math.Max(20, float64(len(label))*fontSize*0.55)

	if hasRightGlue && !hasBottomGlue {
		// Vertical: rotate 90° so text runs bottom-to-top, anchored at bottom-right corner.
		gap := glueMarginPt * 0.15
		cx := tileSize.WidthPt - glueMarginPt/2
		cy := tileSize.HeightPt - gap - labelWidthEstimate/2
		pdf.Rotate(90, cx, cy)
		pdf.SetXY(cx-labelWidthEstimate/2, cy+fontSize*0.35)
		pdf.Text(label)
		pdf.RotateReset()
	} else {
		// Horizontal: bottom-right corner, inside bottom glue margin if present.
		gap := glueMarginPt * 0.15
		x := tileSize.WidthPt - gap - labelWidthEstimate
		y := tileSize.HeightPt - gap
		pdf.SetXY(x, y)
		pdf.Text(label)
	}

	return nil
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
			hasRightGlue := ix < tileColumns-1
			hasBottomGlue := iy < tileRows-1
			label := fmt.Sprintf("%d.%d", ix+1, iy+1)
			viewport := getTileViewport(ix, iy, tileSize, glueMarginPt)

			offsetX := -viewport.x1
			offsetY := -viewport.y1

			appendImportedPage(pdf, tplID, offsetX, offsetY, posterW, posterH)
			drawCutLines(pdf, tileSize, glueMarginPt, hasRightGlue, hasBottomGlue)
			if err := drawTileLabel(pdf, tileSize, glueMarginPt, label, hasRightGlue, hasBottomGlue); err != nil {
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
