package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type logEntry struct {
	Service string `json:"service"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func main() {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 5 * time.Second}

	logs := []logEntry{
		{Service: "payment-api", Level: "error", Message: "Database connection failed"},
		{Service: "payment-api", Level: "error", Message: "Timeout while querying users"},
		{Service: "payment-api", Level: "error", Message: "Retry budget exhausted"},
	}

	for _, entry := range logs {
		if err := postLog(client, baseURL, entry); err != nil {
			fmt.Fprintf(os.Stderr, "post log failed: %v\n", err)
			os.Exit(1)
		}
	}

	if err := fetchIncidents(client, baseURL); err != nil {
		fmt.Fprintf(os.Stderr, "fetch incidents failed: %v\n", err)
		os.Exit(1)
	}
}

func postLog(client *http.Client, baseURL string, entry logEntry) error {
	body, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/logs", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	fmt.Printf("POST /logs -> %d %s\n", resp.StatusCode, string(payload))
	return nil
}

func fetchIncidents(client *http.Client, baseURL string) error {
	resp, err := client.Get(baseURL + "/incidents")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	fmt.Printf("GET /incidents -> %d %s\n", resp.StatusCode, string(payload))
	return nil
}
