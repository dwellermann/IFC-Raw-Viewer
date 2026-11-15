//go:build !wasm

package index_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"stepview/internal/document"
	"stepview/internal/index"
	"stepview/internal/search"
)

func TestLargeMappedDocumentEndToEnd(t *testing.T) {
	const records = 250_000
	path := filepath.Join(t.TempDir(), "large.ifc")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := bufio.NewWriterSize(file, 1<<20)
	for id := 1; id <= records; id++ {
		if _, err := fmt.Fprintf(writer, "#%d=IFCWALL('WallName-%09d',#2,#3,$,$,$,$,.NOTDEFINED.);                         \n", id, id); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	doc, err := document.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := doc.Close(); err != nil {
			t.Error(err)
		}
	}()

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	started := time.Now()
	builder := index.NewBuilder(doc)
	builder.Wait()
	service := search.New(doc, builder)
	line, found := service.JumpToEntity(records)
	if !found || line != records-1 {
		t.Fatalf("entity lookup = (%d, %v)", line, found)
	}
	if matches := service.Search("wallname-000000123", 10); len(matches) != 1 || matches[0] != 122 {
		t.Fatalf("text search = %v", matches)
	}
	if referrers := service.FindReferrers(2, records); len(referrers) != records {
		t.Fatalf("reverse search matches = %d, want %d", len(referrers), records)
	}
	runtime.ReadMemStats(&after)
	t.Logf("records=%d bytes=%d duration=%s heap_delta=%d", records, doc.Len(), time.Since(started).Round(time.Millisecond), after.HeapAlloc-before.HeapAlloc)
}
