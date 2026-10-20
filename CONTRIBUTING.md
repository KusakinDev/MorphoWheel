# Contributing to MorphoWheel

Thanks for helping out. Bug reports, format requests and pull requests are all
welcome.

## Development setup

Requirements: macOS 12+, Go (version in `go.mod`), Xcode Command Line Tools
(`xcode-select --install`, needed for cgo).

```bash
git clone https://github.com/KusakinDev/MorphoWheel.git
cd MorphoWheel
make test      # go vet + go test
make run       # build build/MorphoWheel.app and open it
```

External converters (ffmpeg, ImageMagick, pandoc, LibreOffice) are optional for
development; tests that need a missing tool are skipped. `./start.sh
--skip-deps` sets everything else up without touching Homebrew.

## Project layout

| Path | Purpose |
|------|---------|
| `main.go` | entry point, runs operations, reads the Finder selection |
| `ui/` | cgo bridge: menu bar item, hotkey, ring window (`bridge.m`) |
| `internal/kinds/` | file type detection by extension |
| `internal/ops/` | operation registry (`convert.go`, `tools.go`) |
| `internal/tools/` | locates and runs external converters |

## Adding a conversion or tool

1. Add the extension to `internal/kinds/kinds.go` if the input type is new.
2. Register an `Op` in `internal/ops/convert.go` (Shift), `internal/ops/tools.go`
   (Control or Option). `Kinds` decides which files see the sector, `OutExt` the
   output extension.
3. Add a test in `internal/ops/ops_test.go`. Skip it with `t.Skip` when it needs
   a converter that may not be installed (see `tools.Has`).
4. Mention the format in the table in `README.md`.

## Pull requests

- Keep a PR focused on one change.
- Run `make test` and `gofmt -l .` before pushing; CI runs the same checks.
- Originals must never be overwritten: write results through `ops.OutputPath`.
- All user-visible text and code comments are in English.
- UI changes (`ui/bridge.m`) cannot be covered by tests, so describe in the PR
  how you checked them (macOS version, what you dragged).

## Reporting bugs

Use the bug report template. Include your macOS version, the file type, and for
UI problems whether Input Monitoring and Automation permissions are granted.
