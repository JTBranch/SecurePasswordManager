package molecules

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AppHeaderProps struct {
	OnSearch       func(string)
	OnCreateSecret func()
	OnMenuAction   func()
	OnExport       func()
	OnThemeChange  func(themeName string)
	OnThemeApplied func()
	CurrentTheme   string
}

// AppHeader renders a compact, responsive header with a small app label, search,
// theme selector and create button. Uses simple layout primitives for predictable behavior.
func AppHeader(props AppHeaderProps, win fyne.Window) fyne.CanvasObject {
	// App label / pseudo-logo with optional image
	var logo fyne.CanvasObject
	// load dark variant if available; for light/system themes use the standard main icon
	var lightRes, darkRes fyne.Resource
	var img *canvas.Image
	// helper: try multiple locations so assets work from bundles or different CWDs
	findAsset := func(name string) ([]byte, error) {
		// try direct path relative to cwd
		p := filepath.Join("ui", "assets", name)
		if b, err := os.ReadFile(p); err == nil {
			return b, nil
		}
		// try next to the executable (app bundle case)
		if ex, err := os.Executable(); err == nil {
			dir := filepath.Dir(ex)
			// macOS app bundles put resources in Contents/Resources
			tryPaths := []string{
				filepath.Join(dir, name),
				filepath.Join(dir, "Contents", "Resources", name),
				filepath.Join(dir, "ui", "assets", name),
			}
			for _, tp := range tryPaths {
				if b, err := os.ReadFile(tp); err == nil {
					return b, nil
				}
			}
		}
		// finally, try name directly
		return os.ReadFile(name)
	}

	// main icon used for light/system themes
	if b, err := findAsset("main-icon.png"); err == nil {
		lightRes = fyne.NewStaticResource("main-icon.png", b)
	}
	// dark-specific icon (optional)
	if b, err := findAsset("main-icon-dark.png"); err == nil {
		darkRes = fyne.NewStaticResource("main-icon-dark.png", b)
	}
	// fallback to main-icon.png if specific variants not present
	if lightRes == nil {
		if b, err := os.ReadFile("ui/assets/main-icon.png"); err == nil {
			lightRes = fyne.NewStaticResource("main-icon.png", b)
		}
	}
	// choose starting resource based on CurrentTheme: dark -> darkRes, otherwise main icon
	var startRes fyne.Resource
	if props.CurrentTheme == "dark" && darkRes != nil {
		startRes = darkRes
	} else if lightRes != nil {
		startRes = lightRes
	} else if darkRes != nil {
		startRes = darkRes
	}
	if startRes != nil {
		img = canvas.NewImageFromResource(startRes)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(36, 36))
		// Put the image inside a fixed-size box so it won't collapse/stetch in different layouts
		// use a transparent rectangle to force a fixed box size, then overlay the image
		rect := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
		rect.SetMinSize(fyne.NewSize(48, 48))
		box := container.NewStack(rect, img)
		logo = container.NewCenter(box)
	} else {
		// fallback simple label centered
		logo = container.NewCenter(widget.NewLabelWithStyle("GoPass", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search secrets…")
	searchEntry.OnChanged = props.OnSearch

	// Settings menu: contains theme choices and export action
	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() { /* menu opens in OnTapped below */ })
	settingsBtn.OnTapped = func() {
		var pop *widget.PopUp
		lightBtn := widget.NewButton("Theme: Light", func() {
			if pop != nil {
				pop.Hide()
			}
			if props.OnThemeChange != nil {
				props.OnThemeChange("light")
			}
			// update header image to light variant if available
			if img != nil {
				if lightRes != nil {
					img.Resource = lightRes
				}
				img.Refresh()
			}
			if props.OnThemeApplied != nil {
				props.OnThemeApplied()
			}
		})
		darkBtn := widget.NewButton("Theme: Dark", func() {
			if pop != nil {
				pop.Hide()
			}
			if props.OnThemeChange != nil {
				props.OnThemeChange("dark")
			}
			// update header image to dark variant if available
			if img != nil {
				if darkRes != nil {
					img.Resource = darkRes
				}
				img.Refresh()
			}
			if props.OnThemeApplied != nil {
				props.OnThemeApplied()
			}
		})
		systemBtn := widget.NewButton("Theme: System", func() {
			if pop != nil {
				pop.Hide()
			}
			if props.OnThemeChange != nil {
				props.OnThemeChange("system")
			}
			// for system, prefer light variant as a reasonable default
			if img != nil {
				if lightRes != nil {
					img.Resource = lightRes
				}
				img.Refresh()
			}
			if props.OnThemeApplied != nil {
				props.OnThemeApplied()
			}
		})
		// menu layout: title + theme buttons in a row, then other settings
		title := widget.NewLabelWithStyle("Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		// theme options as a horizontal set for a compact look
		themeRow := container.NewHBox(lightBtn, darkBtn, systemBtn)
		close := widget.NewButton("Close", func() { /* attached below */ })
		menu := container.NewVBox(title, themeRow, widget.NewSeparator(), close)
		pop = widget.NewModalPopUp(menu, win.Canvas())
		close.OnTapped = func() { pop.Hide() }
		pop.Show()
	}

	exportBtn := widget.NewButtonWithIcon("Export", theme.DocumentIcon(), func() {
		if props.OnExport != nil {
			props.OnExport()
		}
	})
	createBtn := widget.NewButton("Create", props.OnCreateSecret)

	left := container.NewHBox(logo)
	// place searchEntry directly into the border layout center so it stretches
	middle := searchEntry
	right := container.NewHBox(settingsBtn, exportBtn, createBtn)

	// Use a responsive BorderLayout to keep search prominent.
	row := container.New(layout.NewBorderLayout(nil, nil, left, right), left, middle, right)
	return row
}
