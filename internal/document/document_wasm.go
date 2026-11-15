//go:build wasm

package document

import "fmt"

// Open is unavailable in browsers; use NewMemoryDocument with uploaded bytes.
func Open(path string) (*MemoryDocument, error) {
	return nil, fmt.Errorf("opening local paths is unavailable on wasm")
}
