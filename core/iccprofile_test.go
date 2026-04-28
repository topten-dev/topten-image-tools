package core

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

// fakeICC builds a stand-in "ICC profile" of the given length. The bytes are
// not a valid profile — we only need to verify byte-for-byte round-tripping
// through embed/extract.
func fakeICC(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i % 251)
	}
	return b
}

func TestEmbedAndExtractICC_JPEG_RoundTrip(t *testing.T) {
	// Encode a tiny JPEG.
	src := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		size int
	}{
		{"small", 200},
		{"medium", 4096},
		// Two cases that exercise multi-segment splitting.
		{"just_over_one_chunk", 65520},
		{"three_chunks", 65519*2 + 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile := fakeICC(tc.size)
			withICC, err := embedICCProfileJPEG(buf.Bytes(), profile)
			if err != nil {
				t.Fatalf("embed: %v", err)
			}
			// Output must still be a valid JPEG that decodes to the same pixels.
			img, err := jpeg.Decode(bytes.NewReader(withICC))
			if err != nil {
				t.Fatalf("decode after embed: %v", err)
			}
			if img.Bounds().Dx() != 8 || img.Bounds().Dy() != 8 {
				t.Errorf("bounds changed: %v", img.Bounds())
			}
			// Write to disk and run our extractor against it.
			tmp := filepath.Join(t.TempDir(), "out.jpg")
			if err := os.WriteFile(tmp, withICC, 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := extractICCProfile(tmp)
			if err != nil {
				t.Fatalf("extract: %v", err)
			}
			if !bytes.Equal(got, profile) {
				t.Errorf("profile round-trip mismatch: got %d bytes, want %d", len(got), len(profile))
			}
		})
	}
}

func TestEmbedICC_NoProfile_PassesThrough(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	got, err := embedICCProfileJPEG(buf.Bytes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, buf.Bytes()) {
		t.Error("nil profile must return bytes unchanged")
	}
}

func TestExtractICC_NoProfile(t *testing.T) {
	// Plain JPEG without any APP2 segment.
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	tmp := filepath.Join(t.TempDir(), "noicc.jpg")
	f, _ := os.Create(tmp)
	_ = jpeg.Encode(f, src, &jpeg.Options{Quality: 80})
	f.Close()

	got, err := extractICCProfile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected nil profile, got %d bytes", len(got))
	}
}

func TestRun_PreservesICCProfile_OnJPEGOutput(t *testing.T) {
	// Build a JPEG with a fake ICC profile, run conversion, verify the
	// profile survives in the output.
	srcDir := t.TempDir()
	outDir := t.TempDir()

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{200, 100, 50, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	profile := fakeICC(512)
	withICC, err := embedICCProfileJPEG(buf.Bytes(), profile)
	if err != nil {
		t.Fatal(err)
	}
	srcPath := filepath.Join(srcDir, "tagged.jpg")
	if err := os.WriteFile(srcPath, withICC, 0o644); err != nil {
		t.Fatal(err)
	}

	cancel := make(chan struct{})
	for range Run(Job{
		SourceFiles: []string{srcPath},
		OutputDir:   outDir,
		Format:      FormatJPEG,
		AddSuffix:   true,
	}, cancel) {
	}

	entries, _ := os.ReadDir(outDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 output file, got %d", len(entries))
	}
	got, err := extractICCProfile(filepath.Join(outDir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, profile) {
		t.Errorf("ICC profile not preserved: got %d bytes, want %d", len(got), len(profile))
	}
}
