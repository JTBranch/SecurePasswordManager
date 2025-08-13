package molecules

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

type AppHeaderProps struct {
    OnSearch        func(string)
    OnCreateSecret  func()
}

// headerLayout lays out the search box at 50% width and the button at the far right, with padding.
type headerLayout struct{}

const searchBoxPercent = 0.5
const padding = 16 // px between search box and button

func (headerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
    if len(objects) != 2 {
        return
    }
    searchWidth := int(float32(size.Width) * searchBoxPercent)
    btnMin := objects[1].MinSize()
    btnWidth := btnMin.Width
    btnHeight := btnMin.Height

    // Vertically center the button
    btnY := (size.Height - btnHeight) / 2

    // Position search box
    objects[0].Resize(fyne.NewSize(float32(searchWidth-padding), size.Height))
    objects[0].Move(fyne.NewPos(0, 0))

    // Position button at far right, with padding
    objects[1].Resize(fyne.NewSize(btnWidth, btnHeight))
    objects[1].Move(fyne.NewPos(size.Width-btnWidth, float32(btnY)))
}

func (headerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
    min := fyne.NewSize(0, 0)
    if len(objects) == 2 {
        searchMin := objects[0].MinSize()
        btnMin := objects[1].MinSize()
        min.Width = searchMin.Width + padding + btnMin.Width
        min.Height = fyne.Max(searchMin.Height, btnMin.Height)
    }
    return min
}

// AppHeader renders the search box and create button in a responsive header.
func AppHeader(props AppHeaderProps) fyne.CanvasObject {
    searchEntry := widget.NewEntry()
    searchEntry.SetPlaceHolder("Search secrets...")
    searchEntry.OnChanged = props.OnSearch

    createBtn := widget.NewButton("Create Secret", props.OnCreateSecret)

    return container.New(&headerLayout{}, searchEntry, createBtn)
}
