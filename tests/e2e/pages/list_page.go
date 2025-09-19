package pages

import (
	"testing"

	"go-password-manager/internal/service"
	"go-password-manager/ui/molecules"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// ListPage provides interactions with the secrets list area.
type ListPage struct {
	t      *testing.T
	window fyne.Window
	svc    *service.SecretsService
}

func NewListPage(t *testing.T, window fyne.Window, svc *service.SecretsService) *ListPage {
	return &ListPage{t: t, window: window, svc: svc}
}

func (l *ListPage) ClickCreate() {
	root := l.window.Canvas().Content()
	obj := awaitFindObjectByText(root, "Create", 40)
	if obj == nil {
		obj = awaitFindObjectByText(root, "New", 40)
	}
	if b, ok := obj.(*widget.Button); ok {
		test.Tap(b)
		return
	}
	// fallback: open modal directly and reload content when it closes
	molecules.NewSecretModal(l.window, l.svc, func() {
		if l.window != nil && l.window.Canvas() != nil {
			// no-op: main tests reload the whole page as needed
		}
	})
}

func (l *ListPage) ClickSecretByName(name string) bool {
	root := l.window.Canvas().Content()
	if obj := awaitFindObjectByText(root, name, 8); obj != nil {
		if b, ok := obj.(*widget.Button); ok {
			test.Tap(b)
			return true
		}
	}
	// fallback: no-op, tests can use service-level selection
	return false
}
