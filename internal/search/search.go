// Package search provides allocation-conscious STEP document navigation.
package search

import (
	"stepview/internal/document"
	"stepview/internal/index"
)

// Service searches a document using its compact background indexes.
type Service struct {
	doc     document.Document
	builder *index.Builder
}

func New(doc document.Document, builder *index.Builder) *Service {
	return &Service{doc: doc, builder: builder}
}

// JumpToEntity resolves an entity ID in O(1).
func (s *Service) JumpToEntity(id uint64) (line uint64, found bool) {
	return s.builder.Entities().Get(id)
}

// Search finds physical lines containing query. A #ID prefix uses O(1) lookup.
func (s *Service) Search(query string, limit int) []uint64 {
	if limit <= 0 || query == "" {
		return nil
	}
	if id, ok := entityIDQuery(query); ok {
		line, found := s.JumpToEntity(id)
		if !found {
			return nil
		}
		return []uint64{line}
	}

	results := make([]uint64, 0, min(limit, 16))
	lines := s.builder.Lines()
	for line := uint64(0); line < lines.Len() && len(results) < limit; line++ {
		start, end := lines.Bounds(line, s.doc.Len())
		if containsFoldASCII(s.doc.Slice(start, end), query) {
			results = append(results, line)
		}
	}
	return results
}

// FindReferrers returns entities containing a real (non-string) #targetID reference.
func (s *Service) FindReferrers(targetID uint64, limit int) []uint64 {
	if limit <= 0 {
		return nil
	}
	var digits [20]byte
	target := appendUint(digits[:0], targetID)
	results := make([]uint64, 0, min(limit, 16))
	lines := s.builder.Lines()
	for line := uint64(0); line < lines.Len() && len(results) < limit; line++ {
		start, end := lines.Bounds(line, s.doc.Len())
		if hasReference(s.doc.Slice(start, end), target) {
			results = append(results, line)
		}
	}
	return results
}

func entityIDQuery(query string) (uint64, bool) {
	if len(query) < 2 || query[0] != '#' {
		return 0, false
	}
	var id uint64
	for i := 1; i < len(query); i++ {
		if query[i] < '0' || query[i] > '9' {
			return 0, false
		}
		id = id*10 + uint64(query[i]-'0')
	}
	return id, true
}

func containsFoldASCII(line []byte, query string) bool {
	if len(query) > len(line) {
		return false
	}
	for start := 0; start+len(query) <= len(line); start++ {
		match := true
		for i := 0; i < len(query); i++ {
			if foldASCII(line[start+i]) != foldASCII(query[i]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func foldASCII(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

func hasReference(line, target []byte) bool {
	inString := false
	start := 0
	for start < len(line) && line[start] != '(' {
		start++
	}
	if start == len(line) {
		return false
	}
	for i := start + 1; i < len(line); i++ {
		if line[i] == '\'' {
			if inString && i+1 < len(line) && line[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
			continue
		}
		if inString || line[i] != '#' || i+1+len(target) > len(line) {
			continue
		}
		match := true
		for j := range target {
			if line[i+1+j] != target[j] {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		end := i + 1 + len(target)
		if end == len(line) || line[end] < '0' || line[end] > '9' {
			return true
		}
	}
	return false
}

func appendUint(dst []byte, value uint64) []byte {
	var reverse [20]byte
	count := 1
	reverse[0] = byte(value%10) + '0'
	for value /= 10; value > 0; value /= 10 {
		reverse[count] = byte(value%10) + '0'
		count++
	}
	for count > 0 {
		count--
		dst = append(dst, reverse[count])
	}
	return dst
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
