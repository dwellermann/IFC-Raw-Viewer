package ui

import (
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	colorLineNumber fyne.ThemeColorName = "stepview.lineNumber"
	colorEntityID   fyne.ThemeColorName = "stepview.entityID"
	colorEntityType fyne.ThemeColorName = "stepview.entityType"
	colorString     fyne.ThemeColorName = "stepview.string"
	colorReference  fyne.ThemeColorName = "stepview.reference"
	colorKeyword    fyne.ThemeColorName = "stepview.keyword"
)

type wellermannTheme struct {
	base    fyne.Theme
	font    fyne.Resource
	density float32
	light   bool
}

func newWellermannTheme(density float32, light bool) fyne.Theme {
	base := theme.DarkTheme()
	if light {
		base = theme.LightTheme()
	}
	custom := &wellermannTheme{base: base, density: density, light: light}
	for _, path := range []string{
		"/usr/share/fonts/truetype/jetbrains-mono/JetBrainsMono-Regular.ttf",
		"/usr/share/fonts/TTF/JetBrainsMono-Regular.ttf",
		"/usr/share/fonts/noto/NotoSansMono-Regular.ttf",
	} {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if resource, err := fyne.LoadResourceFromPath(path); err == nil {
			custom.font = resource
			break
		}
	}
	return custom
}

func (t *wellermannTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.light {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 246, G: 245, B: 239, A: 255}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 37, G: 36, B: 32, A: 255}
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 173, G: 96, B: 0, A: 255}
		case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameMenuBackground:
			return color.NRGBA{R: 232, G: 230, B: 221, A: 255}
		case theme.ColorNameScrollBar:
			return color.NRGBA{R: 173, G: 96, B: 0, A: 255}
		case theme.ColorNameScrollBarBackground:
			return color.NRGBA{R: 206, G: 204, B: 195, A: 255}
		case colorLineNumber:
			return color.NRGBA{R: 104, G: 101, B: 91, A: 255}
		case colorEntityID, colorReference:
			return color.NRGBA{R: 0, G: 95, B: 145, A: 255}
		case colorEntityType, colorKeyword:
			return color.NRGBA{R: 150, G: 77, B: 0, A: 255}
		case colorString:
			return color.NRGBA{R: 27, G: 112, B: 64, A: 255}
		}
	}
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 18, G: 18, B: 17, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 245, G: 242, B: 234, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 255, G: 177, B: 55, A: 255}
	case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameMenuBackground:
		return color.NRGBA{R: 34, G: 34, B: 31, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 255, G: 177, B: 55, A: 255}
	case theme.ColorNameScrollBarBackground:
		return color.NRGBA{R: 52, G: 51, B: 47, A: 255}
	case colorLineNumber:
		return color.NRGBA{R: 161, G: 160, B: 151, A: 255}
	case colorEntityID, colorReference:
		return color.NRGBA{R: 83, G: 182, B: 232, A: 255}
	case colorEntityType, colorKeyword:
		return color.NRGBA{R: 255, G: 177, B: 55, A: 255}
	case colorString:
		return color.NRGBA{R: 122, G: 218, B: 157, A: 255}
	default:
		return t.base.Color(name, variant)
	}
}

func (t *wellermannTheme) Font(style fyne.TextStyle) fyne.Resource {
	if t.font != nil {
		return t.font
	}
	return t.base.Font(style)
}

func (t *wellermannTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *wellermannTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText, theme.SizeNameCaptionText, theme.SizeNameHeadingText, theme.SizeNameSubHeadingText:
		return t.base.Size(name) * t.density
	case theme.SizeNamePadding, theme.SizeNameInnerPadding:
		return t.base.Size(name) * t.density
	case theme.SizeNameLineSpacing:
		return 0
	}
	return t.base.Size(name)
}
