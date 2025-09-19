package pages

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// traverse walks the canvas tree and invokes fn for each CanvasObject.
func traverse(obj fyne.CanvasObject, fn func(fyne.CanvasObject)) {
	if obj == nil {
		return
	}
	fn(obj)
	type hasObjects interface {
		Objects() []fyne.CanvasObject
	}
	if c, ok := obj.(hasObjects); ok {
		for _, child := range c.Objects() {
			traverse(child, fn)
		}
	}
}

func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// findAllButtons returns all widget.Buttons found under the canvas content.
func findAllButtons(c fyne.Canvas) []*widget.Button {
	var out []*widget.Button
	root := c.Content()
	traverse(root, func(obj fyne.CanvasObject) {
		if b, ok := obj.(*widget.Button); ok {
			out = append(out, b)
		}
	})
	return out
}

// findButtonByText searches under the canvas content for a button matching text.
func findButtonByText(c fyne.Canvas, txt string) *widget.Button {
	root := c.Content()
	var found *widget.Button
	traverse(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		if b, ok := obj.(*widget.Button); ok {
			if b.Text == txt || containsIgnoreCase(b.Text, txt) {
				found = b
			}
		}
	})
	return found
}

// findAllLabels returns all widget.Labels found under the canvas content.
func findAllLabels(c fyne.Canvas) []*widget.Label {
	var out []*widget.Label
	root := c.Content()
	traverse(root, func(obj fyne.CanvasObject) {
		if l, ok := obj.(*widget.Label); ok {
			out = append(out, l)
		}
	})
	return out
}

// findAllEntries returns all widget.Entry objects under the given root object.
func findAllEntries(root fyne.CanvasObject) []*widget.Entry {
	var out []*widget.Entry
	traverse(root, func(obj fyne.CanvasObject) {
		if e, ok := obj.(*widget.Entry); ok {
			out = append(out, e)
		}
	})
	return out
}

// findButtonByTextFromRoot searches under a CanvasObject root for a button matching text.
func findButtonByTextFromRoot(root fyne.CanvasObject, txt string) *widget.Button {
	var found *widget.Button
	traverse(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		if b, ok := obj.(*widget.Button); ok {
			if b.Text == txt || containsIgnoreCase(b.Text, txt) {
				found = b
			}
		}
	})
	return found
}

// awaitFindButtonByText retries finding a button by text for up to `retries` iterations
func awaitFindButtonByText(root fyne.CanvasObject, txt string, retries int) *widget.Button {
	for i := 0; i < retries; i++ {
		if root == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		if b := findButtonByTextFromRoot(root, txt); b != nil {
			return b
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

// awaitFindEntryByPlaceholder retries finding an Entry whose placeholder matches the given text
func awaitFindEntryByPlaceholder(root fyne.CanvasObject, placeholder string, retries int) *widget.Entry {
	for i := 0; i < retries; i++ {
		if root == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		var found *widget.Entry
		match := func(obj fyne.CanvasObject) {
			if found != nil {
				return
			}
			if e, ok := obj.(*widget.Entry); ok {
				if e.PlaceHolder == placeholder || containsIgnoreCase(e.PlaceHolder, placeholder) {
					found = e
				}
			}
		}
		traverse(root, match)
		if found != nil {
			return found
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

// awaitFindObjectByText retries finding any CanvasObject (Button, Label) matching the text
func awaitFindObjectByText(root fyne.CanvasObject, txt string, retries int) fyne.CanvasObject {
	for i := 0; i < retries; i++ {
		if root == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		var found fyne.CanvasObject
		match := func(obj fyne.CanvasObject) {
			if found != nil {
				return
			}
			if matchesObjectText(obj, txt) {
				found = obj
			}
		}
		traverse(root, match)
		if found != nil {
			return found
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

// matchesObjectText checks if the CanvasObject has text matching txt.
func matchesObjectText(obj fyne.CanvasObject, txt string) bool {
	switch v := obj.(type) {
	case *widget.Button:
		return v.Text == txt || containsIgnoreCase(v.Text, txt)
	case *widget.Label:
		return v.Text == txt || containsIgnoreCase(v.Text, txt)
	default:
		return false
	}
}

// findLabelByText searches the canvas for a label with exact text match.
func findLabelByText(c fyne.Canvas, text string) *widget.Label {
	labels := findAllLabels(c)
	for _, l := range labels {
		if l.Text == text {
			return l
		}
	}
	return nil
}

// findContainerContainingChild traverses root and returns the first container that directly contains child.
func findContainerContainingChild(root fyne.CanvasObject, child fyne.CanvasObject) fyne.CanvasObject {
	var found fyne.CanvasObject
	traverse(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		type hasObjects interface{ Objects() []fyne.CanvasObject }
		if c, ok := obj.(hasObjects); ok {
			for _, ch := range c.Objects() {
				if ch == child {
					found = obj
					return
				}
			}
		}
	})
	return found
}

// objectText returns textual content for common widget types.
func objectText(o fyne.CanvasObject) string {
	if l, ok := o.(*widget.Label); ok {
		return l.Text
	}
	if b, ok := o.(*widget.Button); ok {
		return b.Text
	}
	return ""
}

// dumpCanvasDiagnostics walks the root and returns a slice of type/text diagnostics.
func dumpCanvasDiagnostics(root fyne.CanvasObject) []string {
	var out []string
	traverse(root, func(o fyne.CanvasObject) {
		out = append(out, fmt.Sprintf("%T:'%s'", o, objectText(o)))
	})
	return out
}

// findRevealButton searches for a reveal or hide button under root.
func findRevealButton(root fyne.CanvasObject) *widget.Button {
	if root == nil {
		return nil
	}
	// search for eye or hide emoji
	if obj := awaitFindObjectByText(root, "👁", 6); obj != nil {
		if b, ok := obj.(*widget.Button); ok {
			return b
		}
	}
	if obj := awaitFindObjectByText(root, "🙈", 6); obj != nil {
		if b, ok := obj.(*widget.Button); ok {
			return b
		}
	}
	return nil
}
