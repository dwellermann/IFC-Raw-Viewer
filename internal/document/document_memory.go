package document

// MemoryDocument exposes caller-owned immutable bytes. It is the WASM source
// and is useful for tests and embedded samples on every platform.
type MemoryDocument struct {
	data []byte
}

// NewMemoryDocument creates a document over data without copying it.
func NewMemoryDocument(data []byte) *MemoryDocument {
	return &MemoryDocument{data: data}
}

func (d *MemoryDocument) Bytes() []byte {
	return d.data
}

func (d *MemoryDocument) Len() uint64 {
	return uint64(len(d.data))
}

func (d *MemoryDocument) Slice(start, end uint64) []byte {
	return d.data[start:end]
}

func (d *MemoryDocument) LineString(start, end uint64) string {
	return lineString(d.Slice(start, end))
}

func (d *MemoryDocument) Close() error {
	return nil
}
