package molecules

import (
	"go-password-manager/tests/helpers"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

const testWindowTitle = "Test Window"

func TestAppHeaderRender(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestAppHeaderRender", func(tc *helpers.UnitTestCase) {
		testApp := app.New()
		fyne.Do(func() {
			defer testApp.Quit()
			mockWin := testApp.NewWindow(testWindowTitle)

			props := AppHeaderProps{
				OnSearch:       func(string) {},
				OnCreateSecret: func() {},
				OnMenuAction:   func() {},
				OnThemeChange:  func(string) {},
			}
			header := AppHeader(props, mockWin)
			if header == nil {
				t.Fatal("Expected AppHeader to return non-nil object")
			}
		})
	})
}

func TestAppHeaderLayout(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestAppHeaderLayout", func(tc *helpers.UnitTestCase) {
		testApp := app.New()
		fyne.Do(func() {
			defer testApp.Quit()
			mockWin := testApp.NewWindow(testWindowTitle)

			props := AppHeaderProps{
				OnSearch:       func(string) {},
				OnCreateSecret: func() {},
				OnMenuAction:   func() {},
			}
			header := AppHeader(props, mockWin)

			minSize := header.MinSize()
			if minSize.Width <= 0 || minSize.Height <= 0 {
				t.Error("Expected positive minimum size")
			}

			header.Resize(fyne.NewSize(1000, 50))
			size := header.Size()
			if size.Width != 1000 || size.Height != 50 {
				t.Errorf("Expected size 1000x50, got %fx%f", size.Width, size.Height)
			}
		})
	})
}

func TestAppHeaderComponents(t *testing.T) {
	helpers.WithUnitTestCase(t, "Components", func(tc *helpers.UnitTestCase) {
		testApp := app.New()
		fyne.Do(func() {
			defer testApp.Quit()
			mockWin := testApp.NewWindow(testWindowTitle)

			props := AppHeaderProps{
				OnSearch:       func(string) {},
				OnCreateSecret: func() {},
				OnMenuAction:   func() {},
			}
			header := AppHeader(props, mockWin)

			container, ok := header.(*fyne.Container)
			tc.Assert.True(ok, "Expected AppHeader to return a Container")

			tc.Assert.Len(container.Objects, 3, "Expected 3 objects in header container")

			_, ok = container.Objects[0].(*widget.Entry)
			tc.Assert.True(ok, "Expected first object to be Entry (search box)")

			btnCreate, ok := container.Objects[1].(*widget.Button)
			tc.Assert.True(ok, "Expected second object to be Button (create)")
			tc.Assert.Equal("Create Secret", btnCreate.Text, "Create button text mismatch")

			btnMenu, ok := container.Objects[2].(*widget.Button)
			tc.Assert.True(ok, "Expected third object to be Button (menu)")
			tc.Assert.Equal("☰", btnMenu.Text, "Menu button text mismatch")
		})
	})
}

func TestAppHeaderMenuButton(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestAppHeaderMenuButton", func(tc *helpers.UnitTestCase) {
		testApp := app.New()
		fyne.Do(func() {
			defer testApp.Quit()
			mockWin := testApp.NewWindow(testWindowTitle)

			menuClicked := false
			props := AppHeaderProps{
				OnSearch:       func(string) {},
				OnCreateSecret: func() {},
				OnMenuAction:   func() { menuClicked = true },
				OnThemeChange:  func(string) {},
			}
			header := AppHeader(props, mockWin)
			container, ok := header.(*fyne.Container)
			if !ok {
				t.Fatal("AppHeader did not return a container")
			}

			var found bool
			for _, obj := range container.Objects {
				if btn, ok := obj.(*widget.Button); ok && btn.Text == "☰" {
					found = true
					btn.OnTapped()
				}
			}
			if !found {
				t.Error("Menu button not found in AppHeader")
			}
			if !menuClicked {
				t.Error("Menu button OnTapped did not trigger OnMenuAction")
			}
		})
	})
}
