// Package ui wires together all screens and holds shared application state.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/topten-dev/topten-image-tools/core"
	"github.com/topten-dev/topten-image-tools/ui/screens"
)

// AppState holds shared state and provides navigation helpers.
type AppState struct {
	App    fyne.App
	Window fyne.Window

	// Filled by source selection.
	SourceFiles []string
	SourceDir   string // set when mode == "folder"
	Mode        string // "single" | "multiple" | "folder"

	// Filled by the format wizard.
	Format core.Format

	// Filled by output selection.
	OutputDir string
	AddSuffix bool
}

// NewAppState creates an AppState and applies the custom app theme.
func NewAppState(a fyne.App, w fyne.Window) *AppState {
	a.Settings().SetTheme(theme.DarkTheme())
	return &AppState{App: a, Window: w}
}

// ShowWelcome renders the landing screen.
func (s *AppState) ShowWelcome() {
	s.Window.SetContent(screens.Welcome(
		s.Window,
		func(mode string) { s.onModeSelected(mode) },
	))
}

// onModeSelected is called when the user picks single / multiple / folder.
func (s *AppState) onModeSelected(mode string) {
	s.Mode = mode
	s.Window.SetContent(screens.Source(
		s.Window,
		mode,
		func(files []string, dir string) {
			s.SourceFiles = files
			s.SourceDir = dir
			s.showWizard()
		},
		s.ShowWelcome,
	))
}

// showWizard renders the format-selection wizard.
func (s *AppState) showWizard() {
	anyAlpha := false
	for _, f := range s.SourceFiles {
		if core.HasAlpha(f) {
			anyAlpha = true
			break
		}
	}

	s.Window.SetContent(screens.Wizard(
		s.Window,
		anyAlpha,
		func(fmt core.Format) {
			s.Format = fmt
			s.showOutputPicker()
		},
		s.ShowWelcome,
	))
}

// showOutputPicker renders the output-folder picker, then starts conversion.
func (s *AppState) showOutputPicker() {
	s.Window.SetContent(screens.OutputPicker(
		s.Window,
		s.SourceFiles,
		s.Format,
		func(outDir string, addSuffix bool) {
			s.OutputDir = outDir
			s.AddSuffix = addSuffix
			s.showProgress()
		},
		s.showWizard,
	))
}

// showProgress runs the conversion and renders the progress screen.
func (s *AppState) showProgress() {
	cancelCh := make(chan struct{})

	s.Window.SetContent(screens.Progress(
		s.Window,
		core.Job{
			SourceFiles: s.SourceFiles,
			OutputDir:   s.OutputDir,
			Format:      s.Format,
			AddSuffix:   s.AddSuffix,
		},
		cancelCh,
		func(results screens.ConversionResult) {
			s.Window.SetContent(screens.Results(
				s.Window,
				results,
				s.OutputDir,
				s.ShowWelcome,
			))
		},
		func() {
			close(cancelCh)
			s.ShowWelcome()
		},
	))
}
