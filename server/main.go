package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type FileMetadata struct {
	Path         string     `json:"path"`
	Size         int64      `json:"size"`
	LastModified *time.Time `json:"last_modified,omitempty"`
}

type StoredFile struct {
	ID           int        `json:"id"`
	Path         string     `json:"path"`
	Size         int64      `json:"size"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	UploadedAt   time.Time  `json:"uploaded_at"`
}

var db *sql.DB

func main() {

	var err error

	db, err = sql.Open("sqlite3", "./files.db")
	if err != nil {
		log.Fatal(err)
	}

	err = createTable()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/files", filesHandler)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable() error {

	query := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT,
		size INTEGER,
		last_modified DATETIME,
		uploaded_at DATETIME
	)`

	_, err := db.Exec(query)
	return err
}

func filesHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		handleUpload(w, r)

	case http.MethodGet:
		handleList(w, r)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {

	var metas []FileMetadata

	err := json.NewDecoder(r.Body).Decode(&metas)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	for _, meta := range metas {
		uploadedAt := time.Now()

		query := `
		INSERT INTO files(path, size, last_modified, uploaded_at)
		VALUES (?, ?, ?, ?)`

		_, err = db.Exec(query, meta.Path, meta.Size, meta.LastModified, uploadedAt)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func handleList(w http.ResponseWriter, r *http.Request) {

	limit := 20

	queryLimit := r.URL.Query().Get("limit")
	if queryLimit != "" {
		if parsed, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsed
		}
	}

	query := `
	SELECT id, path, size, last_modified, uploaded_at
	FROM files
	ORDER BY uploaded_at DESC
	LIMIT ?`

	rows, err := db.Query(query, limit)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	files := []StoredFile{}

	for rows.Next() {

		var f StoredFile

		err := rows.Scan(
			&f.ID,
			&f.Path,
			&f.Size,
			&f.LastModified,
			&f.UploadedAt,
		)

		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		files = append(files, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
