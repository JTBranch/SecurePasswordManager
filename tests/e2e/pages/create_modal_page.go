package pages

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// CreateModalPage wraps interactions with the NewSecretModal
type CreateModalPage struct {
	t      *testing.T
	window fyne.Window
}

func NewCreateModalPage(t *testing.T, window fyne.Window) *CreateModalPage {
	return &CreateModalPage{t: t, window: window}
}

func (c *CreateModalPage) FillAndSubmit(name, value string) {
	root := c.window.Canvas().Content()
	nameEntry := awaitFindEntryByPlaceholder(root, "Secret name", 10)
	valueEntry := awaitFindEntryByPlaceholder(root, "Secret value", 10)
	if nameEntry != nil {
		nameEntry.SetText(name)
	}
	if valueEntry != nil {
		valueEntry.SetText(value)
	}
	// Try to find Create/Save button
	if b := awaitFindObjectByText(root, "Create", 6); b != nil {
		if btn, ok := b.(*widget.Button); ok {
			test.Tap(btn)
			return
		}
	}
	if b := awaitFindObjectByText(root, "Save", 6); b != nil {
		if btn, ok := b.(*widget.Button); ok {
			test.Tap(btn)
			return
		}
	}
}
