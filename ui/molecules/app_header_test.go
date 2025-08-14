package molecules

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func TestAppHeaderRender(t *testing.T) {
	// Initialize test app
	testApp := app.New()
	defer testApp.Quit()

	// Test with all callbacks
	props := AppHeaderProps{
		OnSearch: func(query string) {
			// Test callback - intentionally empty for testing
		},
		OnCreateSecret: func() {
			// Test callback - intentionally empty for testing
		},
		OnMenuAction: func() {
			// Test callback - intentionally empty for testing
		},
	}

	header := AppHeader(props)
	if header == nil {
		t.Fatal("Expected AppHeader to return non-nil object")
	}
}

func TestAppHeaderLayout(t *testing.T) {
	// Initialize test app
	testApp := app.New()
	defer testApp.Quit()

	props := AppHeaderProps{
		OnSearch: func(string) {
			// No-op for test
		},
		OnCreateSecret: func() {
			// No-op for test
		},
		OnMenuAction: func() {
			// No-op for test
		},
	}

	header := AppHeader(props)

	// Test minimum size
	minSize := header.MinSize()
	if minSize.Width <= 0 || minSize.Height <= 0 {
		t.Error("Expected positive minimum size")
	}

	// Test that header can be resized
	header.Resize(fyne.NewSize(1000, 50))
	size := header.Size()
	if size.Width != 1000 || size.Height != 50 {
		t.Errorf("Expected size 1000x50, got %fx%f", size.Width, size.Height)
	}
}

func TestAppHeaderComponents(t *testing.T) {
	props := AppHeaderProps{
		OnSearch: func(string) {
			// No-op for test
		},
		OnCreateSecret: func() {
			// No-op for test
		},
		OnMenuAction: func() {
			// No-op for test
		},
	}

	header := AppHeader(props)

	// The header should be a container with our custom layout
	container, ok := header.(*fyne.Container)
	if !ok {
		t.Fatal("Expected AppHeader to return a Container")
	}

	// Should have 3 objects: search entry, create button, menu button
	if len(container.Objects) != 3 {
		t.Fatalf("Expected 3 objects in header container, got %d", len(container.Objects))
	}

	// First object should be search entry
	if _, ok := container.Objects[0].(*widget.Entry); !ok {
		t.Error("Expected first object to be Entry (search box)")
	}

	// Second object should be create button
	if btn, ok := container.Objects[1].(*widget.Button); !ok {
		t.Error("Expected second object to be Button (create)")
	} else if btn.Text != "Create Secret" {
		t.Errorf("Expected create button text 'Create Secret', got '%s'", btn.Text)
	}

	// Third object should be menu button
	if btn, ok := container.Objects[2].(*widget.Button); !ok {
		t.Error("Expected third object to be Button (menu)")
	} else if btn.Text != "☰" {
		t.Errorf("Expected menu button text '☰', got '%s'", btn.Text)
	}
}
