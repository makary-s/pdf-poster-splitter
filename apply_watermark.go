package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	imagedraw "image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
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

func applyVectorWatermarkGrid(
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

// #region Raster watermark

func applyRasterWatermarkGrid(
	pdf *gopdf.GoPdf,
	logoPath string,
	logoWidthPt float64,
	logoAlpha float64,
	pageW float64,
	pageH float64,
) error {
	overlayPNG, err := buildRasterWatermarkOverlay(logoPath, logoWidthPt, logoAlpha, pageW, pageH)
	if err != nil {
		return err
	}

	imgHolder, err := gopdf.ImageHolderByBytes(overlayPNG)
	if err != nil {
		return err
	}

	return pdf.ImageByHolderWithOptions(imgHolder, gopdf.ImageOptions{
		X:    0,
		Y:    0,
		Rect: &gopdf.Rect{W: pageW, H: pageH},
	})
}

func buildRasterWatermarkOverlay(
	logoPath string,
	logoWidthPt float64,
	logoAlpha float64,
	pageW float64,
	pageH float64,
) ([]byte, error) {
	scale := getRasterWatermarkScale(pageW, pageH)
	pageWidthPx := scaledWatermarkPixels(pageW, scale)
	pageHeightPx := scaledWatermarkPixels(pageH, scale)
	logoWidthPx := scaledWatermarkPixels(logoWidthPt, scale)

	logoImage, err := loadWatermarkLogoImage(logoPath)
	if err != nil {
		return nil, err
	}

	logoRatio := float64(logoImage.Bounds().Dx()) / float64(logoImage.Bounds().Dy())
	logoHeightPx := scaledWatermarkPixels(float64(logoWidthPx)/logoRatio, 1)
	logoGapPx := scaledWatermarkPixels(logoGapPt, scale)

	scaledLogo := imaging.Resize(logoImage, logoWidthPx, logoHeightPx, imaging.Lanczos)
	alphaLogo := applyAlphaToImage(scaledLogo, logoAlpha)
	rotatedLogo := imaging.Rotate(alphaLogo, watermarkAngle, image.Transparent)

	overlay := image.NewNRGBA(image.Rect(0, 0, pageWidthPx, pageHeightPx))
	box := getWatermarkBounds(float64(pageWidthPx), float64(pageHeightPx), watermarkAngle)
	logoSpanPx := float64(logoWidthPx)/2 + float64(logoGapPx)
	colNum := int((box.w + logoSpanPx*2) / float64(logoWidthPx))
	rowNum := int(box.h / float64(logoHeightPx))

	for col := 0; col < colNum; col++ {
		for row := 0; row < rowNum; row++ {
			y := float64(row) * float64(logoHeightPx+logoGapPx)
			x := float64(col)*(float64(logoWidthPx)+float64(logoGapPx)) - logoSpanPx
			if row%2 == 1 {
				x += float64(logoWidthPx) / 2
			}

			centerX := x + float64(logoWidthPx)/2
			centerY := -y + float64(logoHeightPx)/2
			rotatedCenterX, rotatedCenterY := rotateScreenVector(centerX, centerY, watermarkAngle)
			drawImageCentered(
				overlay,
				rotatedLogo,
				box.x+rotatedCenterX,
				box.y+rotatedCenterY,
			)
		}
	}

	var pngBuffer bytes.Buffer
	if err := png.Encode(&pngBuffer, overlay); err != nil {
		return nil, err
	}

	return pngBuffer.Bytes(), nil
}

func getRasterWatermarkScale(pageW, pageH float64) float64 {
	longestSide := math.Max(pageW, pageH)
	if longestSide <= 0 {
		return 1
	}
	if longestSide <= maxRasterWatermarkSidePx {
		return 1
	}
	return maxRasterWatermarkSidePx / longestSide
}

func scaledWatermarkPixels(value float64, scale float64) int {
	return max(1, int(math.Round(value*scale)))
}

func loadWatermarkLogoImage(logoPath string) (image.Image, error) {
	file, err := os.Open(logoPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	logoImage, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	if logoImage.Bounds().Dy() == 0 {
		return nil, fmt.Errorf("invalid logo dimensions")
	}

	return logoImage, nil
}

func applyAlphaToImage(src image.Image, alpha float64) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, a := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			dst.SetNRGBA(x, y, colorNRGBAFromRGBA64(r, g, b, a, alpha))
		}
	}

	return dst
}

func colorNRGBAFromRGBA64(r, g, b, a uint32, alpha float64) color.NRGBA {
	scaledAlpha := uint8(math.Round((float64(a>>8) * alpha)))
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: scaledAlpha,
	}
}

func rotateScreenVector(x, y, angleDeg float64) (float64, float64) {
	radians := angleDeg * math.Pi / 180
	cosine := math.Cos(radians)
	sine := math.Sin(radians)
	return x*cosine + y*sine, -x*sine + y*cosine
}

func drawImageCentered(dst *image.NRGBA, src image.Image, centerX, centerY float64) {
	srcBounds := src.Bounds()
	drawRect := image.Rect(
		int(math.Round(centerX-float64(srcBounds.Dx())/2)),
		int(math.Round(centerY-float64(srcBounds.Dy())/2)),
		int(math.Round(centerX-float64(srcBounds.Dx())/2))+srcBounds.Dx(),
		int(math.Round(centerY-float64(srcBounds.Dy())/2))+srcBounds.Dy(),
	)
	imagedraw.Draw(dst, drawRect, src, srcBounds.Min, imagedraw.Over)
}

// #endregion

func applyWatermark(sourcePath, logoPath string, logoWidthPt, logoAlpha float64, strategy string, outputPath string) error {
	geometry, err := getFirstPageGeometry(sourcePath)
	if err != nil {
		return err
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: geometry.visualWidthPt, H: geometry.visualHeightPt}})
	pdf.AddPage()

	tplID := pdf.ImportPage(sourcePath, 1, "/MediaBox")
	pdf.UseImportedTemplate(tplID, 0, 0, geometry.visualWidthPt, geometry.visualHeightPt)

	switch strategyKey, ok := parseWatermarkStrategyKey(strategy); {
	case !ok:
		fallthrough
	case strategyKey == watermarkStrategyVectorGrid:
		if err := applyVectorWatermarkGrid(&pdf, logoPath, logoWidthPt, logoAlpha, geometry.visualWidthPt, geometry.visualHeightPt); err != nil {
			return err
		}
	case strategyKey == watermarkStrategyRasterGrid:
		if err := applyRasterWatermarkGrid(&pdf, logoPath, logoWidthPt, logoAlpha, geometry.visualWidthPt, geometry.visualHeightPt); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	pdf.WritePdf(outputPath)
	return nil
}
