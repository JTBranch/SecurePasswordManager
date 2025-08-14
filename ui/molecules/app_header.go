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
}

// headerLayout lays out the search box at 50% width and the buttons at the far right, with padding.
type headerLayout struct{}

const searchBoxPercent = 0.5
const padding = 16      // px between search box and buttons
const buttonSpacing = 8 // px between buttons

func (headerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 3 {
		return
	}
	searchWidth := int(float32(size.Width) * searchBoxPercent)

	// Get button sizes
	menuBtnMin := objects[1].MinSize()
	createBtnMin := objects[2].MinSize()

	// Calculate total button width (both buttons + spacing between them)
	totalBtnWidth := menuBtnMin.Width + buttonSpacing + createBtnMin.Width
	maxBtnHeight := fyne.Max(menuBtnMin.Height, createBtnMin.Height)

	// Vertically center the buttons
	btnY := (size.Height - maxBtnHeight) / 2

	// Position search box
	objects[0].Resize(fyne.NewSize(float32(searchWidth-padding), size.Height))
	objects[0].Move(fyne.NewPos(0, 0))

	// Position menu button (first button)
	objects[1].Resize(fyne.NewSize(menuBtnMin.Width, menuBtnMin.Height))
	objects[1].Move(fyne.NewPos(size.Width-totalBtnWidth, float32(btnY)))

	// Position create button (second button, to the right of menu button)
	objects[2].Resize(fyne.NewSize(createBtnMin.Width, createBtnMin.Height))
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
func AppHeader(props AppHeaderProps) fyne.CanvasObject {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search secrets...")
	searchEntry.OnChanged = props.OnSearch

	menuBtn := widget.NewButton("☰", props.OnMenuAction)
	menuBtn.Importance = widget.MediumImportance

	createBtn := widget.NewButton("Create Secret", props.OnCreateSecret)

	return container.New(&headerLayout{}, searchEntry, createBtn, menuBtn)
}
