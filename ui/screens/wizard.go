package screens

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/topten-dev/topten-image-tools/core"
)

// Wizard renders a simple format-selection screen with three large option cards.
// anyAlpha is true when at least one source file already has an alpha channel.
func Wizard(
	w fyne.Window,
	anyAlpha bool,
	onConfirm func(core.Format),
	onBack func(),
) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Choose an output format", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	backBtn := widget.NewButton("← Back", onBack)

	type option struct {
		emoji  string
		label  string
		desc   string
		format core.Format
	}

	jpegDesc := "Best for photos, landscapes, product shots, and any image with lots of colours and gradients. Produces smaller files."
	if anyAlpha {
		jpegDesc += "\n⚠️ One or more of your files has a transparent background. If you choose JPEG, transparent areas will be filled with a white background. Choose PNG to keep the transparency."
	}

	options := []option{
		{
			emoji:  "📷",
			label:  "JPEG",
			desc:   jpegDesc,
			format: core.FormatJPEG,
		},
		{
			emoji:  "✏️",
			label:  "PNG",
			desc:   "Best for graphics with text, logos, sharp lines, or transparent backgrounds. Lossless — no quality is lost.",
			format: core.FormatPNG,
		},
		{
			emoji:  "⚡",
			label:  "Use defaults",
			desc:   "Not sure? Just use JPEG — it works well for most images.",
			format: core.FormatJPEG,
		},
	}

	cards := container.NewVBox()
	for _, o := range options {
		o := o // capture

		lbl := widget.NewLabelWithStyle(o.emoji+"  "+o.label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		desc := widget.NewLabel(o.desc)
		desc.Wrapping = fyne.TextWrapWord

		selectBtn := widget.NewButton("Select", func() {
			format := o.format
			if anyAlpha && format == core.FormatJPEG && o.label != "Use defaults" {
				// keep JPEG choice but the converter will handle per-file alpha
			}
			onConfirm(format)
		})
		selectBtn.Importance = widget.HighImportance

		bg := canvas.NewRectangle(color.NRGBA{R: 35, G: 38, B: 46, A: 255})
		bg.CornerRadius = 8

		card := container.NewStack(
			bg,
			container.NewPadded(
				container.NewBorder(nil, nil, nil, selectBtn,
					container.NewVBox(lbl, desc),
				),
			),
		)
		cards.Add(card)
	}

	return container.NewPadded(container.NewBorder(
		title,
		container.NewCenter(backBtn),
		nil, nil,
		container.NewVScroll(cards),
	))
}
