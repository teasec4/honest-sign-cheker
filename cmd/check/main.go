package main

import (
	"fmt"
	"onestsignt/internal/compare"
	"onestsignt/internal/csv"
	"onestsignt/internal/excel"
	"onestsignt/internal/output"
	"os"
	"path/filepath"
)

func main() {
	os.MkdirAll("results", 0755)
	output.Init("results/output.txt")

	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	filePairs := []struct {
		csvName   string
		excelName string
	}{
		{"01OS.csv", "01OS-report.xlsx"},
		{"02OS.csv", "02OS-report.xlsx"},
		{"03OS.csv", "03OS-report.xlsx"},
	}

	for _, pair := range filePairs {
		err := processFilePair(workingDir, pair.csvName, pair.excelName)
		if err != nil {
			output.WriteLine("Ошибка: %v", err)
		}
	}

	output.Close()
	fmt.Println("Сохранено в results/output.txt")

}

func processFilePair(workingDir, csvName, excelName string) error {
	csvPath := filepath.Join(workingDir, "files", csvName)
	excelPath := filepath.Join(workingDir, "files", excelName)

	// Проверяем существование файлов
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return fmt.Errorf("CSV файл не найден: %s", csvPath)
	}
	if _, err := os.Stat(excelPath); os.IsNotExist(err) {
		return fmt.Errorf("Excel файл не найден: %s", excelPath)
	}

	// Создаем читателей
	rCSV := csv.NewReaderCSV(csvPath)
	rExcel := excel.NewReaderExcel(excelPath)

	// Создаем сравниватель и выполняем сравнение
	c := compare.NewComparerer(*rCSV, *rExcel)
	return c.Compare()
}
