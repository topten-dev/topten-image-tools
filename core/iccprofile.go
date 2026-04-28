package core

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// extractICCProfile returns the raw ICC profile bytes embedded in the source
// image at path, or (nil, nil) if no profile is present. It supports JPEG,
// PNG and WebP containers — the three input formats this tool decodes that
// commonly carry ICC profiles. Other formats return (nil, nil).
//
// The returned bytes are the unwrapped ICC profile suitable for embedding
// directly into another container.
func extractICCProfile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return iccFromJPEG(data)
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return iccFromPNG(data)
	case len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return iccFromWebP(data)
	}
	return nil, nil
}

// iccFromJPEG walks the APP2 segments and concatenates ICC_PROFILE chunks.
// Per the ICC spec, large profiles are split across multiple APP2 segments
// each carrying a 14-byte header: "ICC_PROFILE\0" + chunk# + total#.
func iccFromJPEG(data []byte) ([]byte, error) {
	const sig = "ICC_PROFILE\x00"
	type chunk struct {
		idx, total byte
		body       []byte
	}
	var chunks []chunk

	i := 0
	for i < len(data)-1 {
		if data[i] != 0xFF {
			i++
			continue
		}
		for i < len(data) && data[i] == 0xFF {
			i++
		}
		if i >= len(data) {
			break
		}
		marker := data[i]
		i++
		// Standalone markers (no length).
		if marker == 0xD8 || marker == 0xD9 || (marker >= 0xD0 && marker <= 0xD7) {
			if marker == 0xD9 {
				break
			}
			continue
		}
		if i+1 >= len(data) {
			break
		}
		segLen := int(data[i])<<8 | int(data[i+1])
		if segLen < 2 || i+segLen > len(data) {
			break
		}
		body := data[i+2 : i+segLen]
		i += segLen

		// SOS — image data follows; nothing else useful for us.
		if marker == 0xDA {
			break
		}
		if marker != 0xE2 || len(body) < len(sig)+2 {
			continue
		}
		if string(body[:len(sig)]) != sig {
			continue
		}
		chunks = append(chunks, chunk{
			idx:   body[len(sig)],
			total: body[len(sig)+1],
			body:  body[len(sig)+2:],
		})
	}

	if len(chunks) == 0 {
		return nil, nil
	}
	// Reassemble in chunk order (chunks are 1-indexed per spec).
	total := chunks[0].total
	if total == 0 {
		total = byte(len(chunks))
	}
	ordered := make([][]byte, total)
	for _, c := range chunks {
		if c.idx >= 1 && int(c.idx) <= len(ordered) {
			ordered[c.idx-1] = c.body
		}
	}
	var out bytes.Buffer
	for _, p := range ordered {
		out.Write(p)
	}
	if out.Len() == 0 {
		return nil, nil
	}
	return out.Bytes(), nil
}

// iccFromPNG returns the decompressed ICC profile bytes from an iCCP chunk.
// Format per PNG spec:  name (1-79 bytes) \0  compressionMethod(1)  zlib(profile)
func iccFromPNG(data []byte) ([]byte, error) {
	pos := 8 // skip PNG signature
	for pos+8 <= len(data) {
		length := binary.BigEndian.Uint32(data[pos : pos+4])
		ctype := string(data[pos+4 : pos+8])
		end := pos + 8 + int(length)
		if end+4 > len(data) {
			return nil, nil
		}
		body := data[pos+8 : end]
		pos = end + 4 // skip CRC

		if ctype != "iCCP" {
			if ctype == "IDAT" || ctype == "IEND" {
				return nil, nil
			}
			continue
		}
		// Find name terminator.
		nul := bytes.IndexByte(body, 0)
		if nul < 0 || nul+2 > len(body) {
			return nil, nil
		}
		// body[nul+1] is compression method (must be 0 = zlib).
		if body[nul+1] != 0 {
			return nil, nil
		}
		// We need a zlib reader without pulling in another import.
		// compress/zlib is stdlib.
		return inflateZlib(body[nul+2:])
	}
	return nil, nil
}

// iccFromWebP returns the raw ICCP chunk bytes from a WebP RIFF container.
// WebP ICCP chunks are not compressed and contain the profile verbatim.
func iccFromWebP(data []byte) ([]byte, error) {
	// RIFF header is 12 bytes ("RIFF" size "WEBP"); chunks follow.
	pos := 12
	for pos+8 <= len(data) {
		ctype := string(data[pos : pos+4])
		size := binary.LittleEndian.Uint32(data[pos+4 : pos+8])
		end := pos + 8 + int(size)
		if end > len(data) {
			return nil, nil
		}
		body := data[pos+8 : end]
		// Chunks are word-aligned.
		pos = end
		if size%2 == 1 {
			pos++
		}
		if ctype == "ICCP" {
			b := make([]byte, len(body))
			copy(b, body)
			return b, nil
		}
		// Stop scanning at the bitstream chunks; ICCP appears before them.
		if ctype == "VP8 " || ctype == "VP8L" || ctype == "ANIM" {
			return nil, nil
		}
	}
	return nil, nil
}

// inflateZlib decompresses zlib-wrapped data (used by PNG iCCP).
func inflateZlib(in []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, fmt.Errorf("zlib: %w", err)
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

// embedICCProfileJPEG returns a copy of jpegBytes with one or more APP2
// "ICC_PROFILE" segments inserted directly after the SOI marker (before any
// JFIF/Exif APP segments — placement is conventional but not strict).
//
// The profile is split into chunks of at most maxICCChunkSize bytes per the
// ICC.1 spec (max 65533 bytes per APP2 segment minus the 14-byte header).
func embedICCProfileJPEG(jpegBytes, icc []byte) ([]byte, error) {
	if len(icc) == 0 {
		return jpegBytes, nil
	}
	if len(jpegBytes) < 2 || jpegBytes[0] != 0xFF || jpegBytes[1] != 0xD8 {
		return nil, fmt.Errorf("not a JPEG (missing SOI)")
	}
	const maxChunk = 65519 // 65533 segment payload max - 14 header bytes
	totalChunks := (len(icc) + maxChunk - 1) / maxChunk
	if totalChunks > 255 {
		return nil, fmt.Errorf("ICC profile too large: %d bytes", len(icc))
	}

	var out bytes.Buffer
	out.Grow(len(jpegBytes) + len(icc) + totalChunks*20)
	out.Write(jpegBytes[:2]) // SOI

	for i := 0; i < totalChunks; i++ {
		start := i * maxChunk
		end := start + maxChunk
		if end > len(icc) {
			end = len(icc)
		}
		chunk := icc[start:end]
		segLen := 2 + 12 + 2 + len(chunk) // length field + sig + idx/total + payload
		out.WriteByte(0xFF)
		out.WriteByte(0xE2) // APP2
		out.WriteByte(byte(segLen >> 8))
		out.WriteByte(byte(segLen & 0xFF))
		out.WriteString("ICC_PROFILE\x00")
		out.WriteByte(byte(i + 1))      // chunk index (1-based)
		out.WriteByte(byte(totalChunks)) // total chunks
		out.Write(chunk)
	}

	out.Write(jpegBytes[2:])
	return out.Bytes(), nil
}
