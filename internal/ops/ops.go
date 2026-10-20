// Package ops is the registry of operations offered in the circular menu.
package ops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"morphowheel/internal/kinds"
)

// Group maps to the modifier key held while dragging.
type Group string

const (
	Convert Group = "convert" // Shift
	Tools   Group = "tools"   // Control
	Actions Group = "actions" // Option
)

// Op is a single menu sector.
type Op struct {
	ID    string
	Title string
	Group Group
	Kinds []kinds.Kind
	// OutExt is the extension of the produced file; empty for operations
	// that do not produce a file (e.g. reveal in Finder).
	OutExt string
	Run    func(ctx context.Context, in, out string) error
}

var registry []Op

func register(o Op) { registry = append(registry, o) }

// For returns the operations of a group applicable to the file, in
// registration order. Converting a file to its own format is skipped.
func For(g Group, path string) []Op {
	k := kinds.Of(path)
	ext := kinds.Ext(path)
	var res []Op
	for _, o := range registry {
		if o.Group != g || !hasKind(o.Kinds, k) {
			continue
		}
		if g == Convert && o.OutExt == ext {
			continue
		}
		res = append(res, o)
	}
	return res
}

// ByID looks an operation up by its identifier.
func ByID(id string) (Op, bool) {
	for _, o := range registry {
		if o.ID == id {
			return o, true
		}
	}
	return Op{}, false
}

func hasKind(list []kinds.Kind, k kinds.Kind) bool {
	for _, x := range list {
		if x == k {
			return true
		}
	}
	return false
}

// OutputPath returns a path next to the source that does not exist yet, so
// originals are never overwritten. suffix is inserted before the extension.
func OutputPath(in, suffix, ext string) string {
	dir := filepath.Dir(in)
	base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
	name := base + suffix
	if ext != "" {
		ext = "." + ext
	}
	p := filepath.Join(dir, name+ext)
	for i := 1; exists(p); i++ {
		p = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
	}
	return p
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

// Execute runs op on in and returns the produced path (empty if none).
func Execute(ctx context.Context, op Op, in string) (string, error) {
	out := ""
	if op.OutExt != "" {
		suffix := ""
		if op.Group != Convert {
			suffix = "-" + op.ID
		}
		out = OutputPath(in, suffix, op.OutExt)
	}
	if err := op.Run(ctx, in, out); err != nil {
		if out != "" {
			os.Remove(out)
		}
		return "", err
	}
	return out, nil
}
