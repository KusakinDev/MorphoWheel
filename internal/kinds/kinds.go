// Package kinds classifies files by extension.
package kinds

import (
	"path/filepath"
	"strings"
)

type Kind string

const (
	Unknown      Kind = ""
	Image        Kind = "image"
	RawImage     Kind = "raw"
	Audio        Kind = "audio"
	Video        Kind = "video"
	PDF          Kind = "pdf"
	Document     Kind = "document" // pandoc-friendly: docx, md, html, epub, rtf, txt
	Office       Kind = "office"   // LibreOffice: doc, odt, pages-less office formats
	Spreadsheet  Kind = "spreadsheet"
	Presentation Kind = "presentation"
	Subtitle     Kind = "subtitle"
	Archive      Kind = "archive"
)

var byExt = map[string]Kind{}

func init() {
	add := func(k Kind, exts ...string) {
		for _, e := range exts {
			byExt[e] = k
		}
	}
	add(Image, "jpg", "jpeg", "png", "gif", "bmp", "tif", "tiff", "heic", "heif", "webp", "avif", "ico")
	add(RawImage, "cr2", "cr3", "nef", "arw", "dng", "orf", "rw2", "raf")
	add(Audio, "mp3", "m4a", "aac", "wav", "flac", "ogg", "opus", "aiff", "wma")
	add(Video, "mp4", "mov", "mkv", "avi", "webm", "m4v", "wmv", "flv")
	add(PDF, "pdf")
	add(Document, "docx", "md", "markdown", "html", "htm", "epub", "rtf", "txt")
	add(Office, "doc", "odt")
	add(Spreadsheet, "xls", "xlsx", "ods", "csv")
	add(Presentation, "ppt", "pptx", "odp")
	add(Subtitle, "srt", "vtt", "ass")
	add(Archive, "zip", "tar", "gz", "tgz", "bz2", "xz", "7z", "rar")
}

// Ext returns the lowercase extension without the dot.
func Ext(path string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
}

// Of returns the kind of file at path, judged by extension.
func Of(path string) Kind {
	return byExt[Ext(path)]
}
