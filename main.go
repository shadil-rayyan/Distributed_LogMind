package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Log struct {
	Service   string `json:"service"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp int64 `json:"timestamp"`
}

type Incident struct {
	Service string `json:"service"`
	Type string `json:"type"`
	Severity string `json:"severity"`
	Message string `json:"message"`
}


var db *sql.DB

var mu sync.Mutex


// ---------------- Incident API ----------------
func incidentHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT service, level, timestamp 
		FROM logs
	`)
	if err != nil {
		http.Error(w, "DB read error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	now := time.Now().Unix()
	window :=int64(60)
	errorCount := make(map[string]int)
	
	for rows.Next() {
		var service, level string
		var timestamp int64

		err := rows.Scan(&service, &level, &timestamp)
		if err != nil {
			continue
		}

		if(now-timestamp > window){
			continue
		}

		if level == "error" {
			errorCount[service]++
		}
	}

	incidents := []Incident{}

	for service, count := range errorCount {
		if count >= 3 {
			incidents = append(incidents, Incident{
				Service:  service,
				Type:     "error_spike",
				Severity: "high",
				Message:  "too many errors detected in last 1 minute",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"window_seconds":window,
		"incidents": incidents,
	})
}
	


// ---------------- INSERT LOG ----------------

func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var log Log
	err := json.NewDecoder(r.Body).Decode(&log)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(
		"INSERT INTO logs(service, level, message, timestamp) VALUES(?,?,?,?)",
		log.Service,
		log.Level,
		log.Message,
		log.Timestamp = time.Now().Unix()
	)

	if err != nil {
		http.Error(w, "DB insert failed", http.StatusInternalServerError)
		return
	}

	fmt.Println("stored log from", log.Service)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("log stored"))
}

// ---------------- GET LOGS ----------------

func getLogsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
		"SELECT service, level, message, timestamp FROM logs",
	)
	if err != nil {
		http.Error(w, "DB read error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []Log

	for rows.Next() {
		var log Log

		err := rows.Scan(
			&log.Service,
			&log.Level,
			&log.Message,
			&log.Timestamp,
		)

		if err != nil {
			continue
		}

		result = append(result, log)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ---------------- DB INIT ----------------

func initDB() {
	var err error

	db, err = sql.Open("sqlite3", "./logmind.db")
	if err != nil {
		panic(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service TEXT,
		level TEXT,
		message TEXT,
		timestamp INTEGER
	);
	`

	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}
}

// ---------------- MAIN ----------------

func main() {
	initDB()

	http.HandleFunc("/logs", logHandler)
	http.HandleFunc("/logs/all", getLogsHandler)
	http.HandleFunc("/incidents", incidentHandler)
	fmt.Println("server starting on 8080")
	http.ListenAndServe(":8080", nil)
}
