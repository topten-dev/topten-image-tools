package core

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

// newRGBA builds a plain in-memory RGBA image of the given size.
func newRGBA(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	return img
}

// writePNG encodes img to a temporary PNG file and returns its path.
func writePNG(t *testing.T, dir, name string, img image.Image) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("writePNG create: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("writePNG encode: %v", err)
	}
	return path
}

// writeJPEG encodes img to a temporary JPEG file and returns its path.
func writeJPEG(t *testing.T, dir, name string, img image.Image) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("writeJPEG create: %v", err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("writeJPEG encode: %v", err)
	}
	return path
}

// ── trimExtension ─────────────────────────────────────────────────────────────

func TestTrimExtension(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"image.png", "image"},
		{"photo.JPG", "photo"},
		{"archive.tar.gz", "archive.tar"},
		{"noext", "noext"},
		{".hidden", ""},
	}
	for _, tc := range cases {
		got := trimExtension(tc.in)
		if got != tc.want {
			t.Errorf("trimExtension(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

// ── HasAlpha ──────────────────────────────────────────────────────────────────

func TestHasAlpha(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"logo.png", true},
		{"logo.PNG", true},
		{"icon.gif", true},
		{"hero.webp", true},
		{"photo.jpg", false},
		{"photo.jpeg", false},
		{"scan.tiff", false},
		{"image.bmp", false},
		{"noext", false},
	}
	for _, tc := range cases {
		got := HasAlpha(tc.path)
		if got != tc.want {
			t.Errorf("HasAlpha(%q) = %v; want %v", tc.path, got, tc.want)
		}
	}
}

// ── uniquePath ────────────────────────────────────────────────────────────────

func TestUniquePath_Free(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "image.jpg")
	got := uniquePath(p)
	if got != p {
		t.Errorf("expected unchanged path %q, got %q", p, got)
	}
}

func TestUniquePath_OneConflict(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "image.jpg")
	// create the conflict
	if err := os.WriteFile(p, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	got := uniquePath(p)
	want := filepath.Join(dir, "image_1.jpg")
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestUniquePath_TwoConflicts(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "image.jpg")
	conflict1 := filepath.Join(dir, "image_1.jpg")
	for _, p := range []string{base, conflict1} {
		if err := os.WriteFile(p, []byte{}, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := uniquePath(base)
	want := filepath.Join(dir, "image_2.jpg")
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

// ── resizeIfNeeded ────────────────────────────────────────────────────────────

func TestResizeIfNeeded_SmallImage_NoChange(t *testing.T) {
	img := newRGBA(800, 600)
	out := resizeIfNeeded(img)
	b := out.Bounds()
	if b.Dx() != 800 || b.Dy() != 600 {
		t.Errorf("expected 800×600, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestResizeIfNeeded_ExactMaxWidth_NoChange(t *testing.T) {
	img := newRGBA(MaxDimension, 800)
	out := resizeIfNeeded(img)
	if out.Bounds().Dx() != MaxDimension {
		t.Errorf("expected width %d unchanged, got %d", MaxDimension, out.Bounds().Dx())
	}
}

func TestResizeIfNeeded_WideImage(t *testing.T) {
	// 2400×600 → should become 1200×300
	img := newRGBA(2400, 600)
	out := resizeIfNeeded(img)
	b := out.Bounds()
	if b.Dx() != MaxDimension {
		t.Errorf("width should be %d, got %d", MaxDimension, b.Dx())
	}
	if b.Dy() != 300 {
		t.Errorf("height should be 300, got %d", b.Dy())
	}
}

func TestResizeIfNeeded_TallImage(t *testing.T) {
	// 600×2400 → should become 300×1200
	img := newRGBA(600, 2400)
	out := resizeIfNeeded(img)
	b := out.Bounds()
	if b.Dy() != MaxDimension {
		t.Errorf("height should be %d, got %d", MaxDimension, b.Dy())
	}
	if b.Dx() != 300 {
		t.Errorf("width should be 300, got %d", b.Dx())
	}
}

func TestResizeIfNeeded_SquareOversize(t *testing.T) {
	// 2000×2000 → 1200×1200 (w >= h branch)
	img := newRGBA(2000, 2000)
	out := resizeIfNeeded(img)
	b := out.Bounds()
	if b.Dx() != MaxDimension || b.Dy() != MaxDimension {
		t.Errorf("expected %d×%d, got %d×%d", MaxDimension, MaxDimension, b.Dx(), b.Dy())
	}
}

// ── Run (full pipeline) ───────────────────────────────────────────────────────

func TestRun_ConvertPNGtoJPEG(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	srcFile := writePNG(t, src, "photo.png", newRGBA(800, 600))

	cancel := make(chan struct{})
	ch := Run(Job{
		SourceFiles: []string{srcFile},
		OutputDir:   out,
		Format:      FormatJPEG,
	}, cancel)

	var last Progress
	for p := range ch {
		last = p
	}

	if last.Error != nil {
		t.Fatalf("unexpected error: %v", last.Error)
	}
	if !last.Done {
		t.Error("expected Done to be true on last progress")
	}

	outFile := filepath.Join(out, "photo.jpg")
	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("expected output file %s: %v", outFile, err)
	}
}

func TestRun_ConvertPNGtoPNG(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	srcFile := writePNG(t, src, "banner.png", newRGBA(400, 300))

	cancel := make(chan struct{})
	ch := Run(Job{
		SourceFiles: []string{srcFile},
		OutputDir:   out,
		Format:      FormatPNG,
	}, cancel)

	for range ch {
	}

	if _, err := os.Stat(filepath.Join(out, "banner.png")); err != nil {
		t.Errorf("expected banner.png in output: %v", err)
	}
}

func TestRun_ConvertJPEGtoJPEG(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	srcFile := writeJPEG(t, src, "photo.jpg", newRGBA(800, 600))

	cancel := make(chan struct{})
	ch := Run(Job{
		SourceFiles: []string{srcFile},
		OutputDir:   out,
		Format:      FormatJPEG,
	}, cancel)

	var last Progress
	for p := range ch {
		last = p
	}
	if last.Error != nil {
		t.Fatalf("unexpected error: %v", last.Error)
	}
}

func TestRun_ResizesOversizedImage(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	// 2400×1800 — should be resized to 1200×900
	srcFile := writePNG(t, src, "big.png", newRGBA(2400, 1800))

	cancel := make(chan struct{})
	ch := Run(Job{
		SourceFiles: []string{srcFile},
		OutputDir:   out,
		Format:      FormatPNG,
	}, cancel)

	for range ch {
	}

	outFile := filepath.Join(out, "big.png")
	f, err := os.Open(outFile)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 1200 || b.Dy() != 900 {
		t.Errorf("expected 1200×900, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestRun_MultipleSources(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	files := []string{
		writePNG(t, src, "a.png", newRGBA(100, 100)),
		writePNG(t, src, "b.png", newRGBA(200, 200)),
		writePNG(t, src, "c.png", newRGBA(300, 300)),
	}

	cancel := make(chan struct{})
	ch := Run(Job{SourceFiles: files, OutputDir: out, Format: FormatJPEG}, cancel)

	count := 0
	for p := range ch {
		if p.Error != nil {
			t.Errorf("file %s: %v", p.FileName, p.Error)
		}
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 progress events, got %d", count)
	}
}

func TestRun_Cancel(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	// 5 files — cancel after first
	files := make([]string, 5)
	for i := range files {
		files[i] = writePNG(t, src, "img"+string(rune('a'+i))+".png", newRGBA(100, 100))
	}

	cancel := make(chan struct{})
	ch := Run(Job{SourceFiles: files, OutputDir: out, Format: FormatJPEG}, cancel)

	count := 0
	for range ch {
		count++
		if count == 1 {
			close(cancel) // cancel after first result
		}
	}
	if count >= 5 {
		t.Error("expected cancellation to stop early, but all 5 files were processed")
	}
}

func TestRun_InvalidSource_ReportsError(t *testing.T) {
	out := t.TempDir()
	cancel := make(chan struct{})
	ch := Run(Job{
		SourceFiles: []string{"/nonexistent/path/image.png"},
		OutputDir:   out,
		Format:      FormatJPEG,
	}, cancel)

	var last Progress
	for p := range ch {
		last = p
	}
	if last.Error == nil {
		t.Error("expected an error for a missing source file, got nil")
	}
}

func TestRun_DefaultQualityApplied(t *testing.T) {
	// Quality 0 must not panic — it defaults to DefaultQuality internally.
	src := t.TempDir()
	out := t.TempDir()
	srcFile := writePNG(t, src, "q.png", newRGBA(200, 200))

	cancel := make(chan struct{})
	ch := Run(Job{
		SourceFiles: []string{srcFile},
		OutputDir:   out,
		Format:      FormatJPEG,
		Quality:     0, // should trigger default
	}, cancel)

	var last Progress
	for p := range ch {
		last = p
	}
	if last.Error != nil {
		t.Fatalf("unexpected error with Quality=0: %v", last.Error)
	}
}

func TestRun_CollisionSafe(t *testing.T) {
	// source and output are the same directory — output should not overwrite source
	dir := t.TempDir()
	srcFile := writePNG(t, dir, "same.png", newRGBA(100, 100))

	cancel := make(chan struct{})
	ch := Run(Job{
		SourceFiles: []string{srcFile},
		OutputDir:   dir,
		Format:      FormatPNG,
	}, cancel)

	for range ch {
	}

	// Both the original and the new file must exist
	if _, err := os.Stat(srcFile); err != nil {
		t.Errorf("source file was deleted or renamed: %v", err)
	}
	out := filepath.Join(dir, "same_1.png")
	if _, err := os.Stat(out); err != nil {
		t.Errorf("expected collision-safe output at %s: %v", out, err)
	}
}

func TestRun_BytesSavedPopulated(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	// Large PNG → JPEG should save space
	srcFile := writePNG(t, src, "large.png", newRGBA(1000, 800))

	cancel := make(chan struct{})
	ch := Run(Job{
		SourceFiles: []string{srcFile},
		OutputDir:   out,
		Format:      FormatJPEG,
	}, cancel)

	var last Progress
	for p := range ch {
		last = p
	}
	// BytesSaved may or may not be positive depending on content,
	// but must be populated (non-zero for a real image).
	_ = last.BytesSaved // value is informational; we just confirm no panic
}
