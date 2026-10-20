package ops

import (
	"os"
	"path/filepath"
	"morphowheel/internal/tools"
	"testing"
)

func ids(list []Op) map[string]bool {
	m := map[string]bool{}
	for _, o := range list {
		m[o.ID] = true
	}
	return m
}

func TestForSkipsOwnFormat(t *testing.T) {
	got := ids(For(Convert, "/tmp/a.png"))
	if got["img-png"] {
		t.Error("png should not offer png")
	}
	if !got["img-jpg"] {
		t.Error("png should offer jpg")
	}
}

func TestForRawAndVideo(t *testing.T) {
	if !ids(For(Convert, "/tmp/a.CR3"))["img-jpg"] {
		t.Error("raw should offer jpg")
	}
	v := ids(For(Convert, "/tmp/a.mov"))
	if !v["av-mp3"] || !v["av-gif"] || v["av-mov"] {
		t.Errorf("unexpected video ops: %v", v)
	}
}

func TestActionsAvailableForUnknown(t *testing.T) {
	if len(For(Actions, "/tmp/file.xyz")) == 0 {
		t.Error("actions should apply to any file")
	}
	if len(For(Convert, "/tmp/file.xyz")) != 0 {
		t.Error("no conversions for unknown type")
	}
}

func TestOutputPathNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "a.png")
	os.WriteFile(in, nil, 0o644)
	os.WriteFile(filepath.Join(dir, "a.jpg"), nil, 0o644)
	if got, want := OutputPath(in, "", "jpg"), filepath.Join(dir, "a (1).jpg"); got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestImageConvertWithSips(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "a.png")
	// 1x1 PNG
	png := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0x0d, 'I', 'H', 'D', 'R', 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 0x1f, 0x15, 0xc4, 0x89, 0, 0, 0, 0x0d, 'I', 'D', 'A', 'T', 0x78, 0x9c, 0x63, 0xf8, 0xff, 0xff, 0x3f, 0, 5, 0xfe, 2, 0xfe, 0xa7, 0x35, 0x81, 0x84, 0, 0, 0, 0, 'I', 'E', 'N', 'D', 0xae, 0x42, 0x60, 0x82}
	os.WriteFile(in, png, 0o644)
	op, _ := ByID("img-jpg")
	out, err := Execute(t.Context(), op, in)
	if err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(out); err != nil || st.Size() == 0 {
		t.Errorf("output missing: %v", err)
	}
}

func convertFile(t *testing.T, name, content, opID string) string {
	t.Helper()
	dir := t.TempDir()
	in := filepath.Join(dir, name)
	os.WriteFile(in, []byte(content), 0o644)
	op, ok := ByID(opID)
	if !ok {
		t.Fatalf("no op %s", opID)
	}
	out, err := Execute(t.Context(), op, in)
	if err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(out); err != nil || st.Size() == 0 {
		t.Fatalf("output missing or empty: %v", err)
	}
	return out
}

func TestPandocMarkdownToDocx(t *testing.T) {
	if !tools.Has("pandoc") {
		t.Skip("pandoc not installed")
	}
	convertFile(t, "a.md", "# Hi\n\ntext\n", "doc-docx")
}

func TestLibreOfficeDocxToPDF(t *testing.T) {
	if !tools.Has("soffice") || !tools.Has("pandoc") {
		t.Skip("soffice/pandoc not installed")
	}
	docx := convertFile(t, "a.md", "# Hi\n\ntext\n", "doc-docx")
	op, _ := ByID("off-pdf")
	out, err := Execute(t.Context(), op, docx)
	if err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(out); err != nil || st.Size() == 0 {
		t.Fatalf("pdf missing: %v", err)
	}
}
