// Package tools locates and runs external converters (ffmpeg, pandoc, ...).
package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrMissing is returned when a required converter is not installed.
type ErrMissing struct {
	Name string
	Hint string
}

func (e ErrMissing) Error() string {
	return fmt.Sprintf("%s не найден (установите: %s)", e.Name, e.Hint)
}

var hints = map[string]string{
	"ffmpeg":  "brew install ffmpeg",
	"pandoc":  "brew install pandoc",
	"soffice": "brew install --cask libreoffice",
	"magick":  "brew install imagemagick",
}

// ExtraDirs are searched after PATH. GUI apps launched from Finder get a
// minimal PATH, so Homebrew locations are listed explicitly.
var ExtraDirs = []string{
	"/opt/homebrew/bin",
	"/usr/local/bin",
	"/Applications/LibreOffice.app/Contents/MacOS",
}

// AppToolsDir is where the app may keep its own downloaded binaries.
func AppToolsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Application Support", "MorphoWheel", "tools")
}

// Find returns the path of a tool, or ErrMissing.
func Find(name string) (string, error) {
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	dirs := append([]string{AppToolsDir()}, ExtraDirs...)
	for _, d := range dirs {
		p := filepath.Join(d, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() && st.Mode()&0o111 != 0 {
			return p, nil
		}
	}
	return "", ErrMissing{Name: name, Hint: hints[name]}
}

// Has reports whether the tool is available.
func Has(name string) bool {
	_, err := Find(name)
	return err == nil
}

// Run executes a tool and returns a trimmed stderr tail on failure.
func Run(ctx context.Context, name string, args ...string) error {
	path, err := Find(name)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if len(msg) > 400 {
			msg = msg[len(msg)-400:]
		}
		return fmt.Errorf("%s: %w: %s", name, err, msg)
	}
	return nil
}
