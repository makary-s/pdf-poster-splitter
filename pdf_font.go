package main

import (
	"github.com/signintech/gopdf"
	"golang.org/x/image/font/gofont/goregular"
)

const appFontFamily = "AppGoRegular"

func ensurePDFFont(pdf *gopdf.GoPdf) error {
	// We always register an embedded font to avoid system font dependencies.
	return pdf.AddTTFFontData(appFontFamily, goregular.TTF)
}
