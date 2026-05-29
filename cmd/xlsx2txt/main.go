package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func main() {
	os.MkdirAll("results", 0755)

	filesDir := "files"
	entries, err := os.ReadDir(filesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "не удалось прочитать папку %s: %v\n", filesDir, err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".xlsx") {
			continue
		}

		xlsxPath := filepath.Join(filesDir, entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		txtPath := filepath.Join("results", baseName+".txt")

		count, err := convertXlsxToTxt(xlsxPath, txtPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ошибка обработки %s: %v\n", entry.Name(), err)
			continue
		}
		fmt.Printf("%s → %s (%d кодов)\n", entry.Name(), txtPath, count)
	}
}

// convertXlsxToTxt читает колонку "DM CODE" из xlsx и пишет коды построчно в txt.
func convertXlsxToTxt(xlsxPath, txtPath string) (int, error) {
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		return 0, fmt.Errorf("открытие xlsx: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return 0, fmt.Errorf("в файле нет листов")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return 0, fmt.Errorf("чтение строк: %w", err)
	}

	if len(rows) == 0 {
		return 0, fmt.Errorf("файл пуст")
	}

	// Ищем колонку "DM CODE" в заголовке (первая строка)
	colIdx := findDMCodeColumn(rows[0])
	if colIdx < 0 {
		return 0, fmt.Errorf("колонка 'DM CODE' не найдена в заголовке")
	}

	// Собираем все непустые коды из колонки
	var codes []string
	for i := 1; i < len(rows); i++ {
		if colIdx >= len(rows[i]) {
			continue
		}
		code := strings.TrimSpace(rows[i][colIdx])
		if code == "" {
			continue
		}
		codes = append(codes, code)
	}

	// Пишем в txt — каждый код с новой строки
	out, err := os.Create(txtPath)
	if err != nil {
		return 0, fmt.Errorf("создание txt: %w", err)
	}
	defer out.Close()

	for _, code := range codes {
		fmt.Fprintln(out, code)
	}

	return len(codes), nil
}

// findDMCodeColumn ищет колонку с заголовком "DM CODE" (без учёта регистра и лишних пробелов).
func findDMCodeColumn(header []string) int {
	for i, cell := range header {
		h := strings.TrimSpace(cell)
		h = strings.ToUpper(h)
		h = strings.Join(strings.Fields(h), " ")
		if h == "DM CODE" {
			return i
		}
	}
	return -1
}
