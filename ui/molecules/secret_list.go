package molecules

import (
	"go-password-manager/internal/domain"
	"go-password-manager/ui/atoms"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func SecretList(secrets []domain.Secret, onSelect func(idx int), onDelete func(name string)) fyne.CanvasObject {
	list := container.NewVBox()
	for i, s := range secrets {
		list.Add(atoms.SecretName(s, func() { onSelect(i) }, func() { onDelete(s.SecretName) }))
	}
	return list
}
