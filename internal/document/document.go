// Package document provides zero-copy IFC document access.
package document

import "unsafe"

// Document exposes immutable bytes backed by a mapped file or caller-owned memory.
type Document interface {
	Bytes() []byte
	Len() uint64
	Slice(start, end uint64) []byte
	LineString(start, end uint64) string
	Close() error
}

func lineString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
