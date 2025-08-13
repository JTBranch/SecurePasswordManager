package molecules

import (
    "fmt"
    "go-password-manager/internal/domain"
    "go-password-manager/internal/service"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func SecretDetail(secret domain.Secret, secretsService *service.SecretsService) fyne.CanvasObject {
    revealed := false
    label := widget.NewLabel(fmt.Sprintf("%s: ******* [%s]", secret.SecretName, secret.Type))
    revealBtn := widget.NewButton("👁", func() {
        revealed = !revealed
        if revealed {
            plain, err := secretsService.DisplaySecret(secret)
            if err == nil {
                label.SetText(fmt.Sprintf("%s: %s [%s]", secret.SecretName, plain, secret.Type))
            }
        } else {
            label.SetText(fmt.Sprintf("%s: ******* [%s]", secret.SecretName, secret.Type))
        }
    })
    return container.NewVBox(label, revealBtn)
}
