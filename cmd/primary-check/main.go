package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"onestsignt/internal/codes"
	"onestsignt/internal/input"
	"onestsignt/internal/match"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	batch := flag.String("batch", "", "имя партии из data/<batch>")
	issuedPath := flag.String("issued", "", "файл с выданными кодами")
	returnedPath := flag.String("returned", "", "файл с кодами от поставщика")
	outDir := flag.String("out", "", "папка для отчетов")
	minPercent := flag.Float64("min", 85, "минимальный процент похожести для кандидата на восстановление")
	flag.Parse()

	paths, err := resolvePrimaryPaths(*batch, *issuedPath, *returnedPath, *outDir)
	if err != nil {
		fatal(err)
	}

	issued, err := input.ReadCodes(paths.issued)
	if err != nil {
		fatal(err)
	}
	returned, err := input.ReadCodes(paths.returned)
	if err != nil {
		fatal(err)
	}

	config := match.DefaultConfig()
	config.MinPercent = *minPercent
	matcher := match.NewMatcher(issued, config)
	results := matcher.MatchReturned(returned)

	if err := os.MkdirAll(paths.outDir, 0755); err != nil {
		fatal(err)
	}

	report := buildPrimaryReport(paths, issued, returned, results, config.MinPercent)
	reportPath := filepath.Join(paths.outDir, "primary-report.txt")
	if err := os.WriteFile(reportPath, []byte(report), 0644); err != nil {
		fatal(err)
	}

	reviewPath := filepath.Join(paths.outDir, "manual-review.tsv")
	if err := writeManualReview(reviewPath, results); err != nil {
		fatal(err)
	}

	fmt.Print(report)
	fmt.Printf("\nСохранено:\n  %s\n  %s\n", reportPath, reviewPath)
}

type primaryPaths struct {
	batch    string
	issued   string
	returned string
	outDir   string
}

func resolvePrimaryPaths(batch, issuedPath, returnedPath, outDir string) (primaryPaths, error) {
	paths := primaryPaths{batch: batch, issued: issuedPath, returned: returnedPath, outDir: outDir}

	if batch != "" {
		dataDir := filepath.Join("data", batch)
		var err error
		if paths.issued == "" {
			paths.issued, err = input.FindNamedFile(dataDir, "issued")
			if err != nil {
				return paths, err
			}
		}
		if paths.returned == "" {
			paths.returned, err = input.FindNamedFile(dataDir, "returned")
			if err != nil {
				return paths, err
			}
		}
		if paths.outDir == "" {
			paths.outDir = filepath.Join("reports", batch)
		}
	}

	if paths.issued == "" || paths.returned == "" {
		return paths, fmt.Errorf("укажи -batch или оба файла: -issued и -returned")
	}
	if paths.outDir == "" {
		paths.outDir = filepath.Join("reports", "manual")
	}

	return paths, nil
}

func buildPrimaryReport(paths primaryPaths, issued codes.Index, returned codes.Index, results []match.Result, minPercent float64) string {
	var builder strings.Builder
	summary := summarize(results)
	duplicates := returned.DuplicateCodes()

	title := "первичная сверка"
	if paths.batch != "" {
		title += " " + paths.batch
	}

	fmt.Fprintf(&builder, "=== %s ===\n", title)
	fmt.Fprintf(&builder, "Файлы:\n")
	fmt.Fprintf(&builder, "  Выдано: %s\n", filepath.Base(paths.issued))
	fmt.Fprintf(&builder, "  Возврат: %s\n", filepath.Base(paths.returned))
	fmt.Fprintf(&builder, "Порог похожести: %.2f%%\n", minPercent)
	fmt.Fprintf(&builder, "Итог:\n")
	fmt.Fprintf(&builder, "  Выдано: %d кодов, %d уникальных\n", issued.TotalCount(), issued.UniqueCount())
	fmt.Fprintf(&builder, "  Возврат: %d кодов, %d уникальных\n", returned.TotalCount(), returned.UniqueCount())
	fmt.Fprintf(&builder, "  Точно совпали: %d кодов, %d уникальных\n", summary.exactTotal, summary.exactUnique)
	fmt.Fprintf(&builder, "  Нужно восстановить: %d кодов, %d уникальных\n", summary.fuzzyTotal, summary.fuzzyUnique)
	fmt.Fprintf(&builder, "  Не распознаны: %d кодов, %d уникальных\n", summary.unknownTotal, summary.unknownUnique)
	fmt.Fprintf(&builder, "  Дубликаты в возврате: %d уникальных кодов\n", len(duplicates))

	if len(duplicates) > 0 {
		fmt.Fprintf(&builder, "\nДубликаты в возврате:\n")
		for index, code := range duplicates {
			fmt.Fprintf(&builder, "%d. %s\n", index+1, code)
			fmt.Fprintf(&builder, "   Что: код встречается %d раза\n", returned.Count(code))
			fmt.Fprintf(&builder, "   Где:\n")
			for _, location := range returned.Locations(code) {
				fmt.Fprintf(&builder, "     - %s\n", location.String())
			}
		}
	}

	writeResultSection(&builder, "Кандидаты на ручное восстановление", results, match.StatusFuzzy)
	writeResultSection(&builder, "Не распознаны", results, match.StatusUnknown)

	if summary.fuzzyUnique == 0 && summary.unknownUnique == 0 && len(duplicates) == 0 {
		fmt.Fprintf(&builder, "\nOK\n")
	}

	return builder.String()
}

type resultSummary struct {
	exactTotal    int
	exactUnique   int
	fuzzyTotal    int
	fuzzyUnique   int
	unknownTotal  int
	unknownUnique int
}

func summarize(results []match.Result) resultSummary {
	var summary resultSummary
	for _, result := range results {
		count := len(result.ReturnedPlaces)
		switch result.Status {
		case match.StatusExact:
			summary.exactTotal += count
			summary.exactUnique++
		case match.StatusFuzzy:
			summary.fuzzyTotal += count
			summary.fuzzyUnique++
		case match.StatusUnknown:
			summary.unknownTotal += count
			summary.unknownUnique++
		}
	}
	return summary
}

func writeResultSection(builder *strings.Builder, title string, results []match.Result, status string) {
	var section []match.Result
	for _, result := range results {
		if result.Status == status {
			section = append(section, result)
		}
	}
	if len(section) == 0 {
		return
	}

	fmt.Fprintf(builder, "\n%s:\n", title)
	for index, result := range section {
		fmt.Fprintf(builder, "%d. [%s] %s\n", index+1, result.Status, result.ReturnedCode)
		fmt.Fprintf(builder, "   Совпадение: %.2f%%\n", result.MatchPercent)
		if result.MatchedCode != "" {
			fmt.Fprintf(builder, "   Похоже на: %s\n", result.MatchedCode)
			fmt.Fprintf(builder, "   Где в выдаче:\n")
			for _, location := range result.MatchedPlaces {
				fmt.Fprintf(builder, "     - %s\n", location.String())
			}
		}
		fmt.Fprintf(builder, "   Где в возврате:\n")
		for _, location := range result.ReturnedPlaces {
			fmt.Fprintf(builder, "     - %s\n", location.String())
		}
	}
}

func writeManualReview(path string, results []match.Result) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	defer writer.Flush()

	if err := writer.Write([]string{
		"status",
		"match_percent",
		"returned_code",
		"best_issued_code",
		"returned_locations",
		"issued_locations",
		"restored_code",
		"comment",
	}); err != nil {
		return err
	}

	for _, result := range results {
		if result.Status == match.StatusExact {
			continue
		}
		if err := writer.Write([]string{
			result.Status,
			fmt.Sprintf("%.2f", result.MatchPercent),
			result.ReturnedCode,
			result.MatchedCode,
			joinLocations(result.ReturnedPlaces),
			joinLocations(result.MatchedPlaces),
			result.MatchedCode,
			"",
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func joinLocations(locations []codes.Location) string {
	parts := make([]string, 0, len(locations))
	for _, location := range locations {
		parts = append(parts, location.String())
	}
	return strings.Join(parts, " | ")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Ошибка:", err)
	os.Exit(1)
}
