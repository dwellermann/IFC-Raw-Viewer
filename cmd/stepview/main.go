package main

import (
	"os"

	"stepview/internal/document"
	"stepview/internal/ui"
)

func main() {
	var doc document.Document = document.NewMemoryDocument(nil)
	name := "Keine Datei geöffnet"
	if len(os.Args) > 1 {
		opened, err := document.Open(os.Args[1])
		if err == nil {
			doc = opened
			name = os.Args[1]
		}
	}
	ui.Run(doc, name)
}
