package screens

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/topten-dev/topten-image-tools/core"
)

// Source renders a screen that lets the user pick one image, several images, or
// a folder, depending on mode. onConfirm is called with the resolved file list
// and (optionally) the source directory.
func Source(
	w fyne.Window,
	mode string,
	onConfirm func(files []string, dir string),
	onBack func(),
) fyne.CanvasObject {

	var selectedFiles []string
	var selectedDir string

	titleText, helpText, btnLabel := modeLabels(mode)

	title := widget.NewLabelWithStyle(titleText, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	help := widget.NewLabel(helpText)
	help.Wrapping = fyne.TextWrapWord
	help.Alignment = fyne.TextAlignCenter

	fileList := widget.NewLabel("No files selected.")
	fileList.Wrapping = fyne.TextWrapWord

	scrollList := container.NewVScroll(fileList)
	scrollList.SetMinSize(fyne.NewSize(600, 200))

	nextBtn := widget.NewButton("Next →", nil)
	nextBtn.Importance = widget.HighImportance
	nextBtn.Disable()

	updateList := func() {
		if len(selectedFiles) == 0 {
			fileList.SetText("No files selected.")
			nextBtn.Disable()
			return
		}
		text := ""
		for _, f := range selectedFiles {
			text += "• " + filepath.Base(f) + "\n"
		}
		fileList.SetText(text)
		nextBtn.Enable()
	}

	pickBtn := widget.NewButton(btnLabel, func() {
		switch mode {
		case "single":
			dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
				if err != nil || r == nil {
					return
				}
				r.Close()
				path := r.URI().Path()
				filtered := core.FilterSupported([]string{path})
				if len(filtered) > 0 {
					selectedFiles = filtered
					selectedDir = ""
					updateList()
				} else {
					dialog.ShowInformation("Unsupported file", "Please select a supported image file (JPG, PNG, GIF, BMP, TIFF, WebP).", w)
				}
			}, w)

		case "multiple":
			dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
				if err != nil || r == nil {
					return
				}
				r.Close()
				path := r.URI().Path()
				// Fyne's file open dialog is single-file; we accumulate picks.
				filtered := core.FilterSupported([]string{path})
				if len(filtered) > 0 {
					selectedFiles = append(selectedFiles, filtered...)
					updateList()
				}
			}, w)

		case "folder":
			dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
				if err != nil || lu == nil {
					return
				}
				dir := lu.Path()
				files, scanErr := core.ScanFolder(dir, false)
				if scanErr != nil {
					dialog.ShowError(scanErr, w)
					return
				}
				if len(files) == 0 {
					dialog.ShowInformation("No images found", "The selected folder contains no supported image files.", w)
					return
				}
				selectedDir = dir
				selectedFiles = files
				updateList()
			}, w)
		}
	})
	pickBtn.Importance = widget.MediumImportance

	clearBtn := widget.NewButton("Clear", func() {
		selectedFiles = nil
		selectedDir = ""
		updateList()
	})

	nextBtn.OnTapped = func() {
		onConfirm(selectedFiles, selectedDir)
	}

	backBtn := widget.NewButton("← Back", onBack)

	buttons := container.NewHBox(backBtn, widget.NewSeparator(), pickBtn, clearBtn)
	if mode == "multiple" {
		hint := widget.NewLabel("Tip: click 'Add Image' multiple times to build up your list.")
		hint.TextStyle = fyne.TextStyle{Italic: true}
		return container.NewPadded(container.NewBorder(
			container.NewVBox(title, help, buttons, hint),
			container.NewCenter(nextBtn),
			nil, nil,
			scrollList,
		))
	}

	return container.NewPadded(container.NewBorder(
		container.NewVBox(title, help, buttons),
		container.NewCenter(nextBtn),
		nil, nil,
		scrollList,
	))
}

func modeLabels(mode string) (title, help, btn string) {
	switch mode {
	case "single":
		return "Select an Image",
			"Choose the image file you want to convert.",
			"Browse…"
	case "multiple":
		return "Select Multiple Images",
			"Click 'Add Image' for each file you want to convert. All files will be processed together.",
			"Add Image"
	default: // folder
		return "Select a Folder",
			"Choose the folder containing images to convert. All supported images at the top level will be included.",
			"Browse Folder…"
	}
}
