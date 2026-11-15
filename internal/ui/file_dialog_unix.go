//go:build !wasm && !windows

package ui

import (
	"fmt"
	"os/exec"
	"strings"
)

// SelectIFCFile uses KDE's native chooser and falls back to Zenity elsewhere.
func SelectIFCFile() (string, error) {
	var command *exec.Cmd
	if path, err := exec.LookPath("kdialog"); err == nil {
		command = exec.Command(path, "--getopenfilename", "", "*.ifc|IFC-Dateien", "--title", "IFC-Raw-Viewer – IFC öffnen")
	} else if path, err := exec.LookPath("zenity"); err == nil {
		command = exec.Command(path, "--file-selection", "--title=IFC-Raw-Viewer – IFC öffnen", "--file-filter=IFC-Dateien | *.ifc")
	} else {
		return "", fmt.Errorf("kein nativer Dateidialog verfügbar: bitte kdialog oder zenity installieren")
	}
	output, err := command.Output()
	if err != nil {
		if _, cancelled := err.(*exec.ExitError); cancelled {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
