package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

func SaveUploadedFile(r *http.Request, tempDir string, fieldName string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", fmt.Errorf("не выбран файл %q", fieldName)
	}
	defer file.Close()

	extension := strings.ToLower(filepath.Ext(header.Filename))
	if extension == "" {
		return "", fmt.Errorf("у файла %q нет расширения", header.Filename)
	}

	path := filepath.Join(tempDir, fieldName+extension)
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	return path, nil
}