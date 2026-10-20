// MorphoWheel for macOS: drag a file onto a sector of the circular menu to
// convert it. Hold Shift (convert), Control (tools) or Option (actions)
// while dragging, or press Control+Option+Space for files selected in Finder.
package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"morphowheel/internal/ops"
	"morphowheel/ui"
)

func main() {
	ui.Run(ui.Callbacks{
		Sectors: func(group, path string) []ui.Item {
			var items []ui.Item
			for _, o := range ops.For(ops.Group(group), path) {
				items = append(items, ui.Item{ID: o.ID, Title: o.Title})
			}
			return items
		},
		Drop:   runOp,
		Hotkey: fromFinder,
	})
}

// runOp applies one operation to every dropped file and reports the outcome.
func runOp(id string, paths []string) {
	op, ok := ops.ByID(id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	var done, failed int
	var lastErr error
	var lastOut string
	for _, p := range paths {
		out, err := ops.Execute(ctx, op, p)
		if err != nil {
			failed++
			lastErr = err
			continue
		}
		done++
		lastOut = out
	}
	switch {
	case failed > 0:
		ui.Notify("MorphoWheel: ошибка", fmt.Sprintf("%d из %d: %v", failed, len(paths), lastErr))
	case lastOut != "":
		ui.Notify("MorphoWheel", fmt.Sprintf("Готово: %s", filepath.Base(lastOut)))
		if done == 1 {
			_ = exec.Command("open", "-R", lastOut).Start()
		}
	}
}

// fromFinder shows the ring for the files currently selected in Finder.
func fromFinder() {
	const script = `tell application "Finder" to set sel to selection as alias list
set out to ""
repeat with f in sel
	set out to out & POSIX path of f & linefeed
end repeat
return out`
	raw, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		ui.Notify("MorphoWheel", "Нет доступа к Finder (System Settings → Privacy → Automation)")
		return
	}
	var paths []string
	for _, l := range strings.Split(string(raw), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			paths = append(paths, l)
		}
	}
	if len(paths) == 0 {
		ui.Notify("MorphoWheel", "В Finder ничего не выбрано")
		return
	}
	ui.ShowRing("convert", paths)
}
