package reporting

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// TestReporter handles test reporting and log capture
type TestReporter struct {
	OutputDir      string
	TestName       string
	LogFile        *os.File
	originalLogger *log.Logger
	originalOutput io.Writer
}

// TestReport represents a test execution report
type TestReport struct {
	TestName    string                 `json:"test_name"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Status      string                 `json:"status"`
	Environment map[string]string      `json:"environment"`
	Logs        []LogEntry             `json:"logs"`
	Artifacts   []string               `json:"artifacts"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// NewTestReporter creates a new test reporter
func NewTestReporter(testName string) (*TestReporter, error) {
	// Ensure tmp/output directory exists
	outputDir := filepath.Join("tmp", "output", "test-reports")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create test-specific log file
	logFileName := fmt.Sprintf("%s-%d.log", testName, time.Now().UnixNano())
	logFilePath := filepath.Join(outputDir, logFileName)

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &TestReporter{
		OutputDir: outputDir,
		TestName:  testName,
		LogFile:   logFile,
	}, nil
}

// StartCapture begins capturing application logs
func (tr *TestReporter) StartCapture() error {
	// Capture the original logger settings
	tr.originalLogger = log.Default()
	tr.originalOutput = log.Writer()

	// Create a multi-writer to write to both original output and our log file
	multiWriter := io.MultiWriter(tr.originalOutput, tr.LogFile)
	log.SetOutput(multiWriter)

	// Log the start of capture
	log.Printf("[TEST-REPORTER] Starting log capture for test: %s", tr.TestName)

	return nil
}

// StopCapture stops capturing logs and generates the report
func (tr *TestReporter) StopCapture() error {
	if tr.LogFile != nil {
		log.Printf("[TEST-REPORTER] Stopping log capture for test: %s", tr.TestName)

		// Restore original log output
		if tr.originalOutput != nil {
			log.SetOutput(tr.originalOutput)
		}

		tr.LogFile.Close()
	}
	return nil
}

// GenerateReport creates a comprehensive test report
func (tr *TestReporter) GenerateReport(status string, startTime, endTime time.Time, metadata map[string]interface{}) error {
	report := TestReport{
		TestName:  tr.TestName,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
		Status:    status,
		Environment: map[string]string{
			"GOOS":       os.Getenv("GOOS"),
			"GOARCH":     os.Getenv("GOARCH"),
			"GO_VERSION": os.Getenv("GO_VERSION"),
			"CI":         os.Getenv("CI"),
			"GITHUB_SHA": os.Getenv("GITHUB_SHA"),
			"GITHUB_REF": os.Getenv("GITHUB_REF"),
		},
		Artifacts: []string{},
		Metadata:  metadata,
	}

	// Find artifacts in the output directory
	artifacts, err := tr.collectArtifacts()
	if err == nil {
		report.Artifacts = artifacts
	}

	// Save report as JSON
	reportPath := filepath.Join(tr.OutputDir, fmt.Sprintf("%s-report.json", tr.TestName))
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer reportFile.Close()

	encoder := json.NewEncoder(reportFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	return nil
}

// collectArtifacts finds all artifacts related to this test
func (tr *TestReporter) collectArtifacts() ([]string, error) {
	var artifacts []string

	// Look for files related to this test
	pattern := filepath.Join(tr.OutputDir, tr.TestName+"*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		rel, err := filepath.Rel(tr.OutputDir, match)
		if err == nil {
			artifacts = append(artifacts, rel)
		}
	}

	return artifacts, nil
}

// CaptureScreenshot captures a screenshot (placeholder for future implementation)
func (tr *TestReporter) CaptureScreenshot(name string) error {
	// Placeholder for screenshot capture functionality
	// This would be implemented when we add screenshot capabilities to E2E tests
	screenshotPath := filepath.Join(tr.OutputDir, fmt.Sprintf("%s-%s-screenshot.png", tr.TestName, name))

	// For now, just create a placeholder file
	file, err := os.Create(screenshotPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Screenshot placeholder for %s at %s\n", name, time.Now().Format(time.RFC3339)))
	return err
}

// LogEvent logs a test event with structured data
func (tr *TestReporter) LogEvent(level, message string, data map[string]interface{}) {
	// Log to standard logger
	logMessage := fmt.Sprintf("[%s] %s", level, message)
	if len(data) > 0 {
		dataJSON, _ := json.Marshal(data)
		logMessage += fmt.Sprintf(" | Data: %s", string(dataJSON))
	}

	log.Println(logMessage)
}

// CreateTestSummary creates an HTML summary of all test reports
func CreateTestSummary(outputDir string) error {
	summaryPath := filepath.Join(outputDir, "test-summary.html")

	// Read all test report files
	reports := []TestReport{}
	pattern := filepath.Join(outputDir, "test-reports", "*-report.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, match := range matches {
		var report TestReport
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	// Generate HTML summary
	html := generateHTMLSummary(reports)

	return os.WriteFile(summaryPath, []byte(html), 0644)
}

// generateHTMLSummary creates an HTML report from test reports
func generateHTMLSummary(reports []TestReport) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Test Execution Summary</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .test-report { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .passed { border-left: 5px solid #4CAF50; }
        .failed { border-left: 5px solid #f44336; }
        .metadata { background: #f9f9f9; padding: 10px; margin: 10px 0; }
        .artifacts { margin: 10px 0; }
        .artifact-link { display: block; margin: 5px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Test Execution Summary</h1>
        <p>Generated: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
        <p>Total Tests: ` + fmt.Sprintf("%d", len(reports)) + `</p>
    </div>`

	for _, report := range reports {
		statusClass := "passed"
		if report.Status != "PASSED" {
			statusClass = "failed"
		}

		html += fmt.Sprintf(`
    <div class="test-report %s">
        <h3>%s - %s</h3>
        <p><strong>Duration:</strong> %s</p>
        <p><strong>Start:</strong> %s</p>
        <p><strong>End:</strong> %s</p>
        
        <div class="metadata">
            <h4>Environment</h4>`,
			statusClass, report.TestName, report.Status,
			report.Duration, report.StartTime.Format("15:04:05"), report.EndTime.Format("15:04:05"))

		for key, value := range report.Environment {
			if value != "" {
				html += fmt.Sprintf("<p><strong>%s:</strong> %s</p>", key, value)
			}
		}

		html += `</div>`

		if len(report.Artifacts) > 0 {
			html += `<div class="artifacts"><h4>Artifacts</h4>`
			for _, artifact := range report.Artifacts {
				html += fmt.Sprintf(`<a href="test-reports/%s" class="artifact-link">📁 %s</a>`, artifact, artifact)
			}
			html += `</div>`
		}

		html += `</div>`
	}

	html += `</body></html>`
	return html
}
