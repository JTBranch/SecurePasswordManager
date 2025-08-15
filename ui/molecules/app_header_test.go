package molecules

import (
	"go-password-manager/tests/helpers"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func TestAppHeaderRender(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestAppHeaderRender", func(tc *helpers.UnitTestCase) {
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
	})
}

func TestAppHeaderLayout(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestAppHeaderLayout", func(tc *helpers.UnitTestCase) {
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
	})
}

func TestAppHeaderComponents(t *testing.T) {
	helpers.WithUnitTestCase(t, "Components", func(tc *helpers.UnitTestCase) {
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

		// The header should be a container with our custom layout
		container, ok := header.(*fyne.Container)
		tc.Assert.True(ok, "Expected AppHeader to return a Container")

		// Should have 3 objects: search entry, create button, menu button
		tc.Assert.Len(container.Objects, 3, "Expected 3 objects in header container")

		// First object should be search entry
		_, ok = container.Objects[0].(*widget.Entry)
		tc.Assert.True(ok, "Expected first object to be Entry (search box)")

		// Second object should be create button
		btnCreate, ok := container.Objects[1].(*widget.Button)
		tc.Assert.True(ok, "Expected second object to be Button (create)")
		tc.Assert.Equal("Create Secret", btnCreate.Text, "Create button text mismatch")

		// Third object should be menu button
		btnMenu, ok := container.Objects[2].(*widget.Button)
		tc.Assert.True(ok, "Expected third object to be Button (menu)")
		tc.Assert.Equal("☰", btnMenu.Text, "Menu button text mismatch")
	})
}
