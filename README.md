<div align="center">
  <h1>IFC-Raw-Viewer</h1>

  <p><strong>High-performance, read-only IFC-SPF data explorer and STEP debugger.</strong></p>

  <p>
    <img src="https://img.shields.io/badge/Status-MK2-2f6f89?style=flat&logo=gitbook&logoColor=white" alt="Status">
    <img src="https://img.shields.io/badge/Backend-Go-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/UI-Fyne-4f46e5?style=flat" alt="Fyne">
    <img src="https://img.shields.io/badge/Platform-Windows%2011%20%7C%20Linux-lightblue?style=flat&logo=linux" alt="Platform">
    <img src="https://github.com/dwellermann/IFC-Editor/actions/workflows/ci.yml/badge.svg" alt="CI">
  </p>
</div>

---

> 🇩🇪 **Hinweis:** Oberfläche und Quellcode sind englisch. Diese README sowie
> Architektur- und Projektdokumentation sind auf Deutsch gepflegt.

## 💡 Warum IFC-Raw-Viewer?

IFC-Raw-Viewer ist für Situationen gedacht, in denen ein Modell nicht
interpretiert, sondern als STEP-Text nachvollzogen werden muss: beim Prüfen
eines Exports, beim Verfolgen von Entity-Referenzen oder beim Debuggen
auffälliger IFC-Daten. Er bleibt bewusst read-only und behandelt auch große
Dateien ohne vollständige Kopie in den Go-Heap.

## ⚡ Kernfunktionen

- **Mmap-first** — IFC-Dateien werden auf Desktop-Systemen read-only
  speicherabgebildet.
- **Großdatei-Indizes** — Byte-Offsets je Zeile und eine paged `#ID → Zeile`-
  Tabelle ohne Hashmap-Overhead.
- **Streaming-Suche** — Volltext, GlobalId, Entity-Sprung und
  Reverse-Referenzen.
- **Virtuelle Ansicht** — Nur sichtbare Zeilen werden gezeichnet; lange
  STEP-Records zeigen eine begrenzte Vorschau und lassen sich über
  **„Raw-Editor öffnen“** im Inspector vollständig seitenweise lesen und
  abschnittsweise kopieren.
- **Stabile Großdatei-Navigation** — Die Tabellenansicht nutzt
  bei Dateien über 500.000 Zeilen 50.000-Zeilen-Fenster mit
  Vor-/Zurück-Navigation, damit die floatbasierte UI-Scrollposition auch bei
  Millionen Zeilen präzise bleibt.
- **Native Dateiauswahl** — Windows Explorer sowie KDE `kdialog` mit
  `zenity`-Fallback unter Linux.
- **Keine Telemetrie** — Keine Cloud-Dienste, keine Netzwerkanfragen und keine
  Veränderung der geöffneten IFC-Datei.
- **Theme** — Kompakter Wellermann-Dark-Mode sowie ein kontrastreicher Light
  Mode über *Einstellungen*.

## 📦 Download

Die aktuelle Portable-Version steht auf der
[GitHub-Releases-Seite](../../releases/latest) bereit.

| Plattform | Datei |
|---|---|
| Windows 11 amd64 | `IFC-Raw-Viewer_mk2_VERSION_windows_amd64.zip` |
| Linux amd64 | Lokaler Build via `make build-linux` |
| Browser / WASM | Lokaler Build via `make build-wasm` |

## 🚀 Schnellstart

**Windows 11** — ZIP entpacken und `IFC-Raw-Viewer_mk2_VERSION_windows_amd64.exe`
starten. Kein Installer erforderlich.

**Linux**
```bash
make verify-linux
make run-linux
```

Eine IFC direkt öffnen:
```bash
./bin/IFC-Raw-Viewer_mk2 modell.ifc
```

**Browser / WASM**
```bash
make build-wasm
```

## 🛡️ Datenschutz & Datenhoheit

IFC-Raw-Viewer arbeitet lokal. Geöffnete IFC-Dateien werden ausschließlich
read-only verarbeitet; es gibt weder Telemetrie noch Cloud-Synchronisierung.

## 📚 Dokumentation

- [Architektur](ARCHITECTURE.md)
- [Projekt- und Verifikationshistorie](TODO.md)
- [Release-Workflow](.github/workflows/release.yml)

## 🌐 Projekt

Digitales Handwerk für IFC-Rohdaten von
[wellermann.de](https://wellermann.de).
