package helpers

import (
	"fyne.io/fyne/v2"
)

// CopyToClipboard copies text to clipboard and shows a temporary tooltip
func CopyToClipboard(text string, window fyne.Window) {
	// Copy to clipboard
	fyne.CurrentApp().Clipboard().SetContent(text)

	// Create a temporary notification
	notification := fyne.NewNotification("Copied!", "Secret copied to clipboard")
	fyne.CurrentApp().SendNotification(notification)
}
