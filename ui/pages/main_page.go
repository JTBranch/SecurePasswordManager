package pages

import (
    "go-password-manager/internal/service"
    "go-password-manager/ui/molecules"
    "go-password-manager/ui/atoms"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "strings"
)

var secretsService = service.NewSecretsService("1.0.0", "jack.branch")

func MainPage(win fyne.Window) fyne.CanvasObject {
    fileData, _ := secretsService.LoadAllSecrets()
    var selectedIdx int = -1
    listBox := container.NewVBox()
    detailBox := container.NewVBox(widget.NewLabel("Select a secret"))

    var updateList func()
    updateDetail := func() {
        detailBox.Objects = nil
        if selectedIdx >= 0 && selectedIdx < len(fileData.Secrets) {
            detailBox.Add(molecules.SecretDetail(fileData.Secrets[selectedIdx], secretsService))
        } else {
            detailBox.Add(widget.NewLabel("Select a secret"))
        }
        detailBox.Refresh()
    }

    updateList = func() {
        fileData, _ = secretsService.LoadAllSecrets()
        listBox.Objects = nil
        for i, s := range fileData.Secrets {
            listBox.Add(atoms.SecretName(s, func(idx int) func() {
                return func() {
                    selectedIdx = idx
                    updateDetail()
                }
            }(i), func() {
                _ = secretsService.DeleteSecret(s.SecretName)
                selectedIdx = -1
                updateList()
                updateDetail()
            }))
        }
        listBox.Refresh()
    }

    // --- AppHeader logic moved to component ---
    header := molecules.AppHeader(molecules.AppHeaderProps{
        OnSearch: func(query string) {
            fileData, _ = secretsService.LoadAllSecrets()
            listBox.Objects = nil
            for i, s := range fileData.Secrets {
                if query == "" || containsIgnoreCase(s.SecretName, query) {
                    listBox.Add(atoms.SecretName(s, func(idx int) func() {
                        return func() {
                            selectedIdx = idx
                            updateDetail()
                        }
                    }(i), func() {
                        _ = secretsService.DeleteSecret(s.SecretName)
                        selectedIdx = -1
                        updateList()
                        updateDetail()
                    }))
                }
            }
            listBox.Refresh()
        },
        OnCreateSecret: func() {
            molecules.NewSecretModal(win, secretsService, func() {
                updateList()
            })
        },
    })

    updateList()

    split := container.NewHSplit(listBox, detailBox)
    split.SetOffset(0.3) // This sets the split ratio, not a fixed size

    content := container.NewBorder(
        header, // top
        nil,    // bottom
        nil,    // left
        nil,    // right
        container.NewHSplit(listBox, detailBox),
    )
    return content
}

// Helper for case-insensitive substring search
func containsIgnoreCase(s, substr string) bool {
    s = strings.ToLower(s)
    substr = strings.ToLower(substr)
    return strings.Contains(s, substr)
}
