package ops

import (
	"context"
	"os"
	"path/filepath"

	"morphowheel/internal/kinds"
	"morphowheel/internal/tools"
)

var (
	imageKinds = []kinds.Kind{kinds.Image, kinds.RawImage}
	avKinds    = []kinds.Kind{kinds.Audio, kinds.Video}
	pandocIn   = []kinds.Kind{kinds.Document}
	officeIn   = []kinds.Kind{kinds.Office, kinds.Spreadsheet, kinds.Presentation, kinds.Document}
)

func init() {
	// Images (RAW included): sips reads everything ImageIO supports.
	for _, f := range []string{"jpg", "png", "heic", "tiff", "gif", "bmp"} {
		f := f
		register(Op{ID: "img-" + f, Title: upper(f), Group: Convert, Kinds: imageKinds, OutExt: f,
			Run: func(ctx context.Context, in, out string) error { return sipsConvert(ctx, in, out, f) }})
	}
	for _, f := range []string{"webp", "avif"} {
		f := f
		register(Op{ID: "img-" + f, Title: upper(f), Group: Convert, Kinds: imageKinds, OutExt: f,
			Run: func(ctx context.Context, in, out string) error { return magickConvert(ctx, in, out) }})
	}
	register(Op{ID: "img-pdf", Title: "PDF", Group: Convert, Kinds: imageKinds, OutExt: "pdf",
		Run: func(ctx context.Context, in, out string) error { return magickConvert(ctx, in, out) }})

	// Audio and video via ffmpeg.
	for _, f := range []string{"mp3", "m4a", "wav", "flac", "ogg", "opus"} {
		f := f
		register(Op{ID: "av-" + f, Title: upper(f), Group: Convert, Kinds: avKinds, OutExt: f,
			Run: func(ctx context.Context, in, out string) error { return ffmpeg(ctx, in, out, "-vn") }})
	}
	for _, f := range []string{"mp4", "mov", "mkv", "webm"} {
		f := f
		register(Op{ID: "av-" + f, Title: upper(f), Group: Convert, Kinds: []kinds.Kind{kinds.Video}, OutExt: f,
			Run: func(ctx context.Context, in, out string) error { return ffmpeg(ctx, in, out) }})
	}
	register(Op{ID: "av-gif", Title: "GIF", Group: Convert, Kinds: []kinds.Kind{kinds.Video}, OutExt: "gif",
		Run: func(ctx context.Context, in, out string) error {
			return ffmpeg(ctx, in, out, "-vf", "fps=12,scale=480:-1:flags=lanczos")
		}})

	// Text documents via pandoc.
	for _, f := range []string{"docx", "md", "html", "epub", "rtf", "txt"} {
		f := f
		register(Op{ID: "doc-" + f, Title: upper(f), Group: Convert, Kinds: pandocIn, OutExt: f,
			Run: func(ctx context.Context, in, out string) error {
				return tools.Run(ctx, "pandoc", in, "-o", out)
			}})
	}

	// Office, spreadsheets, presentations and PDF output via LibreOffice.
	for _, f := range []string{"pdf", "docx", "odt"} {
		f := f
		register(Op{ID: "off-" + f, Title: upper(f), Group: Convert, Kinds: []kinds.Kind{kinds.Office, kinds.Document}, OutExt: f,
			Run: func(ctx context.Context, in, out string) error { return soffice(ctx, in, out, f) }})
	}
	for _, f := range []string{"pdf", "xlsx", "ods", "csv"} {
		f := f
		register(Op{ID: "sheet-" + f, Title: upper(f), Group: Convert, Kinds: []kinds.Kind{kinds.Spreadsheet}, OutExt: f,
			Run: func(ctx context.Context, in, out string) error { return soffice(ctx, in, out, f) }})
	}
	for _, f := range []string{"pdf", "pptx", "odp"} {
		f := f
		register(Op{ID: "pres-" + f, Title: upper(f), Group: Convert, Kinds: []kinds.Kind{kinds.Presentation}, OutExt: f,
			Run: func(ctx context.Context, in, out string) error { return soffice(ctx, in, out, f) }})
	}

	// PDF → image (first page, sips) and PDF → text/docx (LibreOffice/pandoc not
	// able to read PDF, so only the image route is offered).
	register(Op{ID: "pdf-png", Title: "PNG", Group: Convert, Kinds: []kinds.Kind{kinds.PDF}, OutExt: "png",
		Run: func(ctx context.Context, in, out string) error { return sipsConvert(ctx, in, out, "png") }})
	register(Op{ID: "pdf-jpg", Title: "JPG", Group: Convert, Kinds: []kinds.Kind{kinds.PDF}, OutExt: "jpg",
		Run: func(ctx context.Context, in, out string) error { return sipsConvert(ctx, in, out, "jpg") }})

	// Subtitles.
	for _, f := range []string{"srt", "vtt", "ass"} {
		f := f
		register(Op{ID: "sub-" + f, Title: upper(f), Group: Convert, Kinds: []kinds.Kind{kinds.Subtitle}, OutExt: f,
			Run: func(ctx context.Context, in, out string) error { return ffmpeg(ctx, in, out) }})
	}

	// Archives: extract into a folder named after the archive.
	register(Op{ID: "arc-extract", Title: "Распаковать", Group: Convert, Kinds: []kinds.Kind{kinds.Archive}, OutExt: "",
		Run: func(ctx context.Context, in, _ string) error {
			dir := OutputPath(in, "", "")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			if err := tools.Run(ctx, "tar", "-xf", in, "-C", dir); err != nil {
				os.RemoveAll(dir)
				return err
			}
			return nil
		}})
}

func upper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}

func sipsConvert(ctx context.Context, in, out, format string) error {
	if format == "jpg" {
		format = "jpeg"
	}
	return tools.Run(ctx, "sips", "-s", "format", format, in, "--out", out)
}

func magickConvert(ctx context.Context, in, out string) error {
	// sips can decode RAW; feed ImageMagick a PNG intermediate in that case.
	if kinds.Of(in) == kinds.RawImage {
		tmp, err := os.MkdirTemp("", "morphowheel")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmp)
		mid := filepath.Join(tmp, "raw.png")
		if err := sipsConvert(ctx, in, mid, "png"); err != nil {
			return err
		}
		in = mid
	}
	return tools.Run(ctx, "magick", in, out)
}

func ffmpeg(ctx context.Context, in, out string, extra ...string) error {
	args := []string{"-y", "-hide_banner", "-loglevel", "error", "-i", in}
	args = append(args, extra...)
	args = append(args, out)
	return tools.Run(ctx, "ffmpeg", args...)
}

// soffice converts via a throwaway output directory, then moves the result to
// the requested path, because LibreOffice picks the output name itself.
func soffice(ctx context.Context, in, out, format string) error {
	tmp, err := os.MkdirTemp("", "morphowheel")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := tools.Run(ctx, "soffice", "--headless", "--convert-to", format, "--outdir", tmp, in); err != nil {
		return err
	}
	matches, _ := filepath.Glob(filepath.Join(tmp, "*"))
	if len(matches) == 0 {
		return os.ErrNotExist
	}
	return os.Rename(matches[0], out)
}
