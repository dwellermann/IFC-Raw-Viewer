package search

import (
	"testing"

	"stepview/internal/document"
	"stepview/internal/index"
)

func TestSearchAndNavigation(t *testing.T) {
	doc := document.NewMemoryDocument([]byte(
		"#1=IFCPROJECT('Project Alpha',#3);\n" +
			"#42=IFCWALL('2Xabc',#1,#100);\n" +
			"#100=IFCLOCALPLACEMENT($);\n" +
			"#101=IFCLABEL('string #42 is not a reference');\n"))
	builder := index.NewBuilder(doc)
	builder.Wait()
	search := New(doc, builder)

	line, found := search.JumpToEntity(42)
	if !found || line != 1 {
		t.Fatalf("JumpToEntity = (%d, %v), want (1, true)", line, found)
	}
	if got := search.Search("#100", 10); len(got) != 1 || got[0] != 2 {
		t.Fatalf("ID search = %v", got)
	}
	if got := search.Search("pRoJeCt aLpHa", 10); len(got) != 1 || got[0] != 0 {
		t.Fatalf("text search = %v", got)
	}
	if got := search.FindReferrers(1, 10); len(got) != 1 || got[0] != 1 {
		t.Fatalf("referrers for #1 = %v", got)
	}
	if got := search.FindReferrers(42, 10); len(got) != 0 {
		t.Fatalf("string reference was included: %v", got)
	}
}
