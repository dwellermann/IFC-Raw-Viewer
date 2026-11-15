# STEPView Architecture

STEPView is a read-only IFC-SPF viewer designed for very large files.

1. **Mapped bytes:** Desktop documents use read-only `mmap-go`; document and parser APIs operate on byte slices. Text is converted with `unsafe.String` only for immediate rendering.
2. **FSM parsing:** The project does not use `regexp`. STEP parsing, highlighting, and search use byte scanners.
3. **Compact indexing:** Line starts are `[]uint64`. Entity lookup uses a 4096-slot paged slice table rather than a map, and is built asynchronously.
4. **Platform separation:** Desktop-specific file mapping and WASM memory documents are isolated by build tags. A sample IFC is always available.
5. **Virtual UI:** The Fyne frontend uses pooled virtual lists, native menus, split panes, file dialogs, and scroll containers. The viewer renders only visible rows.

## Linux build

Use `make build-linux`, `make run-linux`, or `make verify-linux`. For a
portable Windows 11 amd64 executable and ZIP, use `make build-windows`; it
uses the Fyne-supported `fyne-cross` Podman build environment. The prior Gio
implementation is preserved under `archive/gio-ui/` and is not compiled.
