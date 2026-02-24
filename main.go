package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/topten-dev/topten-image-tools/ui"
)

func main() {
	a := app.NewWithID("dev.topten.image-tools")
	w := a.NewWindow("TopTen Image Tools")
	w.Resize(fyne.NewSize(800, 580))
	w.SetFixedSize(false)
	w.CenterOnScreen()

	appState := ui.NewAppState(a, w)
	appState.ShowWelcome()

	w.ShowAndRun()
}
