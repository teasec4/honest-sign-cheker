package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"onestsignt/internal/checker"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const maxUploadSize = 512 << 20

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "адрес API-сервера")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/primary-check", handlePrimaryCheck)
	mux.HandleFunc("/api/duplicate-check", handleDuplicateCheck)

	fmt.Printf("API-сервер запущен: http://%s\n", *addr)
	if err := http.ListenAndServe(*addr, withCORS(mux)); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handlePrimaryCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "разрешен только POST")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tempDir, err := os.MkdirTemp("", "onestsignt-api-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	issuedPath, err := saveUploadedFile(r, tempDir, "issued")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	returnedPath, err := saveUploadedFile(r, tempDir, "returned")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	minPercent := 85.0
	if raw := strings.TrimSpace(r.FormValue("minPercent")); raw != "" {
		minPercent, err = strconv.ParseFloat(raw, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "некорректный процент совпадения")
			return
		}
	}

	report, err := checker.RunPrimary(issuedPath, returnedPath, minPercent)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func handleDuplicateCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "разрешен только POST")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tempDir, err := os.MkdirTemp("", "onestsignt-api-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	restoredPath, err := saveUploadedFile(r, tempDir, "restored")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	report, err := checker.RunDuplicates(restoredPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func saveUploadedFile(r *http.Request, tempDir string, fieldName string) (string, error) {
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
