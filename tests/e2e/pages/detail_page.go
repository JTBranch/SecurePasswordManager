package pages

import (
	"testing"
	"time"

	"go-password-manager/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// DetailPage provides interactions with the secret detail panel.
type DetailPage struct {
	t      *testing.T
	window fyne.Window
	svc    *service.SecretsService
}

func NewDetailPage(t *testing.T, window fyne.Window, svc *service.SecretsService) *DetailPage {
	return &DetailPage{t: t, window: window, svc: svc}
}

func (d *DetailPage) ToggleReveal() {
	root := d.window.Canvas().Content()
	if b := findRevealButton(root); b != nil {
		test.Tap(b)
		time.Sleep(60 * time.Millisecond)
	}
}

func (d *DetailPage) ClickEdit() {
	root := d.window.Canvas().Content()
	if obj := awaitFindObjectByText(root, "✏️", 8); obj != nil {
		if b, ok := obj.(*widget.Button); ok {
			test.Tap(b)
			return
		}
	}
	if obj := awaitFindObjectByText(root, "💾", 4); obj != nil {
		if b, ok := obj.(*widget.Button); ok {
			test.Tap(b)
			return
		}
	}
	// fallback: no-op
}

func (d *DetailPage) UpdateValue(newValue string) {
	root := d.window.Canvas().Content()
	entries := findAllEntries(root)
	for _, e := range entries {
		if e.Visible() {
			e.SetText(newValue)
			return
		}
	}
}
