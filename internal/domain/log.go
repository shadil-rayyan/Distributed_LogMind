package domain

// Log represents a single log entry.
type Log struct {
	Service   string `json:"service"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// Incident represents an aggregated event based on log patterns.
type Incident struct {
	Service   string `json:"service"`
	Type      string `json:"type"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Count     int    `json:"error_count"`
	UpdatedAt int64  `json:"updated_at"`
}
