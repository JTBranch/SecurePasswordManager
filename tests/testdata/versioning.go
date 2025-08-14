package testdata

// VersionTestData provides immutable data for secret versioning tests
type VersionTestData struct {
	SecretName  string
	Versions    []string
	Description string
}

// UpdateSequence represents a sequence of updates for version testing
type UpdateSequence struct {
	Name        string
	Description string
	Steps       []VersionTestData
}

var (
	// VersioningTestData provides test data for versioning scenarios
	VersioningTestData = struct {
		SimpleVersioning   VersionTestData
		MultipleVersions   VersionTestData
		LongVersionHistory VersionTestData
	}{
		SimpleVersioning: VersionTestData{
			SecretName:  "SimpleVersionTest",
			Versions:    []string{"InitialValue", "UpdatedValue"},
			Description: "Simple two-version test scenario",
		},
		MultipleVersions: VersionTestData{
			SecretName: "MultiVersionTest",
			Versions: []string{
				"Version1_Initial",
				"Version2_FirstUpdate",
				"Version3_SecondUpdate",
				"Version4_FinalUpdate",
			},
			Description: "Multiple version test scenario",
		},
		LongVersionHistory: VersionTestData{
			SecretName: "LongHistoryTest",
			Versions: []string{
				"V1_Original",
				"V2_Patch",
				"V3_MinorUpdate",
				"V4_MajorUpdate",
				"V5_SecurityUpdate",
				"V6_FeatureAddition",
				"V7_BugFix",
				"V8_Performance",
				"V9_Final",
				"V10_LongTermSupport",
			},
			Description: "Long version history test scenario",
		},
	}

	// UpdateSequences provides predefined update sequences for testing
	UpdateSequences = struct {
		Basic    UpdateSequence
		Advanced UpdateSequence
	}{
		Basic: UpdateSequence{
			Name:        "BasicUpdateSequence",
			Description: "Basic secret update sequence",
			Steps: []VersionTestData{
				VersioningTestData.SimpleVersioning,
			},
		},
		Advanced: UpdateSequence{
			Name:        "AdvancedUpdateSequence",
			Description: "Advanced multi-secret update sequence",
			Steps: []VersionTestData{
				VersioningTestData.MultipleVersions,
				VersioningTestData.LongVersionHistory,
			},
		},
	}
)

// CloneVersionTestData returns a deep copy of VersionTestData
func (vtd VersionTestData) CloneVersionTestData() VersionTestData {
	versions := make([]string, len(vtd.Versions))
	copy(versions, vtd.Versions)

	return VersionTestData{
		SecretName:  vtd.SecretName,
		Versions:    versions,
		Description: vtd.Description,
	}
}

// CloneUpdateSequence returns a deep copy of UpdateSequence
func (us UpdateSequence) CloneUpdateSequence() UpdateSequence {
	steps := make([]VersionTestData, len(us.Steps))
	for i, step := range us.Steps {
		steps[i] = step.CloneVersionTestData()
	}

	return UpdateSequence{
		Name:        us.Name,
		Description: us.Description,
		Steps:       steps,
	}
}

// GetInitialVersion returns the first version value
func (vtd VersionTestData) GetInitialVersion() string {
	if len(vtd.Versions) == 0 {
		return ""
	}
	return vtd.Versions[0]
}

// GetLatestVersion returns the last version value
func (vtd VersionTestData) GetLatestVersion() string {
	if len(vtd.Versions) == 0 {
		return ""
	}
	return vtd.Versions[len(vtd.Versions)-1]
}

// GetVersionCount returns the number of versions
func (vtd VersionTestData) GetVersionCount() int {
	return len(vtd.Versions)
}

// GetVersionByIndex returns a version by index (0-based)
func (vtd VersionTestData) GetVersionByIndex(index int) (string, bool) {
	if index < 0 || index >= len(vtd.Versions) {
		return "", false
	}
	return vtd.Versions[index], true
}

// GetAllVersions returns a copy of all versions
func (vtd VersionTestData) GetAllVersions() []string {
	versions := make([]string, len(vtd.Versions))
	copy(versions, vtd.Versions)
	return versions
}
