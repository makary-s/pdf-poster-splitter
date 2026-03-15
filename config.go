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

	glueLineColorR uint8 = 255
	glueLineColorG uint8 = 0
	glueLineColorB uint8 = 0
	glueLineStyle        = "dashed"
	defaultGlueStrategy  = "trailing" // "trailing" | "all"

	watermarkAngle = 30.0

	initialWindowWidth  = 920.0
	initialWindowHeight = 600.0
)
