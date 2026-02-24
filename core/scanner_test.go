package core

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// touch creates an empty file at path for use in scanner tests.
func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("touch %s: %v", path, err)
	}
}

// ── ScanFolder ────────────────────────────────────────────────────────────────

func TestScanFolder_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := ScanFolder(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestScanFolder_ReturnsOnlyImages(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "photo.jpg"))
	touch(t, filepath.Join(dir, "logo.png"))
	touch(t, filepath.Join(dir, "readme.txt"))  // should be excluded
	touch(t, filepath.Join(dir, "data.csv"))    // should be excluded
	touch(t, filepath.Join(dir, "banner.webp")) // should be included

	files, err := ScanFolder(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 image files, got %d: %v", len(files), files)
	}
}

func TestScanFolder_CaseInsensitiveExtension(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "upper.PNG"))
	touch(t, filepath.Join(dir, "mixed.Jpg"))

	files, err := ScanFolder(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files for mixed-case extensions, got %d", len(files))
	}
}

func TestScanFolder_NonRecursive_IgnoresSubdir(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "top.jpg"))

	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, filepath.Join(sub, "nested.jpg"))

	files, err := ScanFolder(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("non-recursive scan: expected 1 file, got %d: %v", len(files), files)
	}
}

func TestScanFolder_Recursive_IncludesSubdir(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "top.jpg"))

	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, filepath.Join(sub, "nested.png"))

	files, err := ScanFolder(dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("recursive scan: expected 2 files, got %d: %v", len(files), files)
	}
}

func TestScanFolder_Recursive_DeepNesting(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, filepath.Join(deep, "deep.png"))

	files, err := ScanFolder(dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 deeply nested file, got %d", len(files))
	}
}

func TestScanFolder_AllSupportedExtensions(t *testing.T) {
	dir := t.TempDir()
	exts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif", ".webp"}
	for _, ext := range exts {
		touch(t, filepath.Join(dir, "file"+ext))
	}

	files, err := ScanFolder(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != len(exts) {
		t.Errorf("expected %d files, got %d", len(exts), len(files))
	}
}

func TestScanFolder_InvalidDir(t *testing.T) {
	_, err := ScanFolder("/nonexistent/directory/path", false)
	if err == nil {
		t.Error("expected an error for a nonexistent directory, got nil")
	}
}

func TestScanFolder_ReturnedPathsAreAbsolute(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "photo.jpg"))

	files, err := ScanFolder(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range files {
		if !filepath.IsAbs(f) {
			t.Errorf("expected absolute path, got %q", f)
		}
	}
}

// ── FilterSupported ───────────────────────────────────────────────────────────

func TestFilterSupported_AllImages(t *testing.T) {
	in := []string{"a.jpg", "b.PNG", "c.gif", "d.webp"}
	out := FilterSupported(in)
	if len(out) != 4 {
		t.Errorf("expected 4, got %d: %v", len(out), out)
	}
}

func TestFilterSupported_MixedInput(t *testing.T) {
	in := []string{"photo.jpg", "doc.pdf", "logo.png", "data.csv", "icon.gif"}
	got := FilterSupported(in)
	sort.Strings(got)
	want := []string{"icon.gif", "logo.png", "photo.jpg"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q; want %q", i, got[i], want[i])
		}
	}
}

func TestFilterSupported_Empty(t *testing.T) {
	out := FilterSupported(nil)
	if len(out) != 0 {
		t.Errorf("expected empty, got %v", out)
	}
}

func TestFilterSupported_NoImagesInInput(t *testing.T) {
	out := FilterSupported([]string{"readme.md", "main.go", "config.json"})
	if len(out) != 0 {
		t.Errorf("expected 0 images, got %v", out)
	}
}

func TestFilterSupported_FullPaths(t *testing.T) {
	in := []string{
		"/home/user/photos/vacation.jpg",
		"/home/user/docs/report.docx",
		"/home/user/icons/logo.PNG",
	}
	out := FilterSupported(in)
	if len(out) != 2 {
		t.Errorf("expected 2 image paths, got %d: %v", len(out), out)
	}
}
