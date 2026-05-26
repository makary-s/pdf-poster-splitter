package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// #region PDF fixtures

func writeSinglePagePDF(t *testing.T, path string, mediaBox [4]float64, rotate int) {
	t.Helper()

	var document bytes.Buffer
	offsets := make([]int, 0, 5)
	document.WriteString("%PDF-1.4\n")

	writeObject := func(objectNumber int, body string) {
		offsets = append(offsets, document.Len())
		fmt.Fprintf(&document, "%d 0 obj\n%s\nendobj\n", objectNumber, body)
	}

	pageRotation := ""
	if rotate != 0 {
		pageRotation = fmt.Sprintf(" /Rotate %d", rotate)
	}

	writeObject(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObject(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObject(3, fmt.Sprintf(
		"<< /Type /Page /Parent 2 0 R /MediaBox [%.2f %.2f %.2f %.2f]%s /Contents 4 0 R >>",
		mediaBox[0], mediaBox[1], mediaBox[2], mediaBox[3], pageRotation,
	))
	writeObject(4, "<< /Length 0 >>\nstream\n\nendstream")

	startXRef := document.Len()
	fmt.Fprintf(&document, "xref\n0 %d\n", len(offsets)+1)
	document.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		fmt.Fprintf(&document, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&document, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets)+1, startXRef)

	if err := os.WriteFile(path, document.Bytes(), 0o644); err != nil {
		t.Fatalf("write PDF fixture: %v", err)
	}
}

func writeLogoPNG(t *testing.T, path string) {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create PNG fixture: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode PNG fixture: %v", err)
	}
}

// #endregion

func TestGetFirstPageGeometryUsesVisualSizeForRotatedPage(t *testing.T) {
	t.Parallel()

	sourcePath := filepath.Join(t.TempDir(), "rotated.pdf")
	writeSinglePagePDF(t, sourcePath, [4]float64{0, 0, 100, 200}, 90)

	geometry, err := getFirstPageGeometry(sourcePath)
	if err != nil {
		t.Fatalf("getFirstPageGeometry: %v", err)
	}

	if geometry.rawWidthPt != 100 || geometry.rawHeightPt != 200 {
		t.Fatalf("unexpected raw size: got %.2fx%.2f", geometry.rawWidthPt, geometry.rawHeightPt)
	}
	if geometry.visualWidthPt != 200 || geometry.visualHeightPt != 100 {
		t.Fatalf("unexpected visual size: got %.2fx%.2f", geometry.visualWidthPt, geometry.visualHeightPt)
	}
	if geometry.rotationDeg != 90 {
		t.Fatalf("unexpected rotation: got %d", geometry.rotationDeg)
	}
}

func TestApplyWatermarkPreservesVisualSizeForRotatedPage(t *testing.T) {
	t.Parallel()

	for _, strategy := range []watermarkStrategyKey{
		watermarkStrategyVectorGrid,
		watermarkStrategyRasterGrid,
	} {
		t.Run(string(strategy), func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			sourcePath := filepath.Join(tempDir, "rotated.pdf")
			logoPath := filepath.Join(tempDir, "logo.png")
			outputPath := filepath.Join(tempDir, "rotated_watermarked.pdf")

			writeSinglePagePDF(t, sourcePath, [4]float64{0, 0, 100, 200}, 90)
			writeLogoPNG(t, logoPath)

			if err := applyWatermark(sourcePath, logoPath, 24, 0.3, string(strategy), outputPath); err != nil {
				t.Fatalf("applyWatermark: %v", err)
			}

			geometry, err := getFirstPageGeometry(outputPath)
			if err != nil {
				t.Fatalf("getFirstPageGeometry(output): %v", err)
			}

			if geometry.visualWidthPt != 200 || geometry.visualHeightPt != 100 {
				t.Fatalf("unexpected exported size: got %.2fx%.2f", geometry.visualWidthPt, geometry.visualHeightPt)
			}
		})
	}
}

func TestWatermarkStrategyFromStoredPrefDefaultsAndMigrates(t *testing.T) {
	t.Parallel()

	key, migrate := watermarkStrategyFromStoredPref("")
	if key != watermarkStrategyVectorGrid || migrate {
		t.Fatalf("unexpected empty pref handling: key=%q migrate=%v", key, migrate)
	}

	key, migrate = watermarkStrategyFromStoredPref("Один растр")
	if key != watermarkStrategyRasterGrid || !migrate {
		t.Fatalf("unexpected legacy pref handling: key=%q migrate=%v", key, migrate)
	}
}

func TestPlannedOutputPathsUsesOriginalNameForWatermarkedExport(t *testing.T) {
	t.Parallel()

	job := posterJob{
		pdfFileName: "/input/poster.pdf",
		targetPath:  "/output",
		posterName:  "poster",
	}
	options := appOptions{Logo: "/logo.png"}
	paths := plannedOutputPaths([]posterJob{job}, options, getTileSizeByName("A4"))

	expectedWatermarked := filepath.Join("/output", "poster.pdf")
	expectedTiled := filepath.Join("/output", "poster_A4.pdf")

	if !slices.Contains(paths, expectedWatermarked) {
		t.Fatalf("expected watermarked path %q, got %v", expectedWatermarked, paths)
	}
	if !slices.Contains(paths, expectedTiled) {
		t.Fatalf("expected tiled path %q, got %v", expectedTiled, paths)
	}
}

func TestCollectPDFFilesSkipsGeneratedArtifactsWhenInputEqualsOutput(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	writeSinglePagePDF(t, filepath.Join(tempDir, "poster.pdf"), [4]float64{0, 0, 100, 100}, 0)
	writeSinglePagePDF(t, filepath.Join(tempDir, "poster_A4.pdf"), [4]float64{0, 0, 100, 100}, 0)
	writeSinglePagePDF(t, filepath.Join(tempDir, "standalone_A4.pdf"), [4]float64{0, 0, 100, 100}, 0)

	jobs, err := collectPDFFiles(tempDir, tempDir)
	if err != nil {
		t.Fatalf("collectPDFFiles: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d: %+v", len(jobs), jobs)
	}

	names := []string{jobs[0].posterName, jobs[1].posterName}
	slices.Sort(names)
	if !slices.Equal(names, []string{"poster", "standalone_A4"}) {
		t.Fatalf("unexpected job names: %v", names)
	}
}

func TestCollectPDFFilesSkipsNestedOutputDirectory(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input")
	outputPath := filepath.Join(inputPath, "out")
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		t.Fatalf("mkdir output path: %v", err)
	}

	writeSinglePagePDF(t, filepath.Join(inputPath, "poster.pdf"), [4]float64{0, 0, 100, 100}, 0)
	writeSinglePagePDF(t, filepath.Join(outputPath, "poster.pdf"), [4]float64{0, 0, 100, 100}, 0)

	jobs, err := collectPDFFiles(inputPath, outputPath)
	if err != nil {
		t.Fatalf("collectPDFFiles: %v", err)
	}

	if len(jobs) != 1 || jobs[0].posterName != "poster" {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}
}
