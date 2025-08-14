package testdata

// UITestData provides immutable data for UI testing scenarios
type UITestData struct {
	WindowTitle     string
	ButtonTexts     map[string]string
	ErrorMessages   map[string]string
	SuccessMessages map[string]string
	FormLabels      map[string]string
}

// WorkflowTestData provides data for complete workflow testing
type WorkflowTestData struct {
	Name            string
	Description     string
	Steps           []WorkflowStep
	ExpectedOutcome string
}

// WorkflowStep represents a single step in a UI workflow
type WorkflowStep struct {
	Action      string
	Target      string
	Input       string
	Expected    string
	Description string
}

// Common workflow descriptions to avoid duplication
const (
	SelectSecretFromListDesc = "Select secret from list"
	ClickEditButtonDesc      = "Click Edit button"
	SaveChangesDesc          = "Save changes"
	ClickDeleteButtonDesc    = "Click Delete button"
	ConfirmDeletionDesc      = "Confirm deletion"
)

var (
	// UIConstants provides immutable UI-related test data
	UIConstants = UITestData{
		WindowTitle: "Password Manager - E2E Test",
		ButtonTexts: map[string]string{
			"create":  "Create Secret",
			"save":    "Save",
			"cancel":  "Cancel",
			"edit":    "Edit",
			"delete":  "Delete",
			"confirm": "Confirm",
			"reveal":  "Reveal",
			"hide":    "Hide",
			"search":  "Search",
			"menu":    "☰",
		},
		ErrorMessages: map[string]string{
			"empty_name":     "Secret name cannot be empty",
			"empty_value":    "Secret value cannot be empty",
			"duplicate_name": "Secret with this name already exists",
			"invalid_chars":  "Secret name contains invalid characters",
			"save_failed":    "Failed to save secret",
			"delete_failed":  "Failed to delete secret",
			"load_failed":    "Failed to load secrets",
		},
		SuccessMessages: map[string]string{
			"secret_created": "Secret created successfully",
			"secret_updated": "Secret updated successfully",
			"secret_deleted": "Secret deleted successfully",
			"data_loaded":    "Data loaded successfully",
		},
		FormLabels: map[string]string{
			"secret_name":  "Secret Name",
			"secret_value": "Secret Value",
			"secret_type":  "Secret Type",
			"description":  "Description",
		},
	}

	// WorkflowTestScenarios provides predefined UI workflow test scenarios
	WorkflowTestScenarios = struct {
		CreateSecret     WorkflowTestData
		EditSecret       WorkflowTestData
		DeleteSecret     WorkflowTestData
		SearchSecrets    WorkflowTestData
		ToggleVisibility WorkflowTestData
	}{
		CreateSecret: WorkflowTestData{
			Name:        "CreateSecretWorkflow",
			Description: "Complete secret creation workflow",
			Steps: []WorkflowStep{
				{
					Action:      "click",
					Target:      "create_button",
					Input:       "",
					Expected:    "modal_opens",
					Description: "Click Create Secret button",
				},
				{
					Action:      "type",
					Target:      "name_field",
					Input:       TestSecrets.Simple.Name,
					Expected:    "text_entered",
					Description: "Enter secret name",
				},
				{
					Action:      "type",
					Target:      "value_field",
					Input:       TestSecrets.Simple.Value,
					Expected:    "text_entered",
					Description: "Enter secret value",
				},
				{
					Action:      "click",
					Target:      "save_button",
					Input:       "",
					Expected:    "secret_saved",
					Description: "Save the secret",
				},
			},
			ExpectedOutcome: "Secret appears in list",
		},
		EditSecret: WorkflowTestData{
			Name:        "EditSecretWorkflow",
			Description: "Complete secret editing workflow",
			Steps: []WorkflowStep{
				{
					Action:      "click",
					Target:      "secret_item",
					Input:       TestSecrets.Versioned.Name,
					Expected:    "detail_view_opens",
					Description: SelectSecretFromListDesc,
				},
				{
					Action:      "click",
					Target:      "edit_button",
					Input:       "",
					Expected:    "edit_modal_opens",
					Description: ClickEditButtonDesc,
				},
				{
					Action:      "clear_and_type",
					Target:      "value_field",
					Input:       "UpdatedValue123",
					Expected:    "text_updated",
					Description: "Update secret value",
				},
				{
					Action:      "click",
					Target:      "save_button",
					Input:       "",
					Expected:    "secret_updated",
					Description: SaveChangesDesc,
				},
			},
			ExpectedOutcome: "New version created",
		},
		DeleteSecret: WorkflowTestData{
			Name:        "DeleteSecretWorkflow",
			Description: "Complete secret deletion workflow",
			Steps: []WorkflowStep{
				{
					Action:      "click",
					Target:      "secret_item",
					Input:       TestSecrets.Temporary.Name,
					Expected:    "detail_view_opens",
					Description: SelectSecretFromListDesc,
				},
				{
					Action:      "click",
					Target:      "delete_button",
					Input:       "",
					Expected:    "confirmation_dialog",
					Description: ClickDeleteButtonDesc,
				},
				{
					Action:      "click",
					Target:      "confirm_button",
					Input:       "",
					Expected:    "secret_deleted",
					Description: ConfirmDeletionDesc,
				},
			},
			ExpectedOutcome: "Secret removed from list",
		},
		SearchSecrets: WorkflowTestData{
			Name:        "SearchSecretsWorkflow",
			Description: "Complete search functionality workflow",
			Steps: []WorkflowStep{
				{
					Action:      "click",
					Target:      "search_field",
					Input:       "",
					Expected:    "search_active",
					Description: "Click search field",
				},
				{
					Action:      "type",
					Target:      "search_field",
					Input:       "Simple",
					Expected:    "results_filtered",
					Description: "Type search term",
				},
				{
					Action:      "verify",
					Target:      "results_list",
					Input:       "",
					Expected:    "matching_secrets_only",
					Description: "Verify filtered results",
				},
				{
					Action:      "clear",
					Target:      "search_field",
					Input:       "",
					Expected:    "all_secrets_visible",
					Description: "Clear search",
				},
			},
			ExpectedOutcome: "Search filters and resets correctly",
		},
		ToggleVisibility: WorkflowTestData{
			Name:        "ToggleVisibilityWorkflow",
			Description: "Complete visibility toggle workflow",
			Steps: []WorkflowStep{
				{
					Action:      "click",
					Target:      "secret_item",
					Input:       TestSecrets.Simple.Name,
					Expected:    "detail_view_opens",
					Description: SelectSecretFromListDesc,
				},
				{
					Action:      "click",
					Target:      "reveal_button",
					Input:       "",
					Expected:    "value_visible",
					Description: "Reveal secret value",
				},
				{
					Action:      "verify",
					Target:      "value_display",
					Input:       "",
					Expected:    "actual_value_shown",
					Description: "Verify value is visible",
				},
				{
					Action:      "click",
					Target:      "hide_button",
					Input:       "",
					Expected:    "value_hidden",
					Description: "Hide secret value",
				},
			},
			ExpectedOutcome: "Value visibility toggles correctly",
		},
	}
)

// CloneUITestData returns a deep copy of UITestData
func (uid UITestData) CloneUITestData() UITestData {
	buttonTexts := make(map[string]string)
	for k, v := range uid.ButtonTexts {
		buttonTexts[k] = v
	}

	errorMessages := make(map[string]string)
	for k, v := range uid.ErrorMessages {
		errorMessages[k] = v
	}

	successMessages := make(map[string]string)
	for k, v := range uid.SuccessMessages {
		successMessages[k] = v
	}

	formLabels := make(map[string]string)
	for k, v := range uid.FormLabels {
		formLabels[k] = v
	}

	return UITestData{
		WindowTitle:     uid.WindowTitle,
		ButtonTexts:     buttonTexts,
		ErrorMessages:   errorMessages,
		SuccessMessages: successMessages,
		FormLabels:      formLabels,
	}
}

// CloneWorkflowStep returns a deep copy of WorkflowStep
func (ws WorkflowStep) CloneWorkflowStep() WorkflowStep {
	return WorkflowStep{
		Action:      ws.Action,
		Target:      ws.Target,
		Input:       ws.Input,
		Expected:    ws.Expected,
		Description: ws.Description,
	}
}

// CloneWorkflowTestData returns a deep copy of WorkflowTestData
func (wtd WorkflowTestData) CloneWorkflowTestData() WorkflowTestData {
	steps := make([]WorkflowStep, len(wtd.Steps))
	for i, step := range wtd.Steps {
		steps[i] = step.CloneWorkflowStep()
	}

	return WorkflowTestData{
		Name:            wtd.Name,
		Description:     wtd.Description,
		Steps:           steps,
		ExpectedOutcome: wtd.ExpectedOutcome,
	}
}

// GetStepCount returns the number of steps in a workflow
func (wtd WorkflowTestData) GetStepCount() int {
	return len(wtd.Steps)
}

// GetStepByIndex returns a workflow step by index
func (wtd WorkflowTestData) GetStepByIndex(index int) (WorkflowStep, bool) {
	if index < 0 || index >= len(wtd.Steps) {
		return WorkflowStep{}, false
	}
	return wtd.Steps[index].CloneWorkflowStep(), true
}

// GetAllSteps returns a copy of all workflow steps
func (wtd WorkflowTestData) GetAllSteps() []WorkflowStep {
	steps := make([]WorkflowStep, len(wtd.Steps))
	for i, step := range wtd.Steps {
		steps[i] = step.CloneWorkflowStep()
	}
	return steps
}
