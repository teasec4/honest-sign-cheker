package input

import (
	"fmt"
	"onestsignt/internal/codes"
	csvreader "onestsignt/internal/csv"
	excelreader "onestsignt/internal/excel"
	"os"
	"path/filepath"
	"strings"
)

var SupportedExtensions = []string{".csv", ".txt", ".xlsx", ".xlsm"}

func ReadCodes(path string) (codes.Index, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".csv", ".txt":
		return csvreader.NewReaderCSV(path).ReadCSV()
	case ".xlsx", ".xlsm":
		return excelreader.NewReaderExcel(path).ReadExcel()
	default:
		return codes.Index{}, fmt.Errorf("неподдерживаемый формат файла %q", path)
	}
}

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
