package main

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/signintech/gopdf"
)

type watermarkBounds struct {
	x float64
	y float64
	w float64
	h float64
}

func getWatermarkLogoSize(logoPath string, logoWidth float64) (float64, float64, error) {
	file, err := os.Open(logoPath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	img, err := png.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	if img.Height == 0 {
		return 0, 0, fmt.Errorf("invalid logo dimensions")
	}

	ratio := float64(img.Width) / float64(img.Height)
	return logoWidth, logoWidth / ratio, nil
}

func getWatermarkBounds(pageW, pageH, rotation float64) watermarkBounds {
	r1 := rotation * math.Pi / 180
	r2 := (90 - rotation) * math.Pi / 180

	h1 := pageW * math.Cos(r2)
	h2 := pageH * math.Sin(r2)
	h := h1 + h2

	w1 := pageW * math.Sin(r2)
	w2 := pageH * math.Cos(r2)
	w := w1 + w2

	x := h1 * math.Sin(r1)
	// gopdf uses y-down coordinates: anchor must be below the bottom edge of the page,
	// mirroring the ReportLab y-up position where the anchor was below y=0.
	y := pageH + h1*math.Cos(r1)

	return watermarkBounds{x: x, y: y, w: w, h: h}
}

func applyWatermarkGrid(
	pdf *gopdf.GoPdf,
	logoPath string,
	logoWidthPt float64,
	logoAlpha float64,
	pageW float64,
	pageH float64,
) error {
	logoW, logoH, err := getWatermarkLogoSize(logoPath, logoWidthPt)
	if err != nil {
		return err
	}

	transparency, err := gopdf.NewTransparency(logoAlpha, "/Normal")
	if err != nil {
		return err
	}

	imgHolder, err := gopdf.ImageHolderByPath(logoPath)
	if err != nil {
		return err
	}

	box := getWatermarkBounds(pageW, pageH, watermarkAngle)
	logoSpan := logoW/2 + logoGapPt
	colNum := int((box.w + logoSpan*2) / logoW)
	rowNum := int(box.h / logoH)

	pdf.Rotate(watermarkAngle, box.x, box.y)
	for col := 0; col < colNum; col++ {
		for row := 0; row < rowNum; row++ {
			y := float64(row) * (logoH + logoGapPt)
			x := float64(col)*(logoW+logoGapPt) - logoSpan
			if row%2 == 1 {
				x += logoW / 2
			}

			if err := pdf.ImageByHolderWithOptions(imgHolder, gopdf.ImageOptions{
				X:            box.x + x,
				Y:            box.y - y,
				Rect:         &gopdf.Rect{W: logoW, H: logoH},
				Transparency: &transparency,
			}); err != nil {
				return err
			}
		}
	}
	pdf.RotateReset()
	return nil
}

func applyWatermark(sourcePath, logoPath string, logoWidthPt, logoAlpha float64, outputPath string) error {
	posterW, posterH, err := getFirstPageSize(sourcePath)
	if err != nil {
		return err
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: posterW, H: posterH}})
	pdf.AddPage()

	tplID := pdf.ImportPage(sourcePath, 1, "/MediaBox")
	pdf.UseImportedTemplate(tplID, 0, 0, posterW, posterH)

	if err := applyWatermarkGrid(&pdf, logoPath, logoWidthPt, logoAlpha, posterW, posterH); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	pdf.WritePdf(outputPath)
	return nil
}
