package main

import (
	"flag"
	"fmt"
	"onestsignt/internal/codes"
	"onestsignt/internal/input"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	batch := flag.String("batch", "", "имя партии из data/<batch>")
	inputPath := flag.String("input", "", "восстановленный файл для проверки")
	outDir := flag.String("out", "", "папка для отчета")
	flag.Parse()

	paths, err := resolveDuplicatePaths(*batch, *inputPath, *outDir)
	if err != nil {
		fatal(err)
	}

	index, err := input.ReadCodes(paths.input)
	if err != nil {
		fatal(err)
	}

	if err := os.MkdirAll(paths.outDir, 0755); err != nil {
		fatal(err)
	}

	report := buildDuplicateReport(paths, index)
	reportPath := filepath.Join(paths.outDir, "duplicates-report.txt")
	if err := os.WriteFile(reportPath, []byte(report), 0644); err != nil {
		fatal(err)
	}

	fmt.Print(report)
	fmt.Printf("\nСохранено:\n  %s\n", reportPath)
}

type duplicatePaths struct {
	batch  string
	input  string
	outDir string
}

func resolveDuplicatePaths(batch, inputPath, outDir string) (duplicatePaths, error) {
	paths := duplicatePaths{batch: batch, input: inputPath, outDir: outDir}

	if batch != "" {
		dataDir := filepath.Join("data", batch)
		var err error
		if paths.input == "" {
			paths.input, err = input.FindNamedFile(dataDir, "restored")
			if err != nil {
				return paths, err
			}
		}
		if paths.outDir == "" {
			paths.outDir = filepath.Join("reports", batch)
		}
	}

	if paths.input == "" {
		return paths, fmt.Errorf("укажи -batch или файл через -input")
	}
	if paths.outDir == "" {
		paths.outDir = filepath.Join("reports", "manual")
	}

	return paths, nil
}

func buildDuplicateReport(paths duplicatePaths, index codes.Index) string {
	var builder strings.Builder
	duplicates := index.DuplicateCodes()

	title := "проверка дубликатов"
	if paths.batch != "" {
		title += " " + paths.batch
	}

	fmt.Fprintf(&builder, "=== %s ===\n", title)
	fmt.Fprintf(&builder, "Файл: %s\n", filepath.Base(paths.input))
	fmt.Fprintf(&builder, "Итог:\n")
	fmt.Fprintf(&builder, "  Всего кодов: %d\n", index.TotalCount())
	fmt.Fprintf(&builder, "  Уникальных: %d\n", index.UniqueCount())
	fmt.Fprintf(&builder, "  Дубликатов: %d уникальных кодов\n", len(duplicates))

	if len(duplicates) == 0 {
		fmt.Fprintf(&builder, "OK\n")
		return builder.String()
	}

	fmt.Fprintf(&builder, "\nПроблемы:\n")
	for indexNumber, code := range duplicates {
		fmt.Fprintf(&builder, "%d. [ДУБЛИКАТ] %s\n", indexNumber+1, code)
		fmt.Fprintf(&builder, "   Что: код встречается %d раза\n", index.Count(code))
		fmt.Fprintf(&builder, "   Где:\n")
		for _, location := range index.Locations(code) {
			fmt.Fprintf(&builder, "     - %s\n", location.String())
		}
	}

	return builder.String()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Ошибка:", err)
	os.Exit(1)
}
