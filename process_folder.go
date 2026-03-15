package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/signintech/gopdf"
)

var ErrProcessingStopped = errors.New("processing stopped")

type appOptions struct {
	InputPath    string
	OutputPath   string
	Logo         string
	LogoWidth    float64
	LogoOpacity  float64 // percent, 0–100
	TitlePage    string
	TileSize     string
	GlueMargin   float64
	GlueStrategy string // "trailing" | "all"
}

type processCallbacks struct {
	SetFileProgress   func(current int, total int)
	SetTileProgress   func(current int, total int)
	ConfirmOverwrite  func() bool // called when planned outputs already exist; return false to abort
}

type posterJob struct {
	pdfFileName string
	targetPath  string
	posterName  string
}

func normalizeOptions(options appOptions) appOptions {
	normalizePath := func(path string) string {
		path = strings.TrimSpace(path)
		if path == "" {
			return ""
		}
		return filepath.Clean(path)
	}

	options.InputPath = normalizePath(options.InputPath)
	options.OutputPath = normalizePath(options.OutputPath)
	options.Logo = normalizePath(options.Logo)
	options.TitlePage = normalizePath(options.TitlePage)
	return options
}

func validatePaths(options appOptions) error {
	if info, err := os.Stat(options.InputPath); err != nil || !info.IsDir() {
		return fmt.Errorf("директория \"Из папки\" не существует: %s", options.InputPath)
	}
	if info, err := os.Stat(options.OutputPath); err != nil || !info.IsDir() {
		return fmt.Errorf("директория \"В папку\" не существует: %s", options.OutputPath)
	}
	if options.Logo != "" {
		info, err := os.Stat(options.Logo)
		if err != nil {
			return fmt.Errorf("файл логотипа не существует: %s", options.Logo)
		}
		if info.IsDir() {
			return fmt.Errorf("логотип должен быть файлом: %s", options.Logo)
		}
	}
	if options.TitlePage != "" {
		info, err := os.Stat(options.TitlePage)
		if err != nil {
			return fmt.Errorf("файл титульного листа не существует: %s", options.TitlePage)
		}
		if info.IsDir() {
			return fmt.Errorf("титульный лист должен быть файлом: %s", options.TitlePage)
		}
	}
	return nil
}

func collectPDFFiles(inputPath, outputPath string) ([]posterJob, error) {
	jobs := make([]posterJob, 0, 32)
	err := filepath.WalkDir(inputPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".pdf" {
			return nil
		}

		relDir, err := filepath.Rel(inputPath, filepath.Dir(path))
		if err != nil {
			return err
		}

		targetPath := filepath.Join(outputPath, relDir)
		posterName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		jobs = append(jobs, posterJob{
			pdfFileName: path,
			targetPath:  targetPath,
			posterName:  posterName,
		})
		return nil
	})
	return jobs, err
}

func prependTitlePage(pdf *gopdf.GoPdf, titlePath string) error {
	titleW, titleH, err := getFirstPageSize(titlePath)
	if err != nil {
		return err
	}

	opt := gopdf.PageOption{
		PageSize: &gopdf.Rect{W: titleW, H: titleH},
	}
	pdf.AddPageWithOption(opt)
	tplID := pdf.ImportPage(titlePath, 1, "/MediaBox")
	pdf.UseImportedTemplate(tplID, 0, 0, titleW, titleH)
	return nil
}

func createTilesPDF(
	posterSourcePath string,
	titlePage string,
	tileSize TileSize,
	glueMarginPt float64,
	glueStrategy string,
	outputPath string,
	ctx context.Context,
	callbacks processCallbacks,
) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: tileSize.WidthPt, H: tileSize.HeightPt}})

	if titlePage != "" {
		if err := prependTitlePage(&pdf, titlePage); err != nil {
			return err
		}
	}

	err := splitIntoTiles(&pdf, posterSourcePath, tileSize, glueMarginPt, glueStrategy, func(currentTile, totalTiles int) bool {
		if callbacks.SetTileProgress != nil {
			callbacks.SetTileProgress(currentTile, totalTiles)
		}
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	})
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ErrProcessingStopped
	default:
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	pdf.WritePdf(outputPath)
	return nil
}

func processFile(
	ctx context.Context,
	job posterJob,
	options appOptions,
	tileSize TileSize,
	callbacks processCallbacks,
) error {
	posterSourcePath := job.pdfFileName

	if options.Logo != "" {
		logoWidthPt := mmToPt(options.LogoWidth)
		logoAlpha := clampFloat64(options.LogoOpacity/100.0, 0, 1)
		watermarkedPath := filepath.Join(job.targetPath, job.posterName+".pdf")
		if err := applyWatermark(job.pdfFileName, options.Logo, logoWidthPt, logoAlpha, watermarkedPath); err != nil {
			return err
		}
		posterSourcePath = watermarkedPath
	}

	glueMarginPt := mmToPt(options.GlueMargin)
	outputPath := filepath.Join(job.targetPath, fmt.Sprintf("%s_%s.pdf", job.posterName, tileSize.FileSuffix))
	return createTilesPDF(posterSourcePath, options.TitlePage, tileSize, glueMarginPt, options.GlueStrategy, outputPath, ctx, callbacks)
}

func plannedOutputPaths(jobs []posterJob, options appOptions, tileSize TileSize) []string {
	paths := make([]string, 0, len(jobs)*2)
	for _, job := range jobs {
		if options.Logo != "" {
			paths = append(paths, filepath.Join(job.targetPath, job.posterName+".pdf"))
		}
		paths = append(paths, filepath.Join(job.targetPath, fmt.Sprintf("%s_%s.pdf", job.posterName, tileSize.FileSuffix)))
	}
	return paths
}

func anyPlannedOutputExists(paths []string) bool {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func processFolder(
	ctx context.Context,
	options appOptions,
	callbacks processCallbacks,
) error {
	options = normalizeOptions(options)
	if err := validatePaths(options); err != nil {
		return err
	}

	jobs, err := collectPDFFiles(options.InputPath, options.OutputPath)
	if err != nil {
		return err
	}

	tileSize := getTileSizeByDisplay(options.TileSize)

	if anyPlannedOutputExists(plannedOutputPaths(jobs, options, tileSize)) {
		if callbacks.ConfirmOverwrite == nil || !callbacks.ConfirmOverwrite() {
			return ErrProcessingStopped
		}
	}

	if callbacks.SetFileProgress != nil {
		callbacks.SetFileProgress(0, len(jobs))
	}
	if callbacks.SetTileProgress != nil {
		callbacks.SetTileProgress(0, 0)
	}

	for idx, job := range jobs {
		select {
		case <-ctx.Done():
			return ErrProcessingStopped
		default:
		}

		if err := processFile(ctx, job, options, tileSize, callbacks); err != nil {
			return err
		}
		if callbacks.SetFileProgress != nil {
			callbacks.SetFileProgress(idx+1, len(jobs))
		}
	}

	return nil
}
