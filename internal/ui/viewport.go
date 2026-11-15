package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const viewportRowHeight float32 = 20

type virtualViewport struct {
	widget.BaseWidget
	viewer *viewer
}

func newVirtualViewport(viewer *viewer) *virtualViewport {
	viewport := &virtualViewport{viewer: viewer}
	viewport.ExtendBaseWidget(viewport)
	return viewport
}

func (v *virtualViewport) CreateRenderer() fyne.WidgetRenderer {
	return &virtualViewportRenderer{viewport: v}
}

func (v *virtualViewport) MinSize() fyne.Size {
	return fyne.NewSize(400, viewportRowHeight*8)
}

func (v *virtualViewport) Tapped(event *fyne.PointEvent) {
	row := int(event.Position.Y / viewportRowHeight)
	if row < 0 || uint64(row)+v.viewer.topLine >= v.viewer.sourceLength() {
		return
	}
	v.viewer.selectLine(v.viewer.lineAt(row), row)
}

func (v *virtualViewport) Scrolled(event *fyne.ScrollEvent) {
	lines := int64(-event.Scrolled.DY / viewportRowHeight)
	if lines == 0 && event.Scrolled.DY != 0 {
		if event.Scrolled.DY < 0 {
			lines = 3
		} else {
			lines = -3
		}
	}
	v.viewer.moveTopLine(lines)
}

type virtualViewportRenderer struct {
	viewport *virtualViewport
	rows     []*viewportRow
	objects  []fyne.CanvasObject
}

func (r *virtualViewportRenderer) Destroy() {}

func (r *virtualViewportRenderer) Layout(size fyne.Size) {
	r.ensureRows(int(size.Height/viewportRowHeight) + 2)
	for index, row := range r.rows {
		row.root.Move(fyne.NewPos(0, float32(index)*viewportRowHeight))
		row.root.Resize(fyne.NewSize(size.Width, viewportRowHeight))
		row.background.Resize(fyne.NewSize(size.Width, viewportRowHeight))
		row.number.Move(fyne.NewPos(4, 2))
		row.text.Move(fyne.NewPos(76, 0))
		row.text.Resize(fyne.NewSize(size.Width-76, viewportRowHeight))
	}
	r.Refresh()
}

func (r *virtualViewportRenderer) MinSize() fyne.Size {
	return r.viewport.MinSize()
}

func (r *virtualViewportRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *virtualViewportRenderer) Refresh() {
	r.ensureRows(int(r.viewport.Size().Height/viewportRowHeight) + 2)
	for rowIndex, row := range r.rows {
		position := r.viewport.viewer.topLine + uint64(rowIndex)
		if position >= r.viewport.viewer.sourceLength() {
			row.root.Hide()
			continue
		}
		row.root.Show()
		r.updateRow(row, r.viewport.viewer.lineAt(rowIndex))
	}
}

func (r *virtualViewportRenderer) ensureRows(count int) {
	if count < 1 {
		count = 1
	}
	for len(r.rows) < count {
		row := newViewportRow()
		r.rows = append(r.rows, row)
		r.objects = append(r.objects, row.root)
	}
}

func (r *virtualViewportRenderer) updateRow(row *viewportRow, line uint64) {
	if r.viewport.viewer.hasSelection && line == r.viewport.viewer.selectedLine {
		row.background.FillColor = selectionColor
	} else {
		row.background.FillColor = color.Transparent
	}
	canvas.Refresh(row.background)
	row.number.Text = fmt.Sprintf("%7d", line+1)
	canvas.Refresh(row.number)

	start, end := r.viewport.viewer.builder.Lines().Bounds(line, r.viewport.viewer.doc.Len())
	bytes := r.viewport.viewer.doc.Slice(start, end)
	for len(bytes) > 0 && (bytes[len(bytes)-1] == '\n' || bytes[len(bytes)-1] == '\r') {
		bytes = bytes[:len(bytes)-1]
	}
	truncated := len(bytes) > linePreviewBytes
	if truncated {
		bytes = bytes[:linePreviewBytes]
	}
	row.text.Objects = syntaxText(bytes)
	if truncated {
		row.text.Objects = append(row.text.Objects, lineText(" ...", mutedColor))
	}
	row.text.Refresh()
}

type viewportRow struct {
	root       *fyne.Container
	background *canvas.Rectangle
	number     *canvas.Text
	text       *fyne.Container
}

func newViewportRow() *viewportRow {
	background := canvas.NewRectangle(color.Transparent)
	number := lineText("", mutedColor)
	text := container.NewHBox()
	return &viewportRow{
		root:       container.NewWithoutLayout(background, number, text),
		background: background,
		number:     number,
		text:       text,
	}
}
