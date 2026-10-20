# MorphoWheel

Free file converter for macOS with a circular drop menu. Drag a file onto a
sector of the wheel and get the converted copy next to the original.

Inspired by [Sweetch](https://github.com/krllmrzv/Sweetch) (Windows), rewritten
in Go with a thin Objective-C layer for the native UI.

## Quick start

```bash
./start.sh
```

The script checks Go and Command Line Tools, installs missing converters via
Homebrew, runs the tests, builds and installs the app, and enables start at
login. Options:

| Flag | Effect |
|------|--------|
| `--no-autostart` | install, but do not start at login |
| `--skip-deps` | do not install anything with Homebrew |
| `--uninstall` | remove the app, the `morphowheel` command and autostart |

## Usage

| Hold while dragging a file | Group | Examples |
|----------------------------|-------|----------|
| **Shift** | Convert | JPG/PNG/HEIC/WebP, MP3/M4A/FLAC, MP4/MOV/GIF, DOCX/MD/PDF |
| **Control** | Tools | resize, compress, extract audio, mute, strip EXIF |
| **Option** | Actions | ZIP, reveal in Finder, copy path, move to Trash |

A ring appears under the cursor; drop the file on a sector. The result is saved
next to the source and never overwrites it (`photo.jpg`, `photo (1).jpg`, ...).

Without dragging: select files in Finder and press **Control+Option+Space**,
then click a sector. While the ring is open, keys `1` / `2` / `3` switch between
Convert / Tools / Actions, `Esc` closes it.

## Supported formats

| Kind | Input | Backend |
|------|-------|---------|
| Images | JPG, PNG, GIF, BMP, TIFF, HEIC, WebP, AVIF | `sips` (WebP/AVIF output: ImageMagick) |
| Camera RAW | CR2, CR3, NEF, ARW, DNG, ORF, RW2, RAF | `sips` |
| Audio | MP3, M4A, AAC, WAV, FLAC, OGG, OPUS, AIFF | ffmpeg |
| Video | MP4, MOV, MKV, AVI, WebM, M4V | ffmpeg |
| Documents | DOCX, MD, HTML, EPUB, RTF, TXT | pandoc |
| Office | DOC, ODT, XLS(X), ODS, CSV, PPT(X), ODP | LibreOffice |
| PDF | PDF → PNG/JPG (first page) | `sips` |
| Subtitles | SRT, VTT, ASS | ffmpeg |
| Archives | ZIP, TAR, GZ, BZ2, XZ, 7Z, RAR (extract) | `tar` |

Which sectors you see depends on the file type. A missing converter produces a
notification with the install command.

## Manual build

Requires Go and Xcode Command Line Tools (cgo).

```bash
make build            # ./MorphoWheel
make app              # build/MorphoWheel.app (ad-hoc signed)
make run              # build and open the app
make test             # go vet + go test
make install          # ~/Applications/MorphoWheel.app + ~/.local/bin/morphowheel
make autostart-on     # install + LaunchAgent (start at login)
make autostart-off
make uninstall
```

Run the installed app from a terminal with `open -a MorphoWheel`; launching the
bare `morphowheel` binary keeps the terminal busy.

## Permissions

macOS asks on first use:

- **Input Monitoring**: to notice Shift/Control/Option while a file is dragged.
- **Automation → Finder**: to read the selection for the hotkey, and for
  "Move to Trash".

The build is ad-hoc signed, so after a rebuild macOS may treat the app as new
and ask again.

## Layout

```
main.go                entry point, runs operations, reads Finder selection
ui/                    cgo bridge: menu bar item, hotkey, ring window (bridge.m)
internal/kinds/        file type detection by extension
internal/ops/          operation registry: convert.go, tools.go
internal/tools/        locates and runs external converters
start.sh, Makefile     setup, build, install, autostart
```

Converters are searched in `PATH`, then
`~/Library/Application Support/MorphoWheel/tools`, `/opt/homebrew/bin`,
`/usr/local/bin` and the LibreOffice app bundle.

## Troubleshooting

- **Ring does not appear while dragging**: check Input Monitoring permission.
- **Hotkey does nothing**: check Automation → Finder; another app may own
  Control+Option+Space.
- **"ffmpeg не найден" after autostart**: apps started at login have a minimal
  `PATH`. `./start.sh` links converters from custom locations into the tools
  directory above; rerun it after installing ffmpeg elsewhere.
- **Logs of the autostarted app**: `log show --last 5m --predicate 'process == "MorphoWheel"'`
