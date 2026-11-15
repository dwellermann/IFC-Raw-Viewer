package ui

import (
	"image/color"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"stepview/internal/document"
)

func TestVirtualViewportMapsRowsToSourceLines(t *testing.T) {
	lines := []uint64{11, 42, 84, 126}
	viewer := viewer{
		activeLines: lines,
		topLine:     2,
	}

	if got, want := viewer.lineAt(1), lines[3]; got != want {
		t.Fatalf("mapped line = %d, want %d", got, want)
	}
}

func TestMaxTopLineUsesSourceLength(t *testing.T) {
	viewer := viewer{activeLines: make([]uint64, 500_001)}
	if got := viewer.maxTopLine(); got != 500_000 {
		t.Fatalf("max top line = %d, want 500000", got)
	}
}

func TestWellermannLightThemeUsesLightBackground(t *testing.T) {
	light := newWellermannTheme(0.82, true)
	got := color.NRGBAModel.Convert(light.Color(theme.ColorNameBackground, theme.VariantLight)).(color.NRGBA)
	want := color.NRGBA{R: 246, G: 245, B: 239, A: 255}
	if got != want {
		t.Fatalf("light background = %#v, want %#v", got, want)
	}
}

func TestLoadKeepsPreviousDocumentOpenUntilItsIndexerFinishes(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()

	previous := newBlockingDocument([]byte("#1=IFCWALL('old',$,$,$,$,$,$,$);\n"))
	viewer := newViewer(application, application.NewWindow("test"), previous, "old.ifc")
	select {
	case <-previous.started:
	case <-time.After(time.Second):
		t.Fatal("previous indexer did not start")
	}

	viewer.load(document.NewMemoryDocument([]byte("#2=IFCWALL('new',$,$,$,$,$,$,$);\n")), "new.ifc")
	select {
	case <-previous.closed:
		t.Fatal("previous document was closed while its indexer was running")
	default:
	}

	close(previous.release)
	select {
	case <-previous.closed:
	case <-time.After(time.Second):
		t.Fatal("previous document was not closed after its indexer finished")
	}
}

type blockingDocument struct {
	data      []byte
	started   chan struct{}
	release   chan struct{}
	closed    chan struct{}
	sliceOnce sync.Once
	closeOnce sync.Once
}

func newBlockingDocument(data []byte) *blockingDocument {
	return &blockingDocument{
		data:    data,
		started: make(chan struct{}),
		release: make(chan struct{}),
		closed:  make(chan struct{}),
	}
}

func (d *blockingDocument) Bytes() []byte { return d.data }

func (d *blockingDocument) Len() uint64 { return uint64(len(d.data)) }

func (d *blockingDocument) Slice(start, end uint64) []byte {
	d.sliceOnce.Do(func() {
		close(d.started)
		<-d.release
	})
	return d.data[start:end]
}

func (d *blockingDocument) LineString(start, end uint64) string {
	return string(d.data[start:end])
}

func (d *blockingDocument) Close() error {
	d.closeOnce.Do(func() { close(d.closed) })
	return nil
}
