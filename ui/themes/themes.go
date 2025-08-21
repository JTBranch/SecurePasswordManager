package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type LightTheme struct{}

func (t *LightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameButton:
		return color.RGBA{R: 30, G: 144, B: 255, A: 255} // DodgerBlue
	case theme.ColorNameBackground:
		return color.RGBA{R: 245, G: 245, B: 245, A: 255} // LightGray
	case theme.ColorNameError:
		return color.RGBA{R: 220, G: 53, B: 69, A: 255} // Bootstrap Danger Red
	case theme.ColorNamePrimary:
		return color.RGBA{R: 50, G: 205, B: 50, A: 255} // LimeGreen
	case theme.ColorNameDisabled:
		return color.RGBA{R: 200, G: 200, B: 200, A: 255} // Gray
	case theme.ColorNameForeground:
		return color.Black
	case theme.ColorNameHover:
		return color.RGBA{R: 210, G: 220, B: 235, A: 255} // Soft blue-gray for hover
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 200, G: 200, B: 200, A: 255} // Gray
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 230, G: 232, B: 236, A: 255} // Neutral light gray
	default:
		// fallback to a reasonable default
		return color.White
	}
}

func (t *LightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}
func (t *LightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
func (t *LightTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

type DarkTheme struct{}

func (t *DarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameButton:
		return color.RGBA{R: 20, G: 80, B: 160, A: 255} // Deep blue for dark mode
	case theme.ColorNameBackground:
		return color.RGBA{R: 24, G: 24, B: 28, A: 255} // Very dark gray
	case theme.ColorNameError:
		return color.RGBA{R: 180, G: 40, B: 60, A: 255} // Darker danger red
	case theme.ColorNamePrimary:
		return color.RGBA{R: 40, G: 160, B: 40, A: 255} // Dark lime green
	case theme.ColorNameDisabled:
		return color.RGBA{R: 80, G: 80, B: 80, A: 255} // Dim gray
	case theme.ColorNameForeground:
		return color.White
	case theme.ColorNameHover:
		return color.RGBA{R: 40, G: 100, B: 180, A: 255} // Soft blue for hover
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 60, G: 60, B: 60, A: 255} // Dark disabled
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 44, G: 48, B: 58, A: 255} // Neutral dark gray
	default:
		return color.Black
	}
}

func (t *DarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}
func (t *DarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
func (t *DarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

func ThemeFromName(name string) fyne.Theme {
	switch name {
	case "dark":
		return &DarkTheme{}
	case "light":
		return &LightTheme{}
	default:
		return &LightTheme{} // fallback
	}
}
