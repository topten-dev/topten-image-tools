# AGENTS.md — Project context for AI coding agents

This file documents the architecture, decisions, conventions, and gotchas for this project. Read it before making any changes.

---

## What this project is

A cross-platform desktop GUI application (Go + Fyne) that converts images to CMS-ready formats. It is aimed at **non-technical end users** who need to resize and reformat images for upload to a CMS.

End users are not developers. All UX copy, labels, and error messages must be plain English.

---

## Tech stack

| Layer | Technology | Notes |
|---|---|---|
| Language | Go 1.26+ | |
| GUI framework | Fyne v2 (latest: 2.7.3) | CGO required; Fyne uses OpenGL |
| Image processing | `github.com/disintegration/imaging` | Lanczos resampling, auto-orientation |
| CI / CD | GitHub Actions | Native matrix build on 4 runners |
| Packaging | `fyne package` | Produces `.app` (macOS), `.exe` (Windows), binary (Linux) |

---

## Project structure

```
topten-image-tools/
├── main.go                        # Entry point — creates Fyne app + window, calls ui.NewAppState
├── FyneApp.toml                   # App metadata consumed by fyne package
├── Icon.png                       # 1024×1024 app icon (required by fyne package)
├── go.mod / go.sum
├── AGENTS.md                      # ← this file
├── docs/
│   └── user-guide.md              # End-user documentation (non-technical)
├── core/
│   ├── converter.go               # All image conversion logic
│   ├── converter_test.go
│   ├── scanner.go                 # Folder scanning & extension filtering
│   └── scanner_test.go
├── ui/
│   ├── app.go                     # AppState struct — shared state + linear screen navigation
│   └── screens/
│       ├── welcome.go             # Home screen — three mode-selection cards
│       ├── source.go              # File / folder picker (adapts to mode)
│       ├── wizard.go              # Format recommendation wizard (2-question flow)
│       ├── output.go              # Output folder picker + "Convert Now" trigger
│       ├── progress.go            # Animated progress bar driven by core.Run goroutine
│       ├── results.go             # Summary screen — space saved, errors, open folder
│       └── screens_test.go        # Headless UI smoke tests (fyne/v2/test)
└── .github/
    └── workflows/
        └── build.yml              # test → build → release pipeline
```

---

## Navigation model

The app uses a **linear push/pop navigation** pattern managed entirely by `ui.AppState` in [ui/app.go](ui/app.go). There is no router. Each screen receives callbacks for "next" and "back". `AppState` methods (`ShowWelcome`, `showWizard`, `showProgress`, etc.) set `w.SetContent(...)` directly.

Screen flow:

```
Welcome → Source → Wizard → OutputPicker → Progress → Results
                ↖──────────────────────────────────────────────
                              (Convert More)
```

---

## Core conversion rules

Defined in [core/converter.go](core/converter.go):

- Max dimension: **1200 px** on either side (`MaxDimension = 1200`)
- Resize uses **Lanczos** resampling (high quality, `imaging.Lanczos`)
- Aspect ratio is **always preserved**; images below 1200 px are **never upscaled**
- JPEG alpha channels are flattened to **white** before encoding
- Default JPEG quality: **85** (`DefaultQuality = 85`)
- Output filenames **never overwrite** existing files — `uniquePath()` appends `_1`, `_2`, etc.
- Conversion runs in a **goroutine**; progress is sent via a `chan Progress`
- Cancellation is handled via a `cancel <-chan struct{}`; closing the channel stops the loop

---

## Format wizard logic

In [ui/screens/wizard.go](ui/screens/wizard.go):

| User answer | Format |
|---|---|
| Photos / natural images | JPG |
| Graphics with text or logos | PNG |
| Transparent background | PNG |
| Banner + has text/logos | PNG |
| Banner + no text (photographic) | JPG |

**Alpha override:** if `core.HasAlpha()` returns true for any source file, a JPG recommendation is silently upgraded to PNG regardless of wizard answers.

`HasAlpha` is extension-based only (`.png`, `.gif`, `.webp` → true). It does not decode the file.

---

## Supported input formats

`.jpg` `.jpeg` `.png` `.gif` `.bmp` `.tiff` `.tif` `.webp`

Defined in `core.SupportedExtensions` (map). Extension matching is **case-insensitive** via `strings.ToLower`.

---

## Key architectural decisions

### Why Go + Fyne?
- Single self-contained binary per platform — no installer, no runtime
- Native CGO build means the binary links against the OS's OpenGL; no bundled renderer
- Fyne's test driver (`fyne.io/fyne/v2/test`) allows headless UI tests in CI without a display

### Why `disintegration/imaging` instead of stdlib?
- Lanczos resampling is not in the Go stdlib image package
- `imaging.AutoOrientation` automatically corrects EXIF rotation on JPEGs

### Why linear navigation instead of a router?
- The conversion flow is strictly sequential with no branching after format selection
- A simple callback chain in `AppState` is easier to follow and test than a generic router

### Why `fyne.Do` for goroutine → UI updates?
- Fyne requires all widget mutations to happen on the main goroutine
- `fyne.Do(func(){...})` is the correct way to post work back from a background goroutine in Fyne v2.5+

---

## Build requirements

### All platforms
- Go 1.26+
- A C compiler (CGO is required by Fyne)

### Linux (including WSL2)
```bash
sudo apt install -y libgl1-mesa-dev xorg-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev
```

### macOS
- Xcode Command Line Tools: `xcode-select --install`

### Windows
- MinGW-w64 (pre-installed on the `windows-latest` GitHub Actions runner via MSYS2)

---

## Running locally

```bash
go mod tidy
go run .
```

### WSL2 note

The app opens a native Windows window via WSLg. If no window appears, try:

```bash
LIBGL_ALWAYS_SOFTWARE=1 go run .
```

If that works, add `export LIBGL_ALWAYS_SOFTWARE=1` to `~/.zshrc` for your dev environment. This is a WSL2/WSLg OpenGL driver issue only — it does not affect Windows or macOS end users.

---

## Running tests

Tests do **not** require a display — Fyne's test driver is headless.

```bash
# All tests
go test ./...

# With race detector (recommended)
go test -v -race ./...

# Core logic only
go test ./core/...

# UI screens only
go test ./ui/screens/...
```

### Test coverage summary

| Package | Count | Key areas |
|---|---|---|
| `core` (converter) | 17 | Resize, format conversion, cancellation, collision-safe naming, error handling |
| `core` (scanner) | 12 | ScanFolder (recursive + non-recursive), FilterSupported, edge cases |
| `ui/screens` | 16 | Headless render smoke tests for every screen and variant |

---

## CI / CD pipeline ([.github/workflows/build.yml](.github/workflows/build.yml))

```
push to main / PR
  └─ test job (ubuntu · windows · macos) — go test -v -race ./...
       └─ build job (needs: test)
            ├─ ubuntu-latest  → topten-image-tools-linux-amd64.tar.gz
            ├─ windows-latest → topten-image-tools-windows-amd64.zip
            ├─ macos-13 (Intel)      → topten-image-tools-macos-intel.zip
            └─ macos-14 (ARM)        → topten-image-tools-macos-arm64.zip

push tag v*.*.*
  └─ same test + build jobs
       └─ release job (needs: build)
            └─ GitHub Release created with all four archives attached
```

To trigger a release: `git tag v1.0.0 && git push origin v1.0.0`

---

## Package metadata

Defined in [FyneApp.toml](FyneApp.toml):

```toml
[Details]
  ID      = "dev.topten.image-tools"
  Name    = "TopTen Image Tools"
  Version = "1.0.0"
  Icon    = "Icon.png"
```

Bump `Version` here before tagging a release.

---

## App icon

- File: `Icon.png` in project root
- Required size: **1024 × 1024 px**
- Required for `fyne package` on macOS; optional but recommended for Windows/Linux
- Not committed to git (listed in `.gitignore` as `*.png` is not — it **is** tracked). Make sure it exists before running the release CI.

---

## Things to keep in mind when editing

- **Never mutate Fyne widgets from a goroutine directly.** Always wrap in `fyne.Do(func(){...})`.
- **`core` has no Fyne dependency** — keep it that way. All GUI concerns live in `ui/`.
- **`screens_test.go` uses the `screens_test` package** (external test package) — it imports `screens` as a user would. New exported symbols in `screens` are automatically available.
- **`uniquePath` is tested** — if you change collision-safe naming logic, update the tests.
- **`HasAlpha` is extension-based only** — it does not decode the file. This is intentional for performance (can be called on thousands of files instantly). Do not change it to decode files without considering the performance impact on large folders.
