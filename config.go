package main

const (
	appTitle = "Pdf poster splitter"
	appScale = "0.8"
)

const (
	defaultGlueMargin  = 10.0
	defaultLogoWidth   = 50.0
	defaultLogoOpacity = 30.0 // percent, 0–100

	cutLineWeightPt = 2.0
	logoGapPt       = 3.0 * pointsPerInch / cmPerInch

	glueLineColorR           uint8 = 255
	glueLineColorG           uint8 = 0
	glueLineColorB           uint8 = 0
	glueLineStyle                  = "dashed"
	defaultGlueStrategy            = "trailing" // "trailing" | "all" | "full"
	defaultWatermarkStrategy       = "vector_grid"

	watermarkAngle = 30.0

	// Raster watermark uses one transparent overlay image per page.
	// Capping the longest side keeps export memory bounded while still
	// producing a fast-opening PDF in viewers.
	maxRasterWatermarkSidePx = 4096

	initialWindowWidth  = 920.0
	initialWindowHeight = 600.0
)
