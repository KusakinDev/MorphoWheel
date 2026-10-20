package ops

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"morphowheel/internal/kinds"
	"morphowheel/internal/tools"
)

func init() {
	// Tools (Control): produce a modified copy.
	for _, w := range []int{1920, 1280, 800} {
		w := w
		register(Op{ID: "resize-" + strconv.Itoa(w), Title: "Width " + strconv.Itoa(w), Group: Tools,
			Kinds: imageKinds, OutExt: "jpg",
			Run: func(ctx context.Context, in, out string) error {
				return tools.Run(ctx, "sips", "--resampleWidth", strconv.Itoa(w), "-s", "format", "jpeg", in, "--out", out)
			}})
	}
	register(Op{ID: "compress-img", Title: "Compress", Group: Tools, Kinds: imageKinds, OutExt: "jpg",
		Run: func(ctx context.Context, in, out string) error {
			return tools.Run(ctx, "sips", "-s", "format", "jpeg", "-s", "formatOptions", "60", in, "--out", out)
		}})
	register(Op{ID: "strip-exif", Title: "Strip EXIF", Group: Tools, Kinds: []kinds.Kind{kinds.Image}, OutExt: "jpg",
		Run: func(ctx context.Context, in, out string) error {
			return tools.Run(ctx, "magick", in, "-strip", out)
		}})
	register(Op{ID: "compress-video", Title: "Compress", Group: Tools, Kinds: []kinds.Kind{kinds.Video}, OutExt: "mp4",
		Run: func(ctx context.Context, in, out string) error {
			return ffmpeg(ctx, in, out, "-c:v", "libx264", "-crf", "28", "-preset", "medium", "-c:a", "aac", "-b:a", "128k")
		}})
	register(Op{ID: "extract-audio", Title: "Extract audio", Group: Tools, Kinds: []kinds.Kind{kinds.Video}, OutExt: "m4a",
		Run: func(ctx context.Context, in, out string) error { return ffmpeg(ctx, in, out, "-vn", "-c:a", "aac") }})
	register(Op{ID: "mute-video", Title: "Mute", Group: Tools, Kinds: []kinds.Kind{kinds.Video}, OutExt: "mp4",
		Run: func(ctx context.Context, in, out string) error { return ffmpeg(ctx, in, out, "-an", "-c:v", "copy") }})
	register(Op{ID: "normalize-audio", Title: "Normalize", Group: Tools, Kinds: []kinds.Kind{kinds.Audio}, OutExt: "m4a",
		Run: func(ctx context.Context, in, out string) error {
			return ffmpeg(ctx, in, out, "-af", "loudnorm", "-c:a", "aac")
		}})

	// Actions (Option): operate on the file itself.
	all := []kinds.Kind{kinds.Image, kinds.RawImage, kinds.Audio, kinds.Video, kinds.PDF, kinds.Document,
		kinds.Office, kinds.Spreadsheet, kinds.Presentation, kinds.Subtitle, kinds.Archive, kinds.Unknown}
	register(Op{ID: "zip", Title: "Zip", Group: Actions, Kinds: all, OutExt: "zip",
		Run: func(ctx context.Context, in, out string) error {
			return tools.Run(ctx, "ditto", "-c", "-k", "--keepParent", in, out)
		}})
	register(Op{ID: "reveal", Title: "Reveal in Finder", Group: Actions, Kinds: all,
		Run: func(ctx context.Context, in, _ string) error {
			return exec.CommandContext(ctx, "open", "-R", in).Run()
		}})
	register(Op{ID: "copy-path", Title: "Copy path", Group: Actions, Kinds: all,
		Run: func(ctx context.Context, in, _ string) error {
			abs, err := filepath.Abs(in)
			if err != nil {
				return err
			}
			cmd := exec.CommandContext(ctx, "pbcopy")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				return err
			}
			if err := cmd.Start(); err != nil {
				return err
			}
			_, _ = stdin.Write([]byte(abs))
			stdin.Close()
			return cmd.Wait()
		}})
	register(Op{ID: "trash", Title: "Move to Trash", Group: Actions, Kinds: all,
		Run: func(ctx context.Context, in, _ string) error {
			abs, err := filepath.Abs(in)
			if err != nil {
				return err
			}
			if _, err := os.Stat(abs); err != nil {
				return err
			}
			script := `tell application "Finder" to delete POSIX file ` + strconv.Quote(abs)
			return exec.CommandContext(ctx, "osascript", "-e", script).Run()
		}})
}
