// Package input — чтение кодов из файлов (CSV, TXT, XLSX, XLSM).
// Диспетчер ReadCodes() направляет в нужный ридер по расширению.
package input

import (
	"fmt"
	"onestsignt/internal/codes"
	"os"
	"path/filepath"
	"strings"
)

var SupportedExtensions = []string{".csv", ".txt", ".xlsx", ".xlsm"}

// ReadCodes — прочитать файл и вернуть Index с кодами.
// Выбирает ридер по расширению: .csv/.txt → ReaderCSV, .xlsx/.xlsm → ReaderExcel.
func ReadCodes(path string) (codes.Index, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".csv", ".txt":
		return NewReaderCSV(path).ReadCSV()
	case ".xlsx", ".xlsm":
		return NewReaderExcel(path).ReadExcel()
	default:
		return codes.Index{}, fmt.Errorf("неподдерживаемый формат файла %q", path)
	}
}

// FindNamedFile — найти файл baseName.* в dir (перебирает SupportedExtensions).
// Используется когда известно имя без расширения, например "issued" → "issued.csv".
func FindNamedFile(dir string, baseName string) (string, error) {
	var matches []string
	for _, extension := range SupportedExtensions {
		path := filepath.Join(dir, baseName+extension)
		if _, err := filepath.Abs(path); err != nil {
			return "", err
		}
		if exists(path) {
			matches = append(matches, path)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("в %s не найден файл %s%s", dir, baseName, strings.Join(SupportedExtensions, "|"))
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("в %s найдено несколько файлов %s.*: %s", dir, baseName, strings.Join(matches, ", "))
	}
	return matches[0], nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
