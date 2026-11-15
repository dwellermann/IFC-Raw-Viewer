package step

import (
	"testing"
	"unsafe"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		entity     bool
		id         uint64
		typ        string
		globalID   string
		references []uint64
	}{
		{"normal", "#12345=IFCWALL('2Xabc...',#42,#12346,$);", true, 12345, "IFCWALL", "2Xabc...", []uint64{42, 12346}},
		{"whitespace and CR", " \t#7 = IFCSPACE('guid',#1);\r", true, 7, "IFCSPACE", "guid", []uint64{1}},
		{"hash string", "#3=IFCLABEL('not #42',#2);", true, 3, "IFCLABEL", "not #42", []uint64{2}},
		{"missing global id", "#4=IFCOWNERHISTORY($,#2);", true, 4, "IFCOWNERHISTORY", "", []uint64{2}},
		{"escaped quote", "#5=IFCLABEL('it''s #not-a-ref',#8);", true, 5, "IFCLABEL", "it''s #not-a-ref", []uint64{8}},
		{"header", "HEADER;", false, 0, "", "", nil},
		{"empty", "", false, 0, "", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsed ParsedEntity
			gotEntity, refs := ParseLine([]byte(tt.line), &parsed, make([]uint64, 0, 8))
			if gotEntity != tt.entity || parsed.ID != tt.id || parsed.Type != tt.typ || parsed.GlobalID != tt.globalID {
				t.Fatalf("parsed = %#v, isEntity = %v", parsed, gotEntity)
			}
			if len(refs) != len(tt.references) {
				t.Fatalf("refs = %v, want %v", refs, tt.references)
			}
			for i := range refs {
				if refs[i] != tt.references[i] {
					t.Fatalf("refs = %v, want %v", refs, tt.references)
				}
			}
		})
	}
}

func TestParseLineUsesSourceBytes(t *testing.T) {
	line := []byte("#1=IFCWALL('guid');")
	var parsed ParsedEntity
	ParseLine(line, &parsed, nil)
	if unsafe.StringData(parsed.Type) != unsafe.SliceData(line[3:]) {
		t.Fatal("type was copied")
	}
}

func TestTokenSpans(t *testing.T) {
	spans := TokenSpans([]byte("#1=IFCWALL('a',#2);"), make([]Span, 0, 8))
	if len(spans) != 4 {
		t.Fatalf("span count = %d, want 4: %#v", len(spans), spans)
	}
	if spans[0].Kind != SpanID || spans[1].Kind != SpanType || spans[2].Kind != SpanString || spans[3].Kind != SpanReference {
		t.Fatalf("unexpected spans: %#v", spans)
	}
}

func BenchmarkParseLine(b *testing.B) {
	line := []byte("#12345=IFCWALL('2Xabc...',#42,#12346,$);")
	var parsed ParsedEntity
	refs := make([]uint64, 0, 8)
	b.ReportAllocs()
	for b.Loop() {
		ParseLine(line, &parsed, refs)
	}
}
