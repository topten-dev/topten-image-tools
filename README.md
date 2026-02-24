# Topten Image Tools

A cross-platform desktop app that converts images into CMS-ready formats with sensible defaults — no technical knowledge required.

📖 **Not a developer?** See the [User Guide](docs/user-guide.md) for step-by-step instructions.

---

## Features

| | |
|---|---|
| **Single image** | Pick one file, choose output format & folder |
| **Multiple images** | Queue several files and convert them in one go |
| **Entire folder** | Convert every image in a folder at once |
| **Format wizard** | Three-card picker — choose JPEG, PNG, or "Use defaults" (JPEG) in one click |
| **Auto-resize** | Every image is scaled down to a maximum of **1 200 px** on either side, aspect ratio preserved, never upscaled |
| **Transparency handling** | When source files with alpha channels are detected, an inline warning appears on the JPEG card — transparent areas are filled with white if JPEG is chosen |
| **Safe output names** | Converted files are always written as `<name>_converted.<ext>` — originals are never overwritten |
| **Portable** | Single self-contained binary — no installer or runtime required |
| **Cross-platform** | macOS (Intel & Apple Silicon), Windows, Linux |

---

## Format wizard

One screen, three cards — click **Select** and move on:

| Card | Output format |
|---|---|
| 📷 **JPEG** | Photos, landscapes, product shots, gradients. Smallest file size. |
| ✏️ **PNG** | Graphics with text or logos, screenshots, transparent backgrounds. Lossless. |
| ⚡ **Use defaults** | Not sure? Picks JPEG — works well for most images. |

If the app detects that any source file has a transparent background, an inline note appears on the JPEG card. Those files are saved as PNG automatically; the rest of the batch follows your chosen format.

---

## Download

Pre-built binaries for every platform are attached to each [GitHub Release](../../releases).

| Platform | Archive | Contents |
|---|---|---|
| macOS Intel | `topten-image-tools-macos-intel.zip` | `.app` bundle |
| macOS Apple Silicon | `topten-image-tools-macos-arm64.zip` | `.app` bundle |
| Windows | `topten-image-tools-windows-amd64.zip` | `.exe` |
| Linux | `topten-image-tools-linux-amd64.tar.gz` | binary |

---

## Building from source

### Prerequisites

| Requirement | Notes |
|---|---|
| Go 1.26+ | <https://go.dev/dl/> |
| C compiler | `gcc` / Xcode CLT / MinGW — required by Fyne's CGO backend |
| Linux only | `sudo apt install libgl1-mesa-dev xorg-dev` |

### Run locally

```bash
git clone https://github.com/topten-dev/topten-image-tools.git
cd topten-image-tools
go mod tidy
go run .
```

### Build a standalone binary

```bash
# Linux / macOS
CGO_ENABLED=1 go build -ldflags="-s -w" -o topten-image-tools .

# Windows (PowerShell)
$env:CGO_ENABLED = "1"
go build -ldflags="-s -w -H=windowsgui" -o topten-image-tools.exe .
```

### Package a macOS .app bundle

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
fyne package -os darwin -name "Topten Image Tools" -appID "dev.topten.image-tools"
```

> **App icon** — place a 1 024 × 1 024 `Icon.png` in the project root before packaging. `FyneApp.toml` already points to it.

---

## Testing

The test suite covers the conversion pipeline, file scanning, and all UI screens (headless, no display required).

```bash
# Run all tests
go test ./...

# With verbose output and race detector
go test -v -race ./...

# One package only
go test ./core/...
```

### What is tested

| Package | Tests | Coverage |
|---|---|---|
| `core` (converter) | 17 | `trimExtension`, `HasAlpha`, `uniquePath` (collision cases), `resizeIfNeeded` (5 dimension scenarios), `Run` end-to-end (PNG→JPG, JPG→JPG, oversize resize, multi-file, cancel, error reporting, default quality, collision-safe output) |
| `core` (scanner) | 12 | `ScanFolder` (empty dir, images-only, case-insensitive extensions, non-recursive, recursive, deep nesting, all 8 extensions, invalid path, absolute paths), `FilterSupported` (mixed input, empty, no images, full paths) |
| `ui/screens` | 16 | Headless render smoke test for every screen variant: Welcome, Source ×3 modes, Wizard ×2 (with/without alpha), OutputPicker ×2, Progress ×3, Results ×4 |

---

## CI / CD

GitHub Actions runs tests and builds natively on each target OS.

```
push to main  →  test (Linux · Windows · macOS)  →  build (4 runners)
tag v*.*.*     →  … same …  →  build  →  release (archives attached automatically)
```

Workflow: [.github/workflows/build.yml](.github/workflows/build.yml)

---

## Project structure

```
topten-image-tools/
├── main.go                        # Entry point
├── FyneApp.toml                   # App metadata (fyne package)
├── Icon.png                       # 1 024×1 024 app icon
├── docs/
│   └── user-guide.md              # End-user usage guide (non-technical)
├── core/
│   ├── converter.go               # Resize + JPEG/PNG encode, byte-savings tracking
│   ├── converter_test.go
│   ├── scanner.go                 # Folder walk & extension filtering
│   └── scanner_test.go
├── ui/
│   ├── app.go                     # Shared state & linear screen navigation
│   └── screens/
│       ├── welcome.go             # Landing — mode selection cards
│       ├── source.go              # File / folder picker
│       ├── wizard.go              # Format recommendation wizard
│       ├── output.go              # Output folder picker
│       ├── progress.go            # Animated progress bar
│       ├── results.go             # Summary, space saved, open-folder button
│       └── screens_test.go        # Headless UI smoke tests
└── .github/
    └── workflows/
        └── build.yml              # Test → build → release pipeline
```

---

## Supported input formats

JPG · JPEG · PNG · GIF · BMP · TIFF · TIF · WebP

---

## License

See [LICENSE](LICENSE).

