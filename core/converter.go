// Package core provides image conversion logic.
package core

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp" // register WebP decoder
)

// MaxDimension is the maximum width or height after resizing.
const MaxDimension = 1200

// Format represents the target image format.
type Format string

const (
	FormatJPEG Format = "jpg"
	FormatPNG  Format = "png"
)

// Job holds the parameters for converting a batch of images.
type Job struct {
	SourceFiles []string
	OutputDir   string
	Format      Format
	Quality     int  // JPEG quality 1–100; ignored for PNG
	AddSuffix   bool // when true, output files get a "_converted" suffix
}

// Progress reports the state of a running conversion.
type Progress struct {
	Total     int
	Current   int
	FileName  string
	Error     error
	Done      bool
	BytesSaved int64
}

// DefaultQuality is the JPEG quality used when none is specified.
const DefaultQuality = 85

// Run converts all images in the job and feeds progress updates through the
// returned channel. The channel is closed when all work is finished.
func Run(job Job, cancel <-chan struct{}) <-chan Progress {
	ch := make(chan Progress, 1)
	if job.Quality == 0 {
		job.Quality = DefaultQuality
	}

	go func() {
		defer close(ch)
		total := len(job.SourceFiles)
		var bytesSaved int64

		for i, src := range job.SourceFiles {
			select {
			case <-cancel:
				return
			default:
			}

			p := Progress{
				Total:    total,
				Current:  i + 1,
				FileName: filepath.Base(src),
			}

			saved, err := convertFile(src, job.OutputDir, job.Format, job.Quality, job.AddSuffix)
			if err != nil {
				p.Error = fmt.Errorf("%s: %w", filepath.Base(src), err)
			}
			bytesSaved += saved
			p.BytesSaved = bytesSaved

			if i == total-1 {
				p.Done = true
			}

			ch <- p
		}
	}()

	return ch
}

// convertFile converts a single image file and returns the byte savings (can be
// negative if the output is larger than the input).
func convertFile(src, outputDir string, format Format, quality int, addSuffix bool) (saved int64, err error) {
	// Open source image.
	img, err := imaging.Open(src, imaging.AutoOrientation(true))
	if err != nil {
		return 0, fmt.Errorf("open: %w", err)
	}

	// Extract any embedded ICC colour profile from the source so we can
	// re-embed it in the output. Without this, profile-tagged sources
	// (common in product photos exported from Capture One / Photoshop or
	// produced as WebP with an ICCP chunk) render with a colour shift —
	// typically a grey cast on what should be white backgrounds — when
	// viewers fall back to the untagged-sRGB default.
	//
	// We pass the bytes through verbatim rather than transforming pixels:
	// proper transforms from arbitrary profiles need a full colour
	// management library (lcms2). Pass-through is correct for any
	// profile-aware viewer (browsers, Photoshop, most CMS image renderers).
	var iccProfile []byte
	if format == FormatJPEG {
		iccProfile, _ = extractICCProfile(src) // best-effort; ignore errors
	}

	// For JPEG, flatten alpha to white BEFORE resizing so Lanczos
	// interpolation operates on fully opaque pixels. Otherwise transparent
	// pixels (RGB 0,0,0) bleed grey into neighbouring colours during
	// resampling—especially visible with WebP sources.
	//
	// IMPORTANT: only flatten when the decoded image actually has an alpha
	// channel. Running Overlay on an already-opaque image (e.g. a JPEG that
	// decoded to YCbCr) is a wasted round-trip through NRGBA blending that
	// can darken near-white pixels by a few levels, producing a visible
	// grey cast on white backgrounds.
	if format == FormatJPEG && imageHasAlpha(img) {
		flat := imaging.New(img.Bounds().Dx(), img.Bounds().Dy(), color.White)
		img = imaging.Overlay(flat, img, image.Point{}, 1.0)
	}

	// Resize if needed, preserving aspect ratio.
	img = resizeIfNeeded(img)

	// Build output path.
	baseName := trimExtension(filepath.Base(src))
	ext := "." + string(format)
	suffix := ""
	if addSuffix {
		suffix = "_converted"
	}
	outPath := filepath.Join(outputDir, baseName+suffix+ext)
	// Avoid overwriting an existing file with the same name.
	outPath = uniquePath(outPath)

	out, err := os.Create(outPath)
	if err != nil {
		return 0, fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	switch format {
	case FormatJPEG:
		// Encode to a buffer first so we can splice the ICC profile in.
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return 0, fmt.Errorf("encode jpeg: %w", err)
		}
		jpegBytes := buf.Bytes()
		if len(iccProfile) > 0 {
			if withICC, err := embedICCProfileJPEG(jpegBytes, iccProfile); err == nil {
				jpegBytes = withICC
			}
			// On embed failure we still write the un-tagged JPEG rather
			// than fail the whole conversion.
		}
		if _, err := out.Write(jpegBytes); err != nil {
			return 0, fmt.Errorf("write jpeg: %w", err)
		}
	case FormatPNG:
		if err := png.Encode(out, img); err != nil {
			return 0, fmt.Errorf("encode png: %w", err)
		}
	default:
		return 0, fmt.Errorf("unsupported format: %s", format)
	}

	srcInfo, _ := os.Stat(src)
	outInfo, _ := os.Stat(outPath)
	if srcInfo != nil && outInfo != nil {
		saved = srcInfo.Size() - outInfo.Size()
	}
	return saved, nil
}

// resizeIfNeeded shrinks the image so neither side exceeds MaxDimension.
// Aspect ratio is always preserved and upscaling never happens.
func resizeIfNeeded(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= MaxDimension && h <= MaxDimension {
		return img
	}
	if w >= h {
		return imaging.Resize(img, MaxDimension, 0, imaging.Lanczos)
	}
	return imaging.Resize(img, 0, MaxDimension, imaging.Lanczos)
}

// trimExtension removes the file extension from a base name.
func trimExtension(name string) string {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext)
}

// uniquePath appends _1, _2, … to the stem until the path does not exist.
func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	stem := strings.TrimSuffix(filepath.Base(path), ext)
	for i := 1; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", stem, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return path
}

// HasAlpha returns true when the source image file appears to contain an alpha
// channel (i.e., is a PNG/GIF/WebP).
func HasAlpha(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	// PNG and GIF always support alpha; WebP can too.
	switch ext {
	case ".png", ".gif", ".webp":
		return true
	}
	return false
}

// imageHasAlpha reports whether the decoded image's color model carries an
// alpha channel. Opaque formats like YCbCr (JPEG), Gray, and CMYK do not, so
// they should never be passed through alpha-flattening.
func imageHasAlpha(img image.Image) bool {
	switch img.ColorModel() {
	case color.NRGBAModel, color.NRGBA64Model,
		color.RGBAModel, color.RGBA64Model,
		color.AlphaModel, color.Alpha16Model:
		return true
	}
	// Paletted images (e.g. GIF, indexed PNG) may include a transparent
	// index entry, so treat them as potentially alpha-bearing.
	if _, ok := img.(*image.Paletted); ok {
		return true
	}
	return false
}

// SupportedExtensions lists the input file extensions this tool accepts.
var SupportedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".tiff": true,
	".tif":  true,
	".webp": true,
}
