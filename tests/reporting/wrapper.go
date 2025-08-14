package reporting

import (
	"fmt"
	"testing"
	"time"
)

// TestWrapper wraps a test function with comprehensive reporting
type TestWrapper struct {
	reporter  *TestReporter
	testName  string
	startTime time.Time
	t         *testing.T
}

// NewTestWrapper creates a new test wrapper with reporting
func NewTestWrapper(t *testing.T, testName string) (*TestWrapper, error) {
	reporter, err := NewTestReporter(testName)
	if err != nil {
		return nil, err
	}

	wrapper := &TestWrapper{
		reporter:  reporter,
		testName:  testName,
		startTime: time.Now(),
		t:         t,
	}

	// Start log capture
	if err := reporter.StartCapture(); err != nil {
		t.Logf("Warning: Failed to start log capture: %v", err)
	}

	// Log test start
	reporter.LogEvent("INFO", fmt.Sprintf("Starting test: %s", testName), map[string]interface{}{
		"test_name":  testName,
		"start_time": wrapper.startTime,
	})

	return wrapper, nil
}

// Finish completes the test and generates the final report
func (tw *TestWrapper) Finish() {
	endTime := time.Now()
	status := "PASSED"

	// Check if test failed
	if tw.t.Failed() {
		status = "FAILED"
	}

	// Log test completion
	tw.reporter.LogEvent("INFO", fmt.Sprintf("Test completed: %s", tw.testName), map[string]interface{}{
		"test_name": tw.testName,
		"status":    status,
		"duration":  endTime.Sub(tw.startTime).String(),
	})

	// Stop log capture
	if err := tw.reporter.StopCapture(); err != nil {
		tw.t.Logf("Warning: Failed to stop log capture: %v", err)
	}

	// Generate report
	metadata := map[string]interface{}{
		"test_framework": "go_test",
		"test_type":      "e2e",
	}

	if err := tw.reporter.GenerateReport(status, tw.startTime, endTime, metadata); err != nil {
		tw.t.Logf("Warning: Failed to generate test report: %v", err)
	}
}

// LogStep logs a test step
func (tw *TestWrapper) LogStep(step string, data map[string]interface{}) {
	tw.reporter.LogEvent("STEP", step, data)
}

// LogError logs an error
func (tw *TestWrapper) LogError(err error, context string) {
	tw.reporter.LogEvent("ERROR", fmt.Sprintf("%s: %v", context, err), map[string]interface{}{
		"error":   err.Error(),
		"context": context,
	})
}

// LogInfo logs informational message
func (tw *TestWrapper) LogInfo(message string, data map[string]interface{}) {
	tw.reporter.LogEvent("INFO", message, data)
}

// CaptureScreenshot captures a screenshot at this point in the test
func (tw *TestWrapper) CaptureScreenshot(name string) {
	if err := tw.reporter.CaptureScreenshot(name); err != nil {
		tw.t.Logf("Warning: Failed to capture screenshot '%s': %v", name, err)
	}
}

// WithReporting is a helper function to wrap test execution with reporting
func WithReporting(t *testing.T, testName string, testFunc func(*TestWrapper)) {
	wrapper, err := NewTestWrapper(t, testName)
	if err != nil {
		t.Fatalf("Failed to create test wrapper: %v", err)
	}
	defer wrapper.Finish()

	testFunc(wrapper)
}
