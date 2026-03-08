package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	prefSourcePath  = "base_path"
	prefTargetPath  = "target_path"
	prefLogoPath    = "logo_path"
	prefLogoWidth   = "logo_width_mm"
	prefLogoOpacity = "logo_opacity"
	prefTitlePath   = "title_path"
	prefPaperSize   = "paper_size"
	prefGlueMargin  = "glue_margin_mm"
)

func runApp() error {
	os.Setenv("FYNE_SCALE", appScale)
	a := app.NewWithID("ru.kary.pdfpostersplitter")
	w := a.NewWindow(appTitle)
	w.Resize(fyne.NewSize(initialWindowWidth, initialWindowHeight))
	w.SetFixedSize(false) // Just ensuring it's not fixed, though it's default

	prefs := a.Preferences()
	runOnUI := func(fn func()) {
		fn()
	}

	// #region settings form
	sourceEntry := widget.NewEntry()
	sourceEntry.SetText(prefs.StringWithFallback(prefSourcePath, ""))

	targetEntry := widget.NewEntry()
	targetEntry.SetText(prefs.StringWithFallback(prefTargetPath, ""))

	logoEntry := widget.NewEntry()
	logoEntry.SetText(prefs.StringWithFallback(prefLogoPath, ""))
	logoEntry.OnChanged = func(value string) {
		prefs.SetString(prefLogoPath, value)
	}

	defaultLogoWidthStr := fmt.Sprintf("%.4g", defaultLogoWidth)
	logoWidthEntry := widget.NewEntry()
	logoWidthEntry.SetText(prefs.StringWithFallback(prefLogoWidth, defaultLogoWidthStr))

	defaultLogoOpacityStr := fmt.Sprintf("%.4g", defaultLogoOpacity)
	logoOpacityEntry := widget.NewEntry()
	logoOpacityEntry.SetText(prefs.StringWithFallback(prefLogoOpacity, defaultLogoOpacityStr))

	titleEntry := widget.NewEntry()
	titleEntry.SetText(prefs.StringWithFallback(prefTitlePath, ""))
	titleEntry.OnChanged = func(value string) {
		prefs.SetString(prefTitlePath, value)
	}

	paperOptions := tileSizeDisplays()
	defaultPaper := defaultTileSize().Display
	paperSelect := widget.NewSelect(paperOptions, func(selected string) {
		prefs.SetString(prefPaperSize, selected)
	})
	paperSelect.SetSelected(prefs.StringWithFallback(prefPaperSize, defaultPaper))
	if paperSelect.Selected == "" {
		paperSelect.SetSelected(defaultPaper)
	}

	defaultMarginStr := fmt.Sprintf("%.4g", defaultGlueMargin)
	glueMarginEntry := widget.NewEntry()
	glueMarginEntry.SetText(prefs.StringWithFallback(prefGlueMargin, defaultMarginStr))
	// #endregion

	// #region file dialogs
	selectFolder := func(onSelected func(path string)) {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			onSelected(uri.Path())
		}, w)
	}

	selectFile := func(filterExt []string, onSelected func(path string)) {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			path := reader.URI().Path()
			_ = reader.Close()
			onSelected(path)
		}, w)
		fileDialog.SetFilter(storage.NewExtensionFileFilter(filterExt))
		fileDialog.Show()
	}
	// #endregion

	// #region validation
	isRunning := false
	startButton := widget.NewButton("Начать", nil)

	var updateStartButton func()
	var resetButton *widget.Button

	sourceEntry.Validator = func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("обязательное поле")
		}
		return nil
	}
	sourceEntry.OnChanged = func(value string) {
		prefs.SetString(prefSourcePath, value)
		updateStartButton()
	}

	targetEntry.Validator = func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("обязательное поле")
		}
		return nil
	}
	targetEntry.OnChanged = func(value string) {
		prefs.SetString(prefTargetPath, value)
		updateStartButton()
	}

	glueMarginEntry.Validator = func(s string) error {
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil || v <= 0 {
			return errors.New("должно быть положительным числом")
		}
		return nil
	}
	glueMarginEntry.OnChanged = func(value string) {
		prefs.SetString(prefGlueMargin, value)
		updateStartButton()
	}

	logoWidthEntry.Validator = func(s string) error {
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil || v <= 0 {
			return errors.New("должно быть положительным числом")
		}
		return nil
	}
	logoWidthEntry.OnChanged = func(value string) {
		prefs.SetString(prefLogoWidth, value)
	}

	logoOpacityEntry.Validator = func(s string) error {
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil || v < 0 || v > 100 {
			return errors.New("должно быть числом от 0 до 100")
		}
		return nil
	}
	logoOpacityEntry.OnChanged = func(value string) {
		prefs.SetString(prefLogoOpacity, value)
	}

	updateStartButton = func() {
		if isRunning {
			// Disable reset button during processing to prevent configuration changes
			if resetButton != nil {
				resetButton.Disable()
			}
			return
		}
		// Re-enable reset button when processing is finished
		if resetButton != nil {
			resetButton.Enable()
		}
		if strings.TrimSpace(sourceEntry.Text) == "" || strings.TrimSpace(targetEntry.Text) == "" {
			startButton.Disable()
			return
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(glueMarginEntry.Text), 64)
		if err != nil || v <= 0 {
			startButton.Disable()
			return
		}
		startButton.Enable()
	}

	resetToDefaults := func() {
		if isRunning {
			return
		}
		sourceEntry.SetText("")
		targetEntry.SetText("")
		logoEntry.SetText("")
		logoWidthEntry.SetText(defaultLogoWidthStr)
		logoOpacityEntry.SetText(defaultLogoOpacityStr)
		titleEntry.SetText("")
		paperSelect.SetSelected(defaultPaper)
		glueMarginEntry.SetText(defaultMarginStr)

		// Explicitly update preferences for those that might not trigger OnChanged if value is same
		prefs.SetString(prefSourcePath, "")
		prefs.SetString(prefTargetPath, "")
		prefs.SetString(prefLogoPath, "")
		prefs.SetString(prefLogoWidth, defaultLogoWidthStr)
		prefs.SetString(prefLogoOpacity, defaultLogoOpacityStr)
		prefs.SetString(prefTitlePath, "")
		prefs.SetString(prefPaperSize, defaultPaper)
		prefs.SetString(prefGlueMargin, defaultMarginStr)

		sourceEntry.Validate()
		targetEntry.Validate()
		glueMarginEntry.Validate()
		logoWidthEntry.Validate()
		logoOpacityEntry.Validate()
		updateStartButton()
	}
	// #endregion

	// #region form helpers
	sectionHeading := func(text string) *widget.RichText {
		return widget.NewRichTextFromMarkdown("## " + text)
	}

	fieldRow := func(label string, input fyne.CanvasObject, action fyne.CanvasObject, hint string) fyne.CanvasObject {
		row := container.NewGridWithColumns(3, widget.NewLabel(label), input, action)
		if hint == "" {
			return row
		}
		hintText := widget.NewRichText(&widget.TextSegment{
			Style: widget.RichTextStyle{
				ColorName: theme.ColorNamePlaceHolder,
				SizeName:  theme.SizeNameCaptionText,
				Inline:    true,
			},
			Text: hint,
		})
		return container.NewVBox(row, hintText)
	}
	// #endregion

	pathsGroup := container.NewVBox(
		fieldRow("Из папки", sourceEntry, widget.NewButton("Найти", func() {
			selectFolder(sourceEntry.SetText)
		}), "PDF-плакаты из этой папки будут разбиты на листы"),
		fieldRow("В папку", targetEntry, widget.NewButton("Найти", func() {
			selectFolder(targetEntry.SetText)
		}), "Сюда сохранятся готовые файлы, структура подпапок сохраняется"),
	)

	tilesGroup := container.NewVBox(
		fieldRow("Формат бумаги", paperSelect, widget.NewLabel(""), "На листы этого формата будет нарезан каждый файл"),
		fieldRow("Отступ для склейки (мм)", glueMarginEntry, widget.NewLabel(""), "Перекрытие между листами для точного совмещения"),
	)

	watermarkGroup := container.NewVBox(
		fieldRow("Логотип", logoEntry, widget.NewButton("Найти", func() {
			selectFile([]string{".png"}, logoEntry.SetText)
		}), "PNG-картинка, накладывается на каждый плакат"),
		fieldRow("Ширина логотипа (мм)", logoWidthEntry, widget.NewLabel(""), "Высота вычисляется пропорционально"),
		fieldRow("Прозрачность логотипа (%)", logoOpacityEntry, widget.NewLabel(""), "0 — полностью прозрачный, 100 — полностью непрозрачный"),
	)

	titleGroup := container.NewVBox(
		fieldRow("Файл", titleEntry, widget.NewButton("Найти", func() {
			selectFile([]string{".pdf"}, titleEntry.SetText)
		}), "PDF-файл, добавляется первой страницей к каждому результату"),
	)

	resetButton = widget.NewButtonWithIcon("Сбросить к настройкам по-умолчанию", theme.ContentUndoIcon(), resetToDefaults)

	form := container.NewVBox(
		sectionHeading("Папки"),
		pathsGroup,
		widget.NewSeparator(),
		sectionHeading("Нарезка"),
		tilesGroup,
		widget.NewSeparator(),
		sectionHeading("Вотермарки"),
		watermarkGroup,
		widget.NewSeparator(),
		sectionHeading("Титульный лист"),
		titleGroup,
		widget.NewSeparator(),
		resetButton,
	)

	fileProgress := widget.NewProgressBar()
	fileProgressLabel := widget.NewLabel("0/0")
	tileProgress := widget.NewProgressBar()
	tileProgressLabel := widget.NewLabel("0/0")

	setProgress := func(progress *widget.ProgressBar, label *widget.Label, current, total int) {
		if total <= 0 {
			progress.SetValue(0)
			label.SetText("0/0")
			return
		}
		progress.SetValue(clampFloat64(float64(current)/float64(total), 0, 1))
		label.SetText(fmt.Sprintf("%d/%d", current, total))
	}

	progressPanel := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, nil, fileProgressLabel, fileProgress),
		container.NewBorder(nil, nil, nil, tileProgressLabel, tileProgress),
	)

	var cancel context.CancelFunc

	finishRun := func() {
		isRunning = false
		cancel = nil
		startButton.SetText("Начать")
		updateStartButton()
	}

	startButton.OnTapped = func() {
		if isRunning {
			if cancel != nil {
				cancel()
			}
			return
		}

		marginVal, _ := strconv.ParseFloat(strings.TrimSpace(glueMarginEntry.Text), 64)

		isRunning = true
		startButton.SetText("Отменить")
		setProgress(fileProgress, fileProgressLabel, 0, 0)
		setProgress(tileProgress, tileProgressLabel, 0, 0)

		ctx, cancelFn := context.WithCancel(context.Background())
		cancel = cancelFn
		logoWidthVal, err := strconv.ParseFloat(strings.TrimSpace(logoWidthEntry.Text), 64)
		if err != nil || logoWidthVal <= 0 {
			logoWidthVal = defaultLogoWidth
		}
		logoOpacityVal, err := strconv.ParseFloat(strings.TrimSpace(logoOpacityEntry.Text), 64)
		if err != nil || logoOpacityVal < 0 || logoOpacityVal > 100 {
			logoOpacityVal = defaultLogoOpacity
		}
		options := appOptions{
			InputPath:   sourceEntry.Text,
			OutputPath:  targetEntry.Text,
			Logo:        logoEntry.Text,
			LogoWidth:   logoWidthVal,
			LogoOpacity: logoOpacityVal,
			TitlePage:   titleEntry.Text,
			TileSize:    paperSelect.Selected,
			GlueMargin:  marginVal,
		}

		go func() {
			err := processFolder(ctx, options, processCallbacks{
				SetFileProgress: func(current, total int) {
					runOnUI(func() {
						setProgress(fileProgress, fileProgressLabel, current, total)
					})
				},
				SetTileProgress: func(current, total int) {
					runOnUI(func() {
						setProgress(tileProgress, tileProgressLabel, current, total)
					})
				},
				ConfirmOverwrite: func() bool {
					confirmed := make(chan bool, 1)
					runOnUI(func() {
						dialog.ShowConfirm(
							"Файлы уже существуют",
							"В папке назначения уже есть файлы, которые будут перезаписаны. Продолжить?",
							func(ok bool) { confirmed <- ok },
							w,
						)
					})
					return <-confirmed
				},
			})

			runOnUI(func() {
				defer finishRun()

				switch {
				case err == nil:
					dialog.ShowInformation("Внимание", "Задача завершена!", w)
				case errors.Is(err, ErrProcessingStopped) || errors.Is(err, context.Canceled):
					dialog.ShowInformation("Внимание", "Остановлено!", w)
				default:
					dialog.ShowError(err, w)
				}
			})
		}()
	}

	// #region layout assembly
	actionPanel := container.NewVBox(
		widget.NewSeparator(),
		progressPanel,
		startButton,
	)

	scrollableForm := container.NewVScroll(form)

	mainContent := container.NewBorder(nil, actionPanel, nil, nil, scrollableForm)

	w.SetContent(container.NewPadded(mainContent))
	w.ShowAndRun()
	return nil
	// #endregion
}
