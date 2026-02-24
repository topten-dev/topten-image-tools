package core

import (
	"os"
	"path/filepath"
	"strings"
)

// ScanFolder returns all supported image files found in dir (non-recursive).
// Set recursive to true to also walk sub-directories.
func ScanFolder(dir string, recursive bool) ([]string, error) {
	var files []string

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && SupportedExtensions[strings.ToLower(filepath.Ext(path))] {
				files = append(files, path)
			}
			return nil
		})
		return files, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if SupportedExtensions[ext] {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files, nil
}

// FilterSupported returns only paths that have a supported image extension.
func FilterSupported(paths []string) []string {
	var out []string
	for _, p := range paths {
		ext := strings.ToLower(filepath.Ext(p))
		if SupportedExtensions[ext] {
			out = append(out, p)
		}
	}
	return out
}
