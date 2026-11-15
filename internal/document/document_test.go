package document

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

func TestMemoryDocumentSliceAndStringShareBackingBytes(t *testing.T) {
	data := []byte("ISO-10303-21\n")
	doc := NewMemoryDocument(data)

	if got := doc.Len(); got != uint64(len(data)) {
		t.Fatalf("Len() = %d, want %d", got, len(data))
	}
	if got := doc.Slice(0, 3); !bytes.Equal(got, []byte("ISO")) {
		t.Fatalf("Slice() = %q", got)
	}
	line := doc.LineString(0, 3)
	if line != "ISO" {
		t.Fatalf("LineString() = %q", line)
	}
	if unsafe.StringData(line) != unsafe.SliceData(data) {
		t.Fatal("LineString copied document bytes")
	}
	if err := doc.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestOpenMappedDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.ifc")
	data := []byte("#1=IFCPROJECT('sample');\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	doc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := doc.Close(); err != nil {
			t.Error(err)
		}
	}()
	if !bytes.Equal(doc.Bytes(), data) {
		t.Fatalf("mapped bytes = %q", doc.Bytes())
	}
}
