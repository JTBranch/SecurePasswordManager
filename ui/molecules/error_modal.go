package molecules

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ErrorModalProps struct {
	Title   string
	Message string
}

func ErrorModal(win fyne.Window, props ErrorModalProps) {
	if win == nil {
		return
	}
	lbl := widget.NewLabel(props.Message)
	var pop *widget.PopUp
	ok := widget.NewButton("OK", func() {
		if pop != nil {
			pop.Hide()
		}
	})
	content := container.NewVBox(widget.NewLabelWithStyle(props.Title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), lbl, ok)
	pop = widget.NewModalPopUp(content, win.Canvas())
	pop.Show()
}
