package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// Welcome renders the landing screen with the three conversion mode buttons.
func Welcome(w fyne.Window, onSelect func(mode string)) fyne.CanvasObject {
	title := canvas.NewText("TopTen Image Tools", color.White)
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("Convert images to CMS-ready format", color.NRGBA{R: 180, G: 180, B: 180, A: 255})
	subtitle.TextSize = 14
	subtitle.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(title),
		container.NewCenter(subtitle),
		widget.NewSeparator(),
	)

	singleCard := makeModeCard(
		"Single Image",
		"Pick one image file, choose the format and output folder.",
		"🖼",
		func() { onSelect("single") },
	)

	multiCard := makeModeCard(
		"Multiple Images",
		"Select several images at once and convert them together.",
		"🗂",
		func() { onSelect("multiple") },
	)

	folderCard := makeModeCard(
		"Entire Folder",
		"Convert all images inside a folder in one go.",
		"📁",
		func() { onSelect("folder") },
	)

	cards := container.NewGridWithColumns(3, singleCard, multiCard, folderCard)

	footer := canvas.NewText("All images are resized to a maximum of 1200 px on either side.", color.NRGBA{R: 130, G: 130, B: 130, A: 255})
	footer.TextSize = 11
	footer.Alignment = fyne.TextAlignCenter

	return container.NewPadded(
		container.NewBorder(
			container.NewVBox(container.NewPadded(header)),
			container.NewCenter(footer),
			nil, nil,
			container.NewCenter(container.NewPadded(cards)),
		),
	)
}

// makeModeCard builds a tappable card widget for one conversion mode.
func makeModeCard(title, desc, icon string, onTap func()) fyne.CanvasObject {
	iconLabel := canvas.NewText(icon, color.White)
	iconLabel.TextSize = 36
	iconLabel.Alignment = fyne.TextAlignCenter

	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	descLabel := widget.NewLabel(desc)
	descLabel.Wrapping = fyne.TextWrapWord
	descLabel.Alignment = fyne.TextAlignCenter

	btn := widget.NewButton("Select", onTap)
	btn.Importance = widget.HighImportance

	bg := canvas.NewRectangle(color.NRGBA{R: 40, G: 44, B: 52, A: 255})
	bg.CornerRadius = 8

	content := container.NewPadded(
		container.NewVBox(
			container.NewCenter(iconLabel),
			titleLabel,
			descLabel,
			container.NewCenter(btn),
		),
	)

	return container.NewStack(bg, content)
}
