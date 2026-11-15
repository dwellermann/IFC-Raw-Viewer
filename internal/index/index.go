// Package index builds compact, concurrent-safe indexes over STEP documents.
package index

import (
	"bytes"
	"sort"
	"sync"
	"sync/atomic"

	"stepview/internal/document"
	"stepview/internal/step"
)

const pageShift = 12
const pageSize = 1 << pageShift

// LineIndex stores only physical line start offsets.
type LineIndex struct {
	starts []uint64
}

// BuildLineIndex scans source for line starts without creating line strings.
func BuildLineIndex(source []byte) *LineIndex {
	starts := make([]uint64, 0, len(source)/48+1)
	if len(source) == 0 {
		return &LineIndex{starts: starts}
	}
	starts = append(starts, 0)
	for offset := 0; offset < len(source); {
		found := bytes.IndexByte(source[offset:], '\n')
		if found < 0 {
			break
		}
		offset += found + 1
		if offset < len(source) {
			starts = append(starts, uint64(offset))
		}
	}
	return &LineIndex{starts: starts}
}

func (i *LineIndex) Len() uint64 {
	return uint64(len(i.starts))
}

func (i *LineIndex) Start(line uint64) uint64 {
	return i.starts[line]
}

// Bounds returns the half-open source range for a physical line.
func (i *LineIndex) Bounds(line uint64, sourceLen uint64) (start, end uint64) {
	start = i.starts[line]
	if line+1 < uint64(len(i.starts)) {
		return start, i.starts[line+1]
	}
	return start, sourceLen
}

// PagedEntityIndex maps sparse entity IDs to line number + 1.
type PagedEntityIndex struct {
	mu    sync.RWMutex
	pages [][]uint64
}

func (i *PagedEntityIndex) Set(id, line uint64) {
	page := id >> pageShift
	slot := id & (pageSize - 1)
	i.mu.Lock()
	defer i.mu.Unlock()
	if page >= uint64(len(i.pages)) {
		pages := make([][]uint64, page+1)
		copy(pages, i.pages)
		i.pages = pages
	}
	if i.pages[page] == nil {
		i.pages[page] = make([]uint64, pageSize)
	}
	i.pages[page][slot] = line + 1
}

// Get looks up id in O(1), returning its zero-based physical line.
func (i *PagedEntityIndex) Get(id uint64) (line uint64, found bool) {
	page := id >> pageShift
	slot := id & (pageSize - 1)
	i.mu.RLock()
	defer i.mu.RUnlock()
	if page >= uint64(len(i.pages)) || i.pages[page] == nil {
		return 0, false
	}
	value := i.pages[page][slot]
	return value - 1, value != 0
}

// TypeEntry is an immutable navigator record.
type TypeEntry struct {
	Name  string
	Count uint64
	Lines []uint64
}

// TypeStats accumulates compact per-type line lists during background indexing.
type TypeStats struct {
	mu      sync.RWMutex
	entries map[string]*TypeEntry
}

func (s *TypeStats) add(name string, line uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.entries[name]
	if entry == nil {
		entry = &TypeEntry{Name: name}
		s.entries[name] = entry
	}
	entry.Count++
	entry.Lines = append(entry.Lines, line)
}

// Entries returns navigator types sorted by name and safe for independent use.
func (s *TypeStats) Entries() []TypeEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]TypeEntry, 0, len(s.entries))
	for _, entry := range s.entries {
		lines := append([]uint64(nil), entry.Lines...)
		entries = append(entries, TypeEntry{Name: entry.Name, Count: entry.Count, Lines: lines})
	}
	sort.Slice(entries, func(a, b int) bool { return entries[a].Name < entries[b].Name })
	return entries
}

// Builder owns immediate line indexing and asynchronous entity/type indexing.
type Builder struct {
	doc      document.Document
	lines    *LineIndex
	entities PagedEntityIndex
	types    TypeStats
	indexed  atomic.Uint64
	ready    atomic.Bool
	done     chan struct{}
}

// NewBuilder creates the line index synchronously, then starts secondary indexing.
func NewBuilder(doc document.Document) *Builder {
	b := &Builder{
		doc:   doc,
		lines: BuildLineIndex(doc.Bytes()),
		done:  make(chan struct{}),
	}
	b.types.entries = make(map[string]*TypeEntry)
	go b.buildEntities()
	return b
}

func (b *Builder) buildEntities() {
	defer close(b.done)
	defer b.ready.Store(true)
	var parsed step.ParsedEntity
	for line := uint64(0); line < b.lines.Len(); line++ {
		start, end := b.lines.Bounds(line, b.doc.Len())
		if entity, _ := step.ParseLine(b.doc.Slice(start, end), &parsed, nil); entity {
			b.entities.Set(parsed.ID, line)
			b.types.add(parsed.Type, line)
		}
		b.indexed.Store(line + 1)
	}
}

func (b *Builder) Lines() *LineIndex { return b.lines }

func (b *Builder) Entities() *PagedEntityIndex { return &b.entities }

func (b *Builder) Types() *TypeStats { return &b.types }

func (b *Builder) IndexedLines() uint64 { return b.indexed.Load() }

func (b *Builder) Progress() float64 {
	total := b.lines.Len()
	if total == 0 {
		return 1
	}
	progress := float64(b.IndexedLines()) / float64(total)
	if progress > 1 {
		return 1
	}
	return progress
}

func (b *Builder) IsReady() bool { return b.ready.Load() }

// Wait waits for asynchronous entity indexing to finish.
func (b *Builder) Wait() { <-b.done }
