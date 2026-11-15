//go:build windows

package ui

import (
	"fmt"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	ofnExplorer       = 0x00080000
	ofnFileMustExist  = 0x00001000
	ofnPathMustExist  = 0x00000800
	ofnHideReadOnly   = 0x00000004
	windowsPathBuffer = 32768
)

type openFileName struct {
	structSize       uint32
	owner            uintptr
	instance         uintptr
	filter           *uint16
	customFilter     *uint16
	maxCustomFilter  uint32
	filterIndex      uint32
	file             *uint16
	maxFile          uint32
	fileTitle        *uint16
	maxFileTitle     uint32
	initialDirectory *uint16
	title            *uint16
	flags            uint32
	fileOffset       uint16
	fileExtension    uint16
	defaultExtension *uint16
	customData       uintptr
	hook             uintptr
	templateName     *uint16
	reserved         uintptr
	reservedFlags    uint32
	exFlags          uint32
}

var (
	comdlg32             = windows.NewLazySystemDLL("comdlg32.dll")
	getOpenFileName      = comdlg32.NewProc("GetOpenFileNameW")
	commDlgExtendedError = comdlg32.NewProc("CommDlgExtendedError")
)

// SelectIFCFile opens Windows Explorer's native common open-file dialog.
// An empty path means that the user cancelled the dialog.
func SelectIFCFile() (string, error) {
	filter := utf16.Encode([]rune("IFC-Dateien (*.ifc;*.stp;*.step)\x00*.ifc;*.stp;*.step\x00Alle Dateien (*.*)\x00*.*\x00"))
	title, err := windows.UTF16PtrFromString("IFC-Raw-Viewer – IFC öffnen")
	if err != nil {
		return "", err
	}
	path := make([]uint16, windowsPathBuffer)
	dialog := openFileName{
		structSize: uint32(unsafe.Sizeof(openFileName{})),
		filter:     &filter[0],
		file:       &path[0],
		maxFile:    uint32(len(path)),
		title:      title,
		flags:      ofnExplorer | ofnFileMustExist | ofnPathMustExist | ofnHideReadOnly,
	}
	result, _, callErr := getOpenFileName.Call(uintptr(unsafe.Pointer(&dialog)))
	if result != 0 {
		return windows.UTF16ToString(path), nil
	}
	code, _, _ := commDlgExtendedError.Call()
	if code == 0 {
		return "", nil
	}
	if callErr != nil && callErr != windows.ERROR_SUCCESS {
		return "", callErr
	}
	return "", fmt.Errorf("Windows-Dateidialog fehlgeschlagen: Code %d", code)
}
