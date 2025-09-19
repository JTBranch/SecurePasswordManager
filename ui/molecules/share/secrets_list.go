package share

import (
	"go-password-manager/internal/sharing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// SecretsList builds multi-select list and binds to view model selections.
func SecretsList(vm *ViewModel, secrets []sharing.ExportSecret) *widget.List {
	if len(vm.SecretNames) == 0 {
		for _, s := range secrets {
			vm.SecretNames = append(vm.SecretNames, s.Name)
		}
	}
	lst := widget.NewList(
		func() int { return len(vm.SecretNames) },
		func() fyne.CanvasObject { return widget.NewCheck("", nil) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			chk := o.(*widget.Check)
			chk.Text = vm.SecretNames[i]
			chk.Checked = vm.SelectedSecret[i]
			chk.OnChanged = func(b bool) { vm.SelectedSecret[i] = b }
		},
	)
	return lst
}
