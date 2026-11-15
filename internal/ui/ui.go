// Package ui provides the Fyne frontend for STEPView.
package ui

import (
	"fmt"
	"image/color"
	"net/url"
	"path/filepath"
	"strings"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"stepview/internal/document"
	"stepview/internal/index"
	"stepview/internal/search"
	"stepview/internal/step"
)

// Run starts the native Fyne application.
func Run(doc document.Document, name string) {
	application := app.NewWithID(ProductAppID)
	setSyntaxColors(false)
	application.Settings().SetTheme(newWellermannTheme(0.82, false))
	window := application.NewWindow(ProductTitle)
	window.Resize(fyne.NewSize(1440, 900))

	viewer := newViewer(application, window, doc, name)
	window.SetMainMenu(viewer.menu())
	window.SetContent(viewer.layout())
	window.ShowAndRun()
}

type viewer struct {
	application fyne.App
	window      fyne.Window
	doc         document.Document
	builder     *index.Builder
	search      *search.Service
	name        string

	searchEntry   *widget.Entry
	filterEntry   *widget.Entry
	status        *widget.Label
	typeList      *widget.List
	lineViewport  *virtualViewport
	lineSlider    *widget.Slider
	inspector     *fyne.Container
	rawLineAction *widget.Button

	types         []index.TypeEntry
	filteredType  []index.TypeEntry
	activeLines   []uint64
	topLine       uint64
	selectedLine  uint64
	selectedRow   int
	hasSelection  bool
	density       float32
	lightMode     bool
	syncingScroll bool
}

const (
	linePreviewBytes   = 2 * 1024
	rawEditorPageBytes = 16 * 1024
)

func newViewer(application fyne.App, window fyne.Window, doc document.Document, name string) *viewer {
	v := &viewer{
		application: application,
		window:      window,
		density:     0.82,
		status:      widget.NewLabel("Index wird aufgebaut …"),
		inspector: container.NewVBox(
			widget.NewLabelWithStyle("Inspector", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Zeile auswählen"),
		),
	}
	v.rawLineAction = widget.NewButton("Raw-Editor öffnen", func() {
		if v.hasSelection {
			v.showRawLineEditor(v.selectedLine)
		}
	})
	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("#ID, GlobalId oder Volltext")
	v.searchEntry.OnSubmitted = func(string) { v.performSearch() }
	v.filterEntry = widget.NewEntry()
	v.filterEntry.SetPlaceHolder("IFC-Typ filtern")
	v.filterEntry.OnChanged = func(string) { v.refreshTypeFilter() }
	v.typeList = widget.NewList(
		func() int { return len(v.filteredType) },
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			entry := v.filteredType[id]
			object.(*widget.Label).SetText(fmt.Sprintf("%s (%d)", entry.Name, entry.Count))
		},
	)
	v.typeList.OnSelected = func(id widget.ListItemID) {
		v.clearSelection()
		v.activeLines = v.filteredType[id].Lines
		v.setTopLine(0)
	}

	v.lineViewport = newVirtualViewport(v)
	v.lineSlider = widget.NewSlider(0, 1)
	v.lineSlider.Orientation = widget.Vertical
	v.lineSlider.OnChanged = func(value float64) {
		if v.syncingScroll {
			return
		}
		v.setTopLine(uint64(value * float64(v.maxTopLine())))
	}
	v.load(doc, name)
	return v
}

func (v *viewer) layout() fyne.CanvasObject {
	brand := widget.NewLabelWithStyle(ProductName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true})
	open := widget.NewButton("Datei öffnen", v.openFile)
	searchButton := widget.NewButton("Suchen", v.performSearch)
	allLines := widget.NewButton("Alle Zeilen", v.showAllLines)
	toolbar := container.NewBorder(nil, nil, brand, container.NewHBox(searchButton, allLines),
		container.NewBorder(nil, nil, open, nil, v.searchEntry))

	left := container.NewBorder(
		container.NewVBox(widget.NewLabelWithStyle("IFC-Typen", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), v.filterEntry, allLines),
		nil, nil, nil, v.typeList,
	)
	center := container.NewBorder(
		widget.NewLabelWithStyle("STEP / Entity-Explorer", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, v.lineSlider, v.lineViewport,
	)
	right := container.NewVScroll(v.inspector)
	leftCenter := container.NewHSplit(left, center)
	leftCenter.Offset = 0.22
	content := container.NewHSplit(leftCenter, right)
	content.Offset = 0.74
	footer := container.NewBorder(nil, nil, nil, widget.NewLabel(ProductMark), v.status)
	return container.NewBorder(toolbar, footer, nil, nil, content)
}

func (v *viewer) menu() *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("Datei",
			fyne.NewMenuItem("IFC öffnen …", v.openFile),
		),
		fyne.NewMenu("Ansicht",
			fyne.NewMenuItem("Alle Zeilen", v.showAllLines),
			fyne.NewMenuItem("Suche fokussieren", func() { v.window.Canvas().Focus(v.searchEntry) }),
		),
		fyne.NewMenu("Einstellungen",
			fyne.NewMenuItem("Kompakte Zeilen", func() { v.setDensity(0.72) }),
			fyne.NewMenuItem("Standard-Zeilengröße", func() { v.setDensity(0.92) }),
			fyne.NewMenuItem("Große Zeilen", func() { v.setDensity(1.12) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Dunkler Modus", func() { v.setLightMode(false) }),
			fyne.NewMenuItem("Light Mode", func() { v.setLightMode(true) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Über "+ProductName, func() {
				v.showAbout()
			}),
		),
	)
}

func (v *viewer) showAbout() {
	website, _ := url.Parse("https://wellermann.de")
	content := container.NewVBox(
		widget.NewLabelWithStyle(ProductName+" / "+ProductMark, fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Monospace: true}),
		widget.NewLabel(ProductCredit),
		widget.NewLabel("Read-only IFC-SPF Daten-Explorer und Debugger"),
		widget.NewLabel("Mmap · Paged Index · Streaming Search"),
		widget.NewHyperlink("wellermann.de", website),
	)
	dialog.ShowCustom(ProductTitle, "Schließen", content, v.window)
}

func (v *viewer) setDensity(density float32) {
	v.density = density
	v.application.Settings().SetTheme(newWellermannTheme(density, v.lightMode))
	v.lineViewport.Refresh()
}

func (v *viewer) setLightMode(light bool) {
	v.lightMode = light
	setSyntaxColors(light)
	v.application.Settings().SetTheme(newWellermannTheme(v.density, light))
	v.lineViewport.Refresh()
	v.inspector.Refresh()
}

func (v *viewer) openFile() {
	path, err := SelectIFCFile()
	if err != nil {
		v.status.SetText(err.Error())
		return
	}
	if path == "" {
		return
	}
	doc, err := document.Open(path)
	if err != nil {
		v.status.SetText(err.Error())
		return
	}
	v.load(doc, filepath.Base(path))
}

func (v *viewer) load(doc document.Document, name string) {
	previousDoc := v.doc
	previousBuilder := v.builder
	v.doc = doc
	v.name = name
	v.builder = index.NewBuilder(doc)
	v.search = search.New(doc, v.builder)
	builder := v.builder

	if previousDoc != nil && previousDoc != doc {
		if previousBuilder == nil {
			_ = previousDoc.Close()
		} else {
			go func(doc document.Document, builder *index.Builder) {
				builder.Wait()
				_ = doc.Close()
			}(previousDoc, previousBuilder)
		}
	}

	v.types = nil
	v.filteredType = nil
	v.activeLines = nil
	v.topLine = 0
	v.hasSelection = false
	v.lineViewport.Refresh()
	v.updateScrollBar()
	v.status.SetText(name + " · Index wird aufgebaut …")

	go func() {
		builder.Wait()
		fyne.Do(func() {
			if v.builder != builder || v.doc != doc {
				return
			}
			v.types = builder.Types().Entries()
			v.refreshTypeFilter()
			v.lineViewport.Refresh()
			v.updateScrollBar()
			v.status.SetText(fmt.Sprintf("%s · %d Zeilen indiziert", v.name, builder.Lines().Len()))
		})
	}()
}

func (v *viewer) refreshTypeFilter() {
	query := strings.ToLower(v.filterEntry.Text)
	v.filteredType = v.filteredType[:0]
	for _, entry := range v.types {
		if query == "" || strings.Contains(strings.ToLower(entry.Name), query) {
			v.filteredType = append(v.filteredType, entry)
		}
	}
	v.typeList.Refresh()
}

func (v *viewer) performSearch() {
	v.clearSelection()
	v.activeLines = v.search.Search(v.searchEntry.Text, 10_000)
	v.setTopLine(0)
	if len(v.activeLines) == 0 {
		v.status.SetText("Keine Treffer")
		return
	}
	v.status.SetText(fmt.Sprintf("%d Treffer", len(v.activeLines)))
	v.selectLine(v.activeLines[0], 0)
}

func (v *viewer) showAllLines() {
	v.clearSelection()
	v.activeLines = nil
	v.setTopLine(0)
}

func (v *viewer) lineAt(id int) uint64 {
	position := v.topLine + uint64(id)
	if v.activeLines != nil {
		return v.activeLines[position]
	}
	return position
}

func (v *viewer) clearSelection() {
	v.hasSelection = false
}

func (v *viewer) sourceLength() uint64 {
	if v.activeLines != nil {
		return uint64(len(v.activeLines))
	}
	if v.builder == nil {
		return 0
	}
	return v.builder.Lines().Len()
}

func (v *viewer) maxTopLine() uint64 {
	total := v.sourceLength()
	if total == 0 {
		return 0
	}
	return total - 1
}

func (v *viewer) setTopLine(start uint64) {
	if start > v.maxTopLine() {
		start = v.maxTopLine()
	}
	v.clearSelection()
	v.topLine = start
	v.lineViewport.Refresh()
	v.updateScrollBar()
}

func (v *viewer) moveTopLine(delta int64) {
	if delta == 0 {
		return
	}
	if delta < 0 {
		delta = -delta
		if uint64(delta) >= v.topLine {
			v.setTopLine(0)
			return
		}
		v.setTopLine(v.topLine - uint64(delta))
		return
	}
	v.setTopLine(v.topLine + uint64(delta))
}

func (v *viewer) updateScrollBar() {
	maximum := v.maxTopLine()
	v.syncingScroll = true
	if maximum == 0 {
		v.lineSlider.Disable()
		v.lineSlider.SetValue(0)
	} else {
		v.lineSlider.Enable()
		v.lineSlider.SetValue(float64(v.topLine) / float64(maximum))
	}
	v.syncingScroll = false
}

func (v *viewer) refreshLineRow(row int) {
	v.lineViewport.Refresh()
}

func (v *viewer) selectLine(line uint64, row int) {
	previousRow := v.selectedRow
	hadSelection := v.hasSelection
	v.selectedLine = line
	v.selectedRow = row
	v.hasSelection = true
	if hadSelection && previousRow != row {
		v.refreshLineRow(previousRow)
	}
	v.refreshLineRow(row)
	start, end := v.builder.Lines().Bounds(line, v.doc.Len())
	var parsed step.ParsedEntity
	entity, refs := step.ParseLine(v.doc.Slice(start, end), &parsed, make([]uint64, 0, 16))
	if !entity {
		v.inspector.Objects = []fyne.CanvasObject{
			widget.NewLabelWithStyle("Inspector", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Keine IFC-Entity"),
			v.rawLineAction,
		}
		v.inspector.Refresh()
		return
	}

	objects := []fyne.CanvasObject{
		widget.NewLabelWithStyle(fmt.Sprintf("#%d  %s", parsed.ID, parsed.Type), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("GlobalId: " + parsed.GlobalID),
		v.rawLineAction,
		widget.NewButton("Zeile kopieren", func() {
			v.window.Clipboard().SetContent(string(v.doc.Slice(start, end)))
		}),
		widget.NewButton("Referenzen suchen", func() {
			v.showReferrers(parsed.ID)
		}),
		widget.NewLabelWithStyle("Ausgehende Referenzen", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for _, id := range refs {
		targetLine, found := v.search.JumpToEntity(id)
		if !found {
			continue
		}
		targetStart, targetEnd := v.builder.Lines().Bounds(targetLine, v.doc.Len())
		var target step.ParsedEntity
		step.ParseLine(v.doc.Slice(targetStart, targetEnd), &target, nil)
		id, targetLine, targetType := id, targetLine, target.Type
		objects = append(objects, widget.NewButton(fmt.Sprintf("#%d → %s", id, targetType), func() {
			v.jumpToLine(targetLine)
		}))
	}
	v.inspector.Objects = objects
	v.inspector.Refresh()
}

func (v *viewer) showRawLineEditor(line uint64) {
	start, end := v.builder.Lines().Bounds(line, v.doc.Len())
	for end > start {
		last := v.doc.Slice(end-1, end)[0]
		if last != '\n' && last != '\r' {
			break
		}
		end--
	}

	total := end - start
	page := uint64(0)
	pageSize := uint64(rawEditorPageBytes)
	content := lineText("", codeColor)
	content.TextSize = 14
	rawView := container.NewScroll(content)
	rawView.SetMinSize(fyne.NewSize(760, 280))
	position := widget.NewLabel("")
	previous := widget.NewButton("Zurück", nil)
	next := widget.NewButton("Weiter", nil)
	copyChunk := widget.NewButton("Abschnitt kopieren", func() {
		chunkStart := start + page*pageSize
		chunkEnd := chunkStart + pageSize
		if chunkEnd > end {
			chunkEnd = end
		}
		v.window.Clipboard().SetContent(string(v.doc.Slice(chunkStart, chunkEnd)))
	})

	update := func() {
		chunkStart := start + page*pageSize
		chunkEnd := chunkStart + pageSize
		if chunkEnd > end {
			chunkEnd = end
		}
		content.Text = v.doc.LineString(chunkStart, chunkEnd)
		canvas.Refresh(content)
		position.SetText(fmt.Sprintf("Read-only · Zeile %d · Bytes %d-%d von %d", line+1, chunkStart-start+1, chunkEnd-start, total))
		if page == 0 {
			previous.Disable()
		} else {
			previous.Enable()
		}
		if chunkEnd == end {
			next.Disable()
		} else {
			next.Enable()
		}
	}
	previous.OnTapped = func() {
		page--
		update()
	}
	next.OnTapped = func() {
		page++
		update()
	}
	update()

	var navigation fyne.CanvasObject
	if total > pageSize {
		navigation = container.NewHBox(previous, next, copyChunk)
	} else {
		navigation = container.NewHBox(copyChunk)
	}
	body := container.NewBorder(position, navigation, nil, nil, rawView)
	rawDialog := dialog.NewCustom(fmt.Sprintf("Raw-Editor · Zeile %d", line+1), "Schließen", body, v.window)
	rawDialog.Resize(fyne.NewSize(820, 420))
	rawDialog.Show()
}

func (v *viewer) showReferrers(id uint64) {
	lines := v.search.FindReferrers(id, 200)
	if len(lines) == 0 {
		v.status.SetText("Keine eingehenden Referenzen")
		return
	}
	v.clearSelection()
	v.activeLines = lines
	v.setTopLine(0)
	v.status.SetText(fmt.Sprintf("%d eingehende Referenzen", len(lines)))
}

func (v *viewer) jumpToLine(line uint64) {
	position := uint64(0)
	sourceChanged := false
	if v.activeLines == nil {
		position = line
	} else {
		found := false
		for i, candidate := range v.activeLines {
			if candidate == line {
				position = uint64(i)
				found = true
				break
			}
		}
		if !found {
			v.activeLines = nil
			position = line
			sourceChanged = true
		}
	}
	if sourceChanged || v.topLine != position {
		v.setTopLine(position)
	}
	v.selectLine(line, 0)
}

var (
	mutedColor      color.Color = color.NRGBA{R: 161, G: 160, B: 151, A: 255}
	entityIDColor   color.Color = color.NRGBA{R: 83, G: 182, B: 232, A: 255}
	entityTypeColor color.Color = color.NRGBA{R: 255, G: 177, B: 55, A: 255}
	stringColor     color.Color = color.NRGBA{R: 122, G: 218, B: 157, A: 255}
	codeColor       color.Color = color.NRGBA{R: 245, G: 242, B: 234, A: 255}
	selectionColor  color.Color = color.NRGBA{R: 44, G: 84, B: 120, A: 255}
)

func setSyntaxColors(light bool) {
	if light {
		mutedColor = color.NRGBA{R: 104, G: 101, B: 91, A: 255}
		entityIDColor = color.NRGBA{R: 0, G: 95, B: 145, A: 255}
		entityTypeColor = color.NRGBA{R: 150, G: 77, B: 0, A: 255}
		stringColor = color.NRGBA{R: 27, G: 112, B: 64, A: 255}
		codeColor = color.NRGBA{R: 37, G: 36, B: 32, A: 255}
		selectionColor = color.NRGBA{R: 193, G: 218, B: 236, A: 255}
		return
	}
	mutedColor = color.NRGBA{R: 161, G: 160, B: 151, A: 255}
	entityIDColor = color.NRGBA{R: 83, G: 182, B: 232, A: 255}
	entityTypeColor = color.NRGBA{R: 255, G: 177, B: 55, A: 255}
	stringColor = color.NRGBA{R: 122, G: 218, B: 157, A: 255}
	codeColor = color.NRGBA{R: 245, G: 242, B: 234, A: 255}
	selectionColor = color.NRGBA{R: 44, G: 84, B: 120, A: 255}
}

func syntaxText(line []byte) []fyne.CanvasObject {
	spans := step.TokenSpans(line, make([]step.Span, 0, 16))
	texts := make([]fyne.CanvasObject, 0, len(spans)*2+1)
	cursor := 0
	for _, span := range spans {
		if int(span.Start) > cursor {
			texts = append(texts, lineText(unsafe.String(unsafe.SliceData(line[cursor:span.Start]), int(span.Start)-cursor), codeColor))
		}
		texts = append(texts, lineText(unsafe.String(unsafe.SliceData(line[span.Start:span.End]), int(span.End-span.Start)), canvasTokenColor(span.Kind)))
		cursor = int(span.End)
	}
	if cursor < len(line) {
		texts = append(texts, lineText(unsafe.String(unsafe.SliceData(line[cursor:]), len(line)-cursor), codeColor))
	}
	return texts
}

func lineText(text string, textColor color.Color) *canvas.Text {
	label := canvas.NewText(text, textColor)
	label.TextSize = 12
	label.TextStyle = fyne.TextStyle{Monospace: true}
	return label
}

func canvasTokenColor(kind step.SpanKind) color.Color {
	switch kind {
	case step.SpanID, step.SpanReference:
		return entityIDColor
	case step.SpanType, step.SpanKeyword:
		return entityTypeColor
	case step.SpanString:
		return stringColor
	default:
		return codeColor
	}
}
