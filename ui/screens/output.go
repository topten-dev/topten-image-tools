package screens

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
	"github.com/topten-dev/topten-image-tools/core"
)

// OutputPicker lets the user choose the output directory before converting.
func OutputPicker(
	w fyne.Window,
	files []string,
	format core.Format,
	onConfirm func(outDir string),
	onBack func(),
) fyne.CanvasObject {
	var outputDir string

	title := widget.NewLabelWithStyle("Choose Output Folder", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	summary := widget.NewLabel(fmt.Sprintf(
		"%d image(s) will be converted to %s",
		len(files), formatLabel(format),
	))
	summary.Wrapping = fyne.TextWrapWord
	summary.Alignment = fyne.TextAlignCenter

	dirLabel := widget.NewLabel("No folder selected.")
	dirLabel.Wrapping = fyne.TextWrapWord

	convertBtn := widget.NewButton("Convert Now", nil)
	convertBtn.Importance = widget.HighImportance
	convertBtn.Disable()

	browseBtn := widget.NewButton("Browse…", func() {
		go func() {
			dir, err := zenity.SelectFile(
				zenity.Title("Select Output Folder"),
				zenity.Directory(),
			)
			if err != nil || dir == "" {
				return
			}
			fyne.Do(func() {
				outputDir = dir
				dirLabel.SetText("Output: " + outputDir)
				convertBtn.Enable()
			})
		}()
	})
	browseBtn.Importance = widget.MediumImportance

	// Default: same folder as first source file.
	if len(files) > 0 {
		outputDir = filepath.Dir(files[0])
		dirLabel.SetText("Output: " + outputDir + "  (same as source — change if needed)")
		convertBtn.Enable()
	}

	convertBtn.OnTapped = func() { onConfirm(outputDir) }
	backBtn := widget.NewButton("← Back", onBack)

	fileList := widget.NewLabel(buildFileList(files, format))
	fileList.Wrapping = fyne.TextWrapWord

	return container.NewPadded(container.NewBorder(
		container.NewVBox(title, summary, container.NewBorder(nil, nil, browseBtn, nil, dirLabel)),
		container.NewCenter(container.NewHBox(backBtn, convertBtn)),
		nil, nil,
		container.NewVScroll(fileList),
	))
}

func formatLabel(f core.Format) string {
	switch f {
	case core.FormatJPEG:
		return "JPEG (.jpg)"
	case core.FormatPNG:
		return "PNG (.png)"
	}
	return string(f)
}

func buildFileList(files []string, format core.Format) string {
	ext := "." + string(format)
	out := ""
	for _, f := range files {
		stem := filepath.Base(f[:len(f)-len(filepath.Ext(f))])
		out += fmt.Sprintf("  %s  →  %s%s\n", filepath.Base(f), stem, ext)
	}
	return out
}
