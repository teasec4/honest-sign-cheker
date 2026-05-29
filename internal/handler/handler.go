package handler

import (
	"net/http"
	"onestsignt/internal/checker"
	"os"
	"strconv"
	"strings"
)
const maxUploadSize = 512 << 20

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandlePrimaryCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "разрешен только POST")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	tempDir, err := os.MkdirTemp("", "onestsignt-api-*")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	issuedPath, err := SaveUploadedFile(r, tempDir, "issued")
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	returnedPath, err := SaveUploadedFile(r, tempDir, "returned")
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	minPercent := 85.0
	if raw := strings.TrimSpace(r.FormValue("minPercent")); raw != "" {
		minPercent, err = strconv.ParseFloat(raw, 64)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "некорректный процент совпадения")
			return
		}
	}

	report, err := checker.RunPrimary(issuedPath, returnedPath, minPercent)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, report)
}

func HandleDuplicateCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "разрешен только POST")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	tempDir, err := os.MkdirTemp("", "onestsignt-api-*")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	restoredPath, err := SaveUploadedFile(r, tempDir, "restored")
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	report, err := checker.RunDuplicates(restoredPath)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, report)
}