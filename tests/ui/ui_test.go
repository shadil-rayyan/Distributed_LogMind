package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDashboardHTMLStructure(t *testing.T) {
	// Find index.html
	paths := []string{"index.html", "../index.html", "../../index.html", "../../../index.html"}
	var indexHtmlPath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			indexHtmlPath = p
			break
		}
	}

	if indexHtmlPath == "" {
		t.Fatalf("Could not locate index.html in search paths")
	}

	contentBytes, err := os.ReadFile(indexHtmlPath)
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}

	content := string(contentBytes)

	// Validate title tag
	expectedTitle := "<title>LogMind Ops Dashboard</title>"
	if !strings.Contains(content, expectedTitle) {
		t.Errorf("Expected title %q not found in HTML", expectedTitle)
	}

	// Validate key elements exist
	requiredElements := []string{
		`id="incident-count"`,
		`id="incident-status-dot"`,
		`id="last-update"`,
		`id="incidents-table-body"`,
		`id="empty-state"`,
	}

	for _, el := range requiredElements {
		if !strings.Contains(content, el) {
			t.Errorf("Required UI element %q not found in HTML", el)
		}
	}

	// Validate JS function signatures and AJAX endpoints exist
	requiredJS := []string{
		"async function fetchIncidents()",
		"function updateDashboard(incidents)",
		"const response = await fetch('/incidents');",
		"setInterval(fetchIncidents, 2000);",
	}

	for _, js := range requiredJS {
		if !strings.Contains(content, js) {
			t.Errorf("Required JavaScript sequence %q not found in HTML script block", js)
		}
	}

	t.Logf("UI Test passed: index.html structure verified at %q", filepath.Clean(indexHtmlPath))
}
