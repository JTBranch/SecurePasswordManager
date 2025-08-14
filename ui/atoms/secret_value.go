package atoms

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SecretValueProps holds the properties for the secret value atom
type SecretValueProps struct {
	Value         string
	IsRevealed    bool
	OnRevealClick func()
	OnValueClick  func() // Called when the value text is clicked to copy
}

// SecretValue creates a reusable secret value display with reveal/hide and copy functionality
func SecretValue(props SecretValueProps) fyne.CanvasObject {
	hiddenText := "*******"

	// Create the value display
	var valueText string
	if props.IsRevealed {
		valueText = props.Value
	} else {
		valueText = hiddenText
	}

	var valueDisplay fyne.CanvasObject

	// Make the value clickable if revealed and onValueClick is provided
	if props.IsRevealed && props.OnValueClick != nil {
		// Create a button that looks like a label for the value
		valueButton := widget.NewButton(valueText, props.OnValueClick)
		valueButton.Importance = widget.LowImportance
		valueDisplay = valueButton
	} else {
		// Create a regular label
		valueDisplay = widget.NewLabel(valueText)
	}

	// Create reveal/hide button
	var revealBtn *widget.Button
	if props.IsRevealed {
		revealBtn = widget.NewButton("🙈", props.OnRevealClick)
	} else {
		revealBtn = widget.NewButton("👁", props.OnRevealClick)
	}

	// Layout: value on left, reveal button on right
	return container.NewBorder(nil, nil, nil, revealBtn, valueDisplay)
}
