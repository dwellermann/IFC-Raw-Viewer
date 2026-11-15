// Package step parses STEP physical-file records without regular expressions.
package step

import "unsafe"

// ParsedEntity describes an entity using views into its source line.
type ParsedEntity struct {
	ID         uint64
	Type       string
	GlobalID   string
	TypeStart  uint32
	TypeEnd    uint32
	RefOffsets []uint64
	Refs       []uint64
}

// SpanKind identifies syntax highlighted source portions.
type SpanKind uint8

const (
	SpanID SpanKind = iota
	SpanType
	SpanString
	SpanReference
	SpanKeyword
)

// Span is a byte range in a physical STEP line.
type Span struct {
	Start uint32
	End   uint32
	Kind  SpanKind
}

// ParseLine parses a single physical line. refBuf is reused as the result
// buffer; callers should reserve it for the expected number of references.
func ParseLine(line []byte, out *ParsedEntity, refBuf []uint64) (isEntity bool, refs []uint64) {
	*out = ParsedEntity{Refs: refBuf[:0], RefOffsets: refBuf[:0]}
	i := skipSpace(line, 0)
	if i >= len(line) || line[i] != '#' {
		return false, refBuf[:0]
	}
	i++
	id, next, ok := number(line, i)
	if !ok {
		return false, refBuf[:0]
	}
	i = skipSpace(line, next)
	if i >= len(line) || line[i] != '=' {
		return false, refBuf[:0]
	}
	i = skipSpace(line, i+1)
	typeStart := i
	for i < len(line) && isTypeChar(line[i]) {
		i++
	}
	if i == typeStart {
		return false, refBuf[:0]
	}
	typeEnd := i
	i = skipSpace(line, i)
	if i >= len(line) || line[i] != '(' {
		return false, refBuf[:0]
	}

	out.ID = id
	out.TypeStart = uint32(typeStart)
	out.TypeEnd = uint32(typeEnd)
	out.Type = byteString(line[typeStart:typeEnd])

	inString := false
	hasGlobalID := false
	for i++; i < len(line); i++ {
		switch line[i] {
		case '\'':
			if inString && i+1 < len(line) && line[i+1] == '\'' {
				i++
				continue
			}
			if !inString && !hasGlobalID {
				start := i + 1
				i++
				for i < len(line) {
					if line[i] != '\'' {
						i++
						continue
					}
					if i+1 < len(line) && line[i+1] == '\'' {
						i += 2
						continue
					}
					out.GlobalID = byteString(line[start:i])
					hasGlobalID = true
					break
				}
				if i >= len(line) {
					return true, out.Refs
				}
				continue
			}
			inString = !inString
		case '#':
			if inString {
				continue
			}
			ref, end, valid := number(line, i+1)
			if valid {
				out.Refs = append(out.Refs, ref)
				i = end - 1
			}
		}
	}
	out.RefOffsets = out.Refs
	return true, out.Refs
}

// TokenSpans appends spans for a visible source line to out without allocating.
func TokenSpans(line []byte, out []Span) []Span {
	out = out[:0]
	i := skipSpace(line, 0)
	scanStart := 0
	if i >= len(line) {
		return out
	}
	if line[i] == '#' {
		start := i
		if _, end, ok := number(line, i+1); ok {
			out = append(out, Span{uint32(start), uint32(end), SpanID})
			i = skipSpace(line, end)
			if i < len(line) && line[i] == '=' {
				i = skipSpace(line, i+1)
				start = i
				for i < len(line) && isTypeChar(line[i]) {
					i++
				}
				if i > start {
					out = append(out, Span{uint32(start), uint32(i), SpanType})
				}
			}
			scanStart = i
		}
	} else if isTypeChar(line[i]) {
		start := i
		for i < len(line) && isTypeChar(line[i]) {
			i++
		}
		out = append(out, Span{uint32(start), uint32(i), SpanKeyword})
	}

	inString := false
	for i = scanStart; i < len(line); i++ {
		switch line[i] {
		case '\'':
			if !inString {
				start := i
				i++
				for i < len(line) {
					if line[i] != '\'' {
						i++
						continue
					}
					if i+1 < len(line) && line[i+1] == '\'' {
						i += 2
						continue
					}
					out = append(out, Span{uint32(start), uint32(i + 1), SpanString})
					break
				}
				inString = false
			} else {
				inString = false
			}
		case '#':
			if inString {
				continue
			}
			if _, end, ok := number(line, i+1); ok {
				out = append(out, Span{uint32(i), uint32(end), SpanReference})
				i = end - 1
			}
		}
	}
	return out
}

func skipSpace(line []byte, i int) int {
	for i < len(line) && (line[i] == ' ' || line[i] == '\t' || line[i] == '\r') {
		i++
	}
	return i
}

func number(line []byte, i int) (value uint64, end int, ok bool) {
	start := i
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		value = value*10 + uint64(line[i]-'0')
		i++
	}
	return value, i, i != start
}

func isTypeChar(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '_' || (b >= '0' && b <= '9')
}

func byteString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
