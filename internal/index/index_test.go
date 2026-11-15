package index

import (
	"fmt"
	"strings"
	"testing"

	"stepview/internal/document"
)

func TestLineIndex(t *testing.T) {
	index := BuildLineIndex([]byte("a\nb\r\nc"))
	if index.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", index.Len())
	}
	start, end := index.Bounds(1, 6)
	if start != 2 || end != 5 {
		t.Fatalf("Bounds(1) = %d, %d", start, end)
	}
}

func TestBuilderSparseIDs(t *testing.T) {
	const records = 100_000
	var source strings.Builder
	source.Grow(records * 32)
	for n := 1; n <= records; n++ {
		id := uint64(n)
		if n == records {
			id = 25_000_000
		}
		fmt.Fprintf(&source, "#%d=IFCWALL('g%d',#1);\n", id, n)
	}
	builder := NewBuilder(document.NewMemoryDocument([]byte(source.String())))
	builder.Wait()

	if !builder.IsReady() || builder.Progress() != 1 {
		t.Fatalf("builder was not ready: progress=%f", builder.Progress())
	}
	if builder.Lines().Len() != records {
		t.Fatalf("line count = %d, want %d", builder.Lines().Len(), records)
	}
	for _, test := range []struct {
		id   uint64
		line uint64
	}{
		{1, 0},
		{25_000_000, records - 1},
	} {
		line, found := builder.Entities().Get(test.id)
		if !found || line != test.line {
			t.Fatalf("Get(%d) = (%d, %v), want (%d, true)", test.id, line, found, test.line)
		}
	}
	if _, found := builder.Entities().Get(24_999_999); found {
		t.Fatal("unexpected sparse entity")
	}
	entries := builder.Types().Entries()
	if len(entries) != 1 || entries[0].Name != "IFCWALL" || entries[0].Count != records {
		t.Fatalf("types = %#v", entries)
	}
}
