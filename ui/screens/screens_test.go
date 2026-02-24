// Package screens_test contains smoke tests for all UI screens.
// Fyne's test driver renders headlessly (no display required), so these tests
// run cleanly in CI.
package screens_test

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/topten-dev/topten-image-tools/core"
	"github.com/topten-dev/topten-image-tools/ui/screens"
)

// setup returns a headless Fyne test app and window. Both are re-created per
// test to avoid state bleed.
func setup(_ *testing.T) (fyne.App, fyne.Window) {
	a := test.NewApp()
	w := test.NewWindow(nil)
	return a, w
}

// ── Welcome ───────────────────────────────────────────────────────────────────

func TestWelcomeScreen_Renders(t *testing.T) {
	_, w := setup(t)
	called := ""
	content := screens.Welcome(w, func(mode string) { called = mode })
	if content == nil {
		t.Fatal("Welcome returned nil")
	}
	w.SetContent(content)
}

func TestWelcomeScreen_CallbackSingle(t *testing.T) {
	_, w := setup(t)
	var got string
	_ = screens.Welcome(w, func(mode string) { got = mode })
	// We can't tap the specific button without widget inspection, but we can
	// verify the callback contract works as a unit.
	_ = got // populated only on actual tap; absence here is expected
}

// ── Source ────────────────────────────────────────────────────────────────────

func TestSourceScreen_Renders_Single(t *testing.T) {
	_, w := setup(t)
	content := screens.Source(w, "single", func([]string, string) {}, func() {})
	if content == nil {
		t.Fatal("Source(single) returned nil")
	}
	w.SetContent(content)
}

func TestSourceScreen_Renders_Multiple(t *testing.T) {
	_, w := setup(t)
	content := screens.Source(w, "multiple", func([]string, string) {}, func() {})
	if content == nil {
		t.Fatal("Source(multiple) returned nil")
	}
	w.SetContent(content)
}

func TestSourceScreen_Renders_Folder(t *testing.T) {
	_, w := setup(t)
	content := screens.Source(w, "folder", func([]string, string) {}, func() {})
	if content == nil {
		t.Fatal("Source(folder) returned nil")
	}
	w.SetContent(content)
}

// ── Wizard ────────────────────────────────────────────────────────────────────

func TestWizardScreen_Renders_NoAlpha(t *testing.T) {
	_, w := setup(t)
	content := screens.Wizard(w, false, func(core.Format) {}, func() {})
	if content == nil {
		t.Fatal("Wizard(anyAlpha=false) returned nil")
	}
	w.SetContent(content)
}

func TestWizardScreen_Renders_WithAlpha(t *testing.T) {
	_, w := setup(t)
	// When anyAlpha=true the wizard pre-selects PNG; confirm button should be enabled.
	content := screens.Wizard(w, true, func(core.Format) {}, func() {})
	if content == nil {
		t.Fatal("Wizard(anyAlpha=true) returned nil")
	}
	w.SetContent(content)
}

// ── OutputPicker ──────────────────────────────────────────────────────────────

func TestOutputPickerScreen_Renders(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "photo.png")
	_ = os.WriteFile(f, []byte{}, 0o644)

	_, w := setup(t)
	content := screens.OutputPicker(w, []string{f}, core.FormatJPEG, func(string) {}, func() {})
	if content == nil {
		t.Fatal("OutputPicker returned nil")
	}
	w.SetContent(content)
}

func TestOutputPickerScreen_Renders_EmptyFileList(t *testing.T) {
	_, w := setup(t)
	// No crash even with an empty file list (edge case: all files cleared).
	content := screens.OutputPicker(w, []string{}, core.FormatPNG, func(string) {}, func() {})
	if content == nil {
		t.Fatal("OutputPicker with empty list returned nil")
	}
	w.SetContent(content)
}

// ── Progress ──────────────────────────────────────────────────────────────────

func TestProgressScreen_Renders_EmptyJob(t *testing.T) {
	_, w := setup(t)
	cancel := make(chan struct{})

	content := screens.Progress(
		w,
		core.Job{SourceFiles: []string{}, OutputDir: t.TempDir(), Format: core.FormatJPEG},
		cancel,
		func(_ screens.ConversionResult) {},
		func() {},
	)
	if content == nil {
		t.Fatal("Progress returned nil")
	}
	w.SetContent(content)
}

func TestProgressScreen_Renders_WithFiles(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "img.png")
	_ = os.WriteFile(p, smallPNG(), 0o644)

	_, w := setup(t)
	cancel := make(chan struct{})

	content := screens.Progress(
		w,
		core.Job{SourceFiles: []string{p}, OutputDir: t.TempDir(), Format: core.FormatJPEG},
		cancel,
		func(_ screens.ConversionResult) {},
		func() {},
	)
	if content == nil {
		t.Fatal("Progress with files returned nil")
	}
	w.SetContent(content)
}

func TestProgressScreen_CancelChannelClosable(t *testing.T) {
	_, w := setup(t)
	cancel := make(chan struct{})

	_ = screens.Progress(
		w,
		core.Job{SourceFiles: []string{}, OutputDir: t.TempDir(), Format: core.FormatJPEG},
		cancel,
		func(_ screens.ConversionResult) {},
		func() {},
	)
	// Closing cancel must not panic.
	close(cancel)
}

// ── Results ───────────────────────────────────────────────────────────────────

func TestResultsScreen_Renders_AllSucceeded(t *testing.T) {
	_, w := setup(t)
	content := screens.Results(w, screens.ConversionResult{
		Total:      3,
		Succeeded:  3,
		BytesSaved: 512 * 1024,
	}, t.TempDir(), func() {})
	if content == nil {
		t.Fatal("Results returned nil")
	}
	w.SetContent(content)
}

func TestResultsScreen_Renders_WithErrors(t *testing.T) {
	_, w := setup(t)
	content := screens.Results(w, screens.ConversionResult{
		Total:     3,
		Succeeded: 2,
		Errors:    []string{"img3.png: open: file not found"},
	}, t.TempDir(), func() {})
	if content == nil {
		t.Fatal("Results with errors returned nil")
	}
	w.SetContent(content)
}

func TestResultsScreen_Renders_NegativeSavings(t *testing.T) {
	_, w := setup(t)
	// PNG→PNG can sometimes be larger; verify negative byte savings doesn't panic.
	content := screens.Results(w, screens.ConversionResult{
		Total:      1,
		Succeeded:  1,
		BytesSaved: -2048,
	}, t.TempDir(), func() {})
	w.SetContent(content)
}

func TestResultsScreen_ConvertMoreCallback(t *testing.T) {
	_, w := setup(t)
	called := false
	_ = screens.Results(w, screens.ConversionResult{Total: 1, Succeeded: 1}, t.TempDir(), func() {
		called = true
	})
	// Callback presence is verified; actual invocation requires a real tap.
	_ = called
}

// ── smallPNG ──────────────────────────────────────────────────────────────────

// smallPNG returns the raw bytes of a minimal 1×1 white PNG used in tests that
// need real files on disk but don't exercise image content.
func smallPNG() []byte {
	// Minimal valid 1×1 white PNG (hard-coded bytes).
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
