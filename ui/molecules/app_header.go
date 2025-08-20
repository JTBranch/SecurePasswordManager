package molecules

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type AppHeaderProps struct {
	OnSearch       func(string)
	OnCreateSecret func()
	OnMenuAction   func()
	OnThemeChange  func(themeName string) // Add this for theme switching
}

// headerLayout lays out the search box at 50% width and the buttons at the far right, with padding.
type headerLayout struct{}

const padding = 16      // px between search box and buttons
const buttonSpacing = 8 // px between buttons

func (headerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 3 {
		return
	}
	menuBtnMin := objects[0].MinSize()
	createBtnMin := objects[2].MinSize()

	// Calculate available width for search box
	searchWidth := size.Width - menuBtnMin.Width - createBtnMin.Width - 2*buttonSpacing

	maxBtnHeight := fyne.Max(menuBtnMin.Height, createBtnMin.Height)
	btnY := (size.Height - maxBtnHeight) / 2

	// Position menu button (far left)
	objects[0].Resize(menuBtnMin)
	objects[0].Move(fyne.NewPos(0, float32(btnY)))

	// Position search box (fills space between buttons)
	objects[1].Resize(fyne.NewSize(searchWidth, size.Height))
	objects[1].Move(fyne.NewPos(menuBtnMin.Width+buttonSpacing, 0))

	// Position create button (far right)
	objects[2].Resize(createBtnMin)
	objects[2].Move(fyne.NewPos(size.Width-createBtnMin.Width, float32(btnY)))
}

func (headerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(0, 0)
	if len(objects) == 3 {
		searchMin := objects[0].MinSize()
		menuBtnMin := objects[1].MinSize()
		createBtnMin := objects[2].MinSize()

		totalBtnWidth := menuBtnMin.Width + buttonSpacing + createBtnMin.Width
		min.Width = searchMin.Width + padding + totalBtnWidth
		min.Height = fyne.Max(searchMin.Height, fyne.Max(menuBtnMin.Height, createBtnMin.Height))
	}
	return min
}

// AppHeader renders the search box, menu button, and create button in a responsive header
func AppHeader(props AppHeaderProps, win fyne.Window) fyne.CanvasObject {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search...")
	searchEntry.OnChanged = props.OnSearch

	menuBtn := widget.NewButton("☰", nil)
	menuBtn.Importance = widget.MediumImportance

	menuBtn.OnTapped = func() {
		themesSubMenu := fyne.NewMenu("Themes",
			fyne.NewMenuItem("Light Theme", func() {
				if props.OnThemeChange != nil {
					props.OnThemeChange("light")
				}
			}),
			fyne.NewMenuItem("Dark Theme", func() {
				if props.OnThemeChange != nil {
					props.OnThemeChange("dark")
				}
			}),
		)
		themesItem := fyne.NewMenuItem("Themes", nil)
		themesItem.ChildMenu = themesSubMenu

		mainMenu := fyne.NewMenu("Menu", themesItem /*, other items here */)
		pop := widget.NewPopUpMenu(mainMenu, win.Canvas())
		pop.ShowAtPosition(menuBtn.Position().AddXY(0, menuBtn.Size().Height))
	}

	createBtn := widget.NewButton("Create Secret", props.OnCreateSecret)

	return container.New(&headerLayout{}, menuBtn, searchEntry, createBtn)
}
