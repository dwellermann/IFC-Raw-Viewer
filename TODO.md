# STEPView Todo

## Phase 1: Projekt-Setup & Manifest
- [x] `go mod init stepview` ausführen (mindestens Go 1.22+).
- [x] Abhängigkeiten installieren: `go get github.com/edsrzf/mmap-go` und `go get gioui.org@latest`.
- [x] `ARCHITECTURE.md` und `TODO.md` sowie die Ordnerstruktur unter `cmd/` und `internal/` anlegen.
- [x] Minimales compilierbares `cmd/stepview/main.go` erstellen.
- [x] **Gate:** Native + WASM Build erfolgreich.

## Phase 2: Document-Layer (`internal/document`)
- [x] Interface `Document` definieren: `Bytes() []byte`, `Len() uint64`, `Slice(start, end uint64) []byte`, `LineString(start, end uint64) string` (via `unsafe.String`), `Close() error`.
- [x] Desktop-Implementierung (`document_mmap.go` mit `//go:build !wasm`) mit `edsrzf/mmap-go` (`mmap.RDONLY`).
- [x] WASM-/Memory-Implementierung (`document_mem.go` mit `//go:build wasm`, plus nutzbar für Unit-Tests als `NewMemoryDocument([]byte)`).
- [x] Unit-Tests schreiben (Öffnen, Slicing, `unsafe.String` ohne Kopie, sauberes `Close()`).
- [x] **Gate:** `go test`, `go vet`, Native Build, WASM Build grün.

## Phase 3: Zero-Allocation STEP-Tokenizer (`internal/step`)
- [x] Struct `ParsedEntity` definieren: `ID uint64`, `Type string` (Zero-Copy via `unsafe.String`), `GlobalId string` (falls erstes String-Argument in `'...'` vorhanden), `TypeStart/TypeEnd`, `RefOffsets` / `Refs []uint64`.
- [x] Handgeschriebenen FSM-Byte-Scanner `ParseLine(line []byte, out *ParsedEntity, refBuf []uint64) (isEntity bool, refs []uint64)` implementieren.
- [x] Zusätzlich eine kleine Token-Span-Funktion für Syntax-Highlighting einer einzelnen sichtbaren Zeile bereitstellen (gibt Byte-Spans für ID, Typ, Strings, Referenzen, Keywords in einen wiederverwendbaren Slice zurück – Zero-Alloc!).
- [x] Umfangreiche Table-Driven Unit-Tests (normale IFC-Entities, Strings mit `#` drin, fehlende GlobalId `$`, leere Zeilen, Windows `\r\n` vs Unix `\n`).
- [x] Go-Benchmark `BenchmarkParseLine` schreiben und verifizieren, dass `0 allocs/op` erreicht wird!
- [x] **Gate:** `go test -bench=. -benchmem ./internal/step` zeigt **0 allocs/op**, Native + WASM Build grün.

## Phase 4: Indizes & Background-Indexer (`internal/index`)
- [x] `LineIndex`: Scannt `doc.Bytes()` blitzschnell nach `\n` (via `bytes.IndexByte` in einer Schleife) und füllt `[]uint64` mit den Start-Offsets jeder Zeile (plus Zeilenlänge berechenbar aus nächstem Offset).
- [x] `PagedEntityIndex`: Implementiert die 2-stufige Page-Table (`[][]uint64`, Page-Size 4096) zur Abbildung `EntityID -> LineNumber` (speichere `lineIdx + 1`, damit `0` = "nicht vorhanden" bedeutet).
- [x] `TypeStats`: Zählt beim Hintergrund-Scan alle vorkommenden IFC-Entity-Typen und speichert optional pro Typ eine kompakte `[]uint64`-Liste der Zeilennummern.
- [x] Asynchroner `Builder`: Baut den `LineIndex` sofort auf, startet dann eine Goroutine für den `PagedEntityIndex` + `TypeStats`. Stellt atomare Progress-Getter bereit.
- [x] Thread-Sicherheit gewährleisten.
- [x] Unit-Tests mit 100.000 synthetischen IFC-Zeilen inkl. riesiger ID-Lücken.
- [x] **Gate:** Alle Tests mit `-race`, Native + WASM Build grün.

## Phase 5: Suche & Navigation (`internal/search`)
- [x] Direkt-Sprung zu Entity-ID (`JumpToEntity(id uint64) (line uint64, found bool)`).
- [x] Schnelle case-insensitive Streaming-Byte-Suche mit O(1)-ID-Erkennung implementieren.
- [x] Reverse-Referenz-Suche (`FindReferrers(targetID uint64, limit int)`) implementieren.
- [x] Unit-Tests für ID-Sprung, GlobalId-Suche und Reverse-Referenzen.
- [x] **Gate:** Tests mit `-race`, Native + WASM Build grün.

## Phase 6: Fyne Desktop- & WASM-UI (`internal/ui` & `cmd/stepview`)
- [x] Dunkles technisches IDE-Theme mit STEP-Syntaxfarben implementieren.
- [x] Top-Bar mit Dateimetadaten, Indexierungsfortschritt, Suche und Ladeaktionen implementieren.
- [x] Persistente virtuelle 3-Spalten-UI für Navigator, Viewer und Inspector implementieren.
- [x] CLI-Integration mit Desktop-Dateiargument und leerem Startdokument implementieren.
- [x] Gio-Implementierung unter `archive/gio-ui/` archivieren und Fyne mit nativen Menüs, Dialogen, Split-Panes und virtuellen Listen migrieren.
- [x] **Gate:** `go vet ./...`, `go test -race ./...`, Native und WASM Build fehlerfrei.

## Phase 7: End-to-End Belastungstest & Abschluss
- [x] Integrationstest/Benchmark mit 250.000 synthetischen IFC-Zeilen, Mmap, Indizierung und Suche schreiben.
- [x] Finales Verifikations-Gate (`go vet`, `go test -race -bench=. -benchmem ./...`, Native und WASM Build) ausführen.
- [x] `TODO.md` vollständig abhaken und Build-, Start- und Benchmarkwerte dokumentieren.

Aktuell verifiziert mit Fyne:

```text
go vet ./...
go test -race -bench=. -benchmem ./...
go build ./cmd/stepview
GOOS=js GOARCH=wasm go build ./cmd/stepview
```

Für einen reproduzierbaren Linux-Start:

```text
make verify-linux
make run-linux
```

End-to-End-Test: 250.000 IFC-Zeilen / 22.638.895 Byte in 113 ms, Heap-Delta
11.957.768 Byte. `BenchmarkParseLine`: 799,9 ns/op, 0 B/op, 0 allocs/op.

Windows 11 amd64 (portabel): `make build-windows` erzeugt
`bin/windows-amd64/IFC-Raw-Viewer_mk2.exe` sowie
`bin/IFC-Raw-Viewer_mk2-windows-amd64.zip`.

## Phase 8: MK2 – lange STEP-Records
- [x] Tabellen-Syntaxdarstellung auf die ersten 2 KiB jeder sichtbaren Zeile begrenzen.
- [x] Für längere Zeilen einen `More`-Button und einen seitenweisen Volltextdialog ergänzen.
- [x] `More` im wiederverwendeten Zellrenderer initial ausblenden, damit die Button-Mindesthöhe keine zusätzlichen Zeilenabstände erzeugt.
- [x] Darstellungsspaltengröße begrenzen, damit der horizontale Scrollbereich nicht für extrem lange Records aufgebaut wird.
- [x] Branding und portable Artefakte auf `IFC-Raw-Viewer_mk2` aktualisieren.
- [x] **Gate:** `go vet ./...`, `go test -race ./...`, Native- und WASM-Build sowie Windows-ZIP-Prüfung erfolgreich.

## Phase 9: MK2 – Dateiladung
- [x] Demo-Aktionen und eingebettete Demo-Datei entfernen.
- [x] Beim Neuladen die alte Mmap-Datei erst nach Abschluss ihres Hintergrundindexers schließen.
- [x] Regressionstest für das Nachladen während laufender Indizierung ergänzen.
- [x] **Gate:** `go vet ./...`, `go test -race ./...`, Native- und WASM-Build sowie Windows-ZIP-Prüfung erfolgreich.

## Phase 10: Dokumentation & GitHub Automation
- [x] README an Aufbau und Informationsdichte von Metric Neo angleichen.
- [x] GitHub-Action für Vet, Race-Tests sowie Linux- und WASM-Build ergänzen.
- [x] Tag-basierte Release-Pipeline für ein versioniertes Windows-Portable-Paket ergänzen.
- [x] Release-Tags auf die Produktmarkierung `mk*` statt semantischer `v*`-Tags ausrichten.
- [x] Workflow-YAML sowie Native- und WASM-Build erfolgreich validieren.

## Phase 11: MK2 – Raw-Editor
- [x] Versteckte `More`-Spalte aus der Tabellenansicht entfernen.
- [x] Im Inspector nach jeder Zeilenauswahl `Zeige alles` für den vollständigen Raw-Zugriff anbieten.
- [x] Read-only Raw-Editor mit 16-KiB-Seitennavigation implementieren.
- [x] **Gate:** `go vet ./...`, `go test -race ./...`, Native-/WASM-Build sowie Windows-ZIP-Prüfung erfolgreich.

## Phase 12: MK2 – virtuelle Auswahl
- [x] Vollständiges Tabellen-Refresh bei Auswahl durch gezielte Aktualisierung der alten und neuen Zeile ersetzen.
- [x] Auswahl beim Wechsel von Typfilter, Suche und Referenzliste sauber zurücksetzen.
- [x] **Gate:** `go vet ./...`, `go test -race ./...`, Native- und WASM-Build erfolgreich.

## Phase 13: MK2 – stabile Großdatei-Navigation
- [x] Fyne-Scrollpräzision als Ursache für fehlerhafte Auswahl nach sehr weiten Scrollwegen identifizieren.
- [x] Tabellenansicht auf 50.000-Zeilen-Fenster mit Vor-/Zurück-Navigation begrenzen.
- [x] Sprünge, Suche, Typfilter und Referenzlisten an die Fensterposition anpassen.
- [x] Seitengrenzen-Mapping testen und Windows-Portable-Paket neu bauen.
- [x] Horizontale Auswahl-Scrollposition nach Fyne-`Select` auf den linken Rand zurücksetzen.

## Phase 14: MK2 – Raw-Editor & Theme
- [x] Deaktiviertes Eingabefeld durch kontrastreiche scrollbarere Raw-Textansicht ersetzen.
- [x] Leere Seitennavigation für kurze Zeilen ausblenden und Abschnittskopie ergänzen.
- [x] Inspector-Aktion als persistenten einzelnen `Raw-Editor öffnen`-Button führen.
- [x] Wellermann Light Mode über *Einstellungen* ergänzen.
- [x] Theme-Test, Vet, Race-Tests, Native-/WASM-Build sowie Windows-ZIP-Prüfung erfolgreich ausführen.

## Phase 15: MK2 – kontinuierliche Standardnavigation
- [x] Seitennavigation erst oberhalb von 500.000 Zeilen aktivieren.
- [x] Für kleinere Quellen die Steuerung ausblenden und vollständig kontinuierliches Scrollen beibehalten.
- [x] Grenzwerte für kontinuierliche und seitenweise Quellen testen sowie Artefakte neu bauen.
