package screens

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/topten-dev/topten-image-tools/core"
)

// ConversionResult summarises a completed conversion run for the results screen.
type ConversionResult struct {
	Total      int
	Succeeded  int
	Errors     []string
	BytesSaved int64
}

// Progress renders an animated progress screen and drives the conversion job.
// onDone is called on the Fyne main goroutine when all work is finished.
func Progress(
	w fyne.Window,
	job core.Job,
	cancel chan struct{},
	onDone func(ConversionResult),
	onCancel func(),
) fyne.CanvasObject {

	title := widget.NewLabelWithStyle("Converting…", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	bar := widget.NewProgressBar()
	bar.Min = 0
	bar.Max = float64(len(job.SourceFiles))

	statusLabel := widget.NewLabel("Starting…")
	statusLabel.Wrapping = fyne.TextWrapWord

	cancelBtn := widget.NewButton("Cancel", onCancel)
	cancelBtn.Importance = widget.DangerImportance

	go func() {
		result := ConversionResult{Total: len(job.SourceFiles)}
		ch := core.Run(job, cancel)
		for p := range ch {
			p := p // capture
			fyne.Do(func() {
				bar.SetValue(float64(p.Current))
				if p.Error != nil {
					result.Errors = append(result.Errors, p.Error.Error())
					statusLabel.SetText("⚠ " + p.Error.Error())
				} else {
					statusLabel.SetText(fmt.Sprintf("(%d/%d) %s", p.Current, p.Total, p.FileName))
					result.Succeeded++
				}
				result.BytesSaved = p.BytesSaved
			})
		}
		// Channel is closed — all work done (including the empty-job case).
		fyne.Do(func() { onDone(result) })
	}()

	return container.NewPadded(
		container.NewVBox(
			title,
			bar,
			statusLabel,
			container.NewCenter(cancelBtn),
		),
	)
}
