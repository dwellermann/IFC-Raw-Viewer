//go:build !wasm

package document

import (
	"os"

	"github.com/edsrzf/mmap-go"
)

// MMapDocument keeps an IFC file mapped read-only until Close is called.
type MMapDocument struct {
	data mmap.MMap
}

// Open maps path read-only without copying its contents to the Go heap.
func Open(path string) (*MMapDocument, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return &MMapDocument{data: data}, nil
}

func (d *MMapDocument) Bytes() []byte {
	return d.data
}

func (d *MMapDocument) Len() uint64 {
	return uint64(len(d.data))
}

func (d *MMapDocument) Slice(start, end uint64) []byte {
	return d.data[start:end]
}

func (d *MMapDocument) LineString(start, end uint64) string {
	return lineString(d.Slice(start, end))
}

func (d *MMapDocument) Close() error {
	return d.data.Unmap()
}
