package molecules

import (
	"fmt"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	"go-password-manager/ui/atoms"
	"go-password-manager/ui/helpers"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func SecretHistory(secretName string, secretsService *service.SecretsService, window fyne.Window) fyne.CanvasObject {
	// Load all versions of this secret
	versions, err := secretsService.GetSecretVersions(secretName)
	if err != nil || len(versions) <= 1 {
		// If there's an error or only one version, show a simple message
		return widget.NewLabel("No version history available")
	}

	// Create header for history section
	historyLabel := widget.NewLabel("Version History")
	historyLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create container for version items
	historyBox := container.NewVBox()

	// Add each version (skip the latest one since it's shown above)
	// Versions are sorted descending, so skip the first one (current version)
	for _, version := range versions[1:] {
		historyBox.Add(createVersionItem(version, secretsService, window))
	}

	return container.NewVBox(
		widget.NewSeparator(),
		historyLabel,
		historyBox,
	)
}

func createVersionItem(version domain.SecretVersion, secretsService *service.SecretsService, window fyne.Window) fyne.CanvasObject {
	revealed := false
	var currentPlainValue string

	// Parse the date
	parsedTime, err := time.Parse(time.RFC3339, version.UpdatedAt)
	var dateStr string
	if err != nil {
		dateStr = version.UpdatedAt
	} else {
		dateStr = parsedTime.Format("Jan 2, 2006 15:04")
	}

	// Version label with date
	versionLabel := widget.NewLabel(fmt.Sprintf("Version %d - %s", version.Version, dateStr))
	versionLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Create a container that will hold the secret value atom
	valueContainer := container.NewVBox()

	// Function to update the value display
	var updateValueDisplay func()
	updateValueDisplay = func() {
		valueContainer.Objects = nil
		secretValueAtom := atoms.SecretValue(atoms.SecretValueProps{
			Value:      currentPlainValue,
			IsRevealed: revealed,
			OnRevealClick: func() {
				revealed = !revealed
				if revealed {
					// Decrypt the version directly
					plain, err := secretsService.DecryptSecretVersion(version)
					if err == nil {
						currentPlainValue = plain
					}
				}
				updateValueDisplay()
			},
			OnValueClick: func() {
				if revealed && currentPlainValue != "" {
					helpers.CopyToClipboard(currentPlainValue, window)
				}
			},
		})
		valueContainer.Objects = append(valueContainer.Objects, secretValueAtom)
		valueContainer.Refresh()
	}

	// Initial display
	updateValueDisplay()

	return container.NewVBox(
		versionLabel,
		valueContainer,
	)
}
