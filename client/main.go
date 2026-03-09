package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const serverURL = "http://localhost:8080/files"

type FileMetadata struct {
	Path         string     `json:"path"`
	Size         int64      `json:"size"`
	LastModified *time.Time `json:"last_modified,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: binary-scan <directory>")
	}

	root := os.Args[1]

	err := scanDirectory(root)
	if err != nil {
		log.Fatal(err)
	}
}

func scanDirectory(root string) error {
	var metas []FileMetadata
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			log.Printf("error accessing file: %v", err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			log.Printf("could not determine relative path: %v", err)
			return nil
		}

		meta := FileMetadata{
			Path: relPath,
			Size: info.Size(),
		}

		if isBinaryExecutable(path) {
			mod := info.ModTime()
			meta.LastModified = &mod
		}

		metas = append(metas, meta)

		return nil
	})
	if err != nil {
		return err
	}
	return uploadBatch(metas)
}

func isBinaryExecutable(path string) bool {

	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	content := buffer[:n]

	for _, b := range content {
		if b == 0 {
			return true
		}
	}

	return false
}

func uploadBatch(metas []FileMetadata) error {

	body, err := json.Marshal(metas)
	if err != nil {
		return err
	}

	var lastErr error

	for i := 0; i < 3; i++ {

		resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("server returned %d", resp.StatusCode)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return lastErr
}
