//go:build wasm

package ui

import "fmt"

func SelectIFCFile() (string, error) {
	return "", fmt.Errorf("Dateiauswahl ist im WASM-Build nicht verfügbar")
}
