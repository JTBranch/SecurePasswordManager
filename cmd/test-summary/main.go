package main

import (
	"log"
	"path/filepath"

	"go-password-manager/tests/reporting"
)

func main() {
	// Get the output directory relative to project root
	outputDir := filepath.Join("tmp", "output")

	log.Println("Generating test summary...")
	if err := reporting.CreateTestSummary(outputDir); err != nil {
		log.Printf("Failed to create test summary: %v", err)
	} else {
		log.Println("Test summary generated successfully")
	}
}
