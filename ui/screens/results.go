package screens

import (
	"fmt"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Results renders a summary screen after a conversion run finishes.
func Results(
	w fyne.Window,
	result ConversionResult,
	outputDir string,
	onConvertMore func(),
) fyne.CanvasObject {

	emoji := "✅"
	if len(result.Errors) > 0 {
		emoji = "⚠️"
	}

	headline := widget.NewLabelWithStyle(
		fmt.Sprintf("%s Conversion Complete", emoji),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	stats := fmt.Sprintf(
		"%d of %d image(s) converted successfully.",
		result.Succeeded, result.Total,
	)
	if result.BytesSaved > 0 {
		stats += "\n" + fmt.Sprintf("Space saved: %s", humanBytes(result.BytesSaved))
	} else if result.BytesSaved < 0 {
		stats += "\n" + fmt.Sprintf("Output is %s larger than input (this is normal when converting from JPG to PNG).",
			humanBytes(-result.BytesSaved))
	}
	statsLabel := widget.NewLabel(stats)
	statsLabel.Alignment = fyne.TextAlignCenter
	statsLabel.Wrapping = fyne.TextWrapWord

	openBtn := widget.NewButton("Open Output Folder", func() {
		openFolder(outputDir)
	})
	openBtn.Importance = widget.HighImportance

	convertMoreBtn := widget.NewButton("Convert More Images", onConvertMore)
	convertMoreBtn.Importance = widget.MediumImportance

	buttonRow := container.NewCenter(container.NewHBox(convertMoreBtn, openBtn))

	var errBox fyne.CanvasObject
	if len(result.Errors) > 0 {
		errText := "Errors:\n"
		for _, e := range result.Errors {
			errText += "  • " + e + "\n"
		}
		errLabel := widget.NewLabel(errText)
		errLabel.Wrapping = fyne.TextWrapWord
		errBox = container.NewVScroll(errLabel)
	} else {
		errBox = widget.NewLabel("")
	}

	return container.NewPadded(container.NewCenter(
		container.NewVBox(
			headline,
			statsLabel,
			buttonRow,
			errBox,
		),
	))
}

// humanBytes formats a byte count into a human-readable string.
func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// openFolder attempts to reveal the given path in the OS file manager.
func openFolder(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default: // linux / bsd
		cmd = exec.Command("xdg-open", path)
	}
	_ = cmd.Start()
}
