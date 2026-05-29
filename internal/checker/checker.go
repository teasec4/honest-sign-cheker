// Package checker — основная логика проверок: сверка выдано/возврат и поиск дубликатов.
// Используется как из CLI (cmd/primary-check, cmd/duplicate-check), так и из API (handler).
package checker

import (
	"fmt"
	"onestsignt/internal/codes"
	"onestsignt/internal/input"
	"onestsignt/internal/match"
)

type FileSummary struct {
	Total       int                `json:"total"`
	Unique      int                `json:"unique"`
	Diagnostics *codes.Diagnostics `json:"diagnostics,omitempty"`
}

type PrimarySummary struct {
	Issued          FileSummary `json:"issued"`
	Returned        FileSummary `json:"returned"`
	ExactTotal      int         `json:"exactTotal"`
	ExactUnique     int         `json:"exactUnique"`
	FuzzyTotal      int         `json:"fuzzyTotal"`
	FuzzyUnique     int         `json:"fuzzyUnique"`
	UnknownTotal    int         `json:"unknownTotal"`
	UnknownUnique   int         `json:"unknownUnique"`
	DuplicateUnique int         `json:"duplicateUnique"`
}

type Problem struct {
	Type              string   `json:"type"`
	Code              string   `json:"code"`
	Description       string   `json:"description"`
	Count             int      `json:"count,omitempty"`
	MatchPercent      float64  `json:"matchPercent,omitempty"`
	MatchedCode       string   `json:"matchedCode,omitempty"`
	ReturnedLocations []string `json:"returnedLocations,omitempty"`
	IssuedLocations   []string `json:"issuedLocations,omitempty"`
}

type PrimaryReport struct {
	MinPercent float64        `json:"minPercent"`
	Summary    PrimarySummary `json:"summary"`
	Duplicates []Problem      `json:"duplicates"`
	Fuzzy      []Problem      `json:"fuzzy"`
	Unknown    []Problem      `json:"unknown"`
}

type DuplicateReport struct {
	Summary    FileSummary `json:"summary"`
	Duplicates []Problem   `json:"duplicates"`
}

// RunPrimary — сверка «выдали ↔ вернули».
//  1. Читаем оба файла через input.ReadCodes() → два Index.
//  2. Строим Matcher по выданным кодам (n-граммный индекс).
//  3. Для каждого возвращённого кода: точное / похожее / неизвестное.
//  4. Отдельно ищем дубликаты в возврате.
func RunPrimary(issuedPath, returnedPath string, minPercent float64) (PrimaryReport, error) {
	// Читаем оба файла.
	issued, err := input.ReadCodes(issuedPath)
	if err != nil {
		return PrimaryReport{}, fmt.Errorf("выданные коды: %w", err)
	}
	returned, err := input.ReadCodes(returnedPath)
	if err != nil {
		return PrimaryReport{}, fmt.Errorf("возврат поставщика: %w", err)
	}

	config := match.DefaultConfig()
	if minPercent > 0 {
		config.MinPercent = minPercent
	}

	// Сопоставляем возвращённые коды с выданными.
	results := match.NewMatcher(issued, config).MatchReturned(returned)
	report := PrimaryReport{
		MinPercent: config.MinPercent,
		Summary: PrimarySummary{
			Issued: FileSummary{
				Total:       issued.TotalCount(),
				Unique:      issued.UniqueCount(),
				Diagnostics: issued.Diagnostics(),
			},
			Returned: FileSummary{
				Total:       returned.TotalCount(),
				Unique:      returned.UniqueCount(),
				Diagnostics: returned.Diagnostics(),
			},
		},
		Duplicates: []Problem{},
		Fuzzy:      []Problem{},
		Unknown:    []Problem{},
	}

	// Дубликаты: коды, которые встретились >1 раза в возврате.
	for _, code := range returned.DuplicateCodes() {
		report.Duplicates = append(report.Duplicates, Problem{
			Type:              "ДУБЛИКАТ В ВОЗВРАТЕ",
			Code:              code,
			Description:       fmt.Sprintf("код встречается в файле поставщика %d раза", returned.Count(code)),
			Count:             returned.Count(code),
			ReturnedLocations: stringifyLocations(returned.Locations(code)),
		})
	}
	report.Summary.DuplicateUnique = len(report.Duplicates)

	// Раскладываем результаты по категориям.
	for _, result := range results {
		count := len(result.ReturnedPlaces)
		switch result.Status {
		case match.StatusExact:
			report.Summary.ExactTotal += count
			report.Summary.ExactUnique++
		case match.StatusFuzzy:
			report.Summary.FuzzyTotal += count
			report.Summary.FuzzyUnique++
			report.Fuzzy = append(report.Fuzzy, problemFromMatchResult(result))
		case match.StatusUnknown:
			report.Summary.UnknownTotal += count
			report.Summary.UnknownUnique++
			report.Unknown = append(report.Unknown, problemFromMatchResult(result))
		}
	}

	return report, nil
}

// RunDuplicates — поиск дубликатов внутри одного файла.
// Читает файл, возвращает все коды с len(locations) > 1.
func RunDuplicates(path string) (DuplicateReport, error) {
	index, err := input.ReadCodes(path)
	if err != nil {
		return DuplicateReport{}, err
	}

	report := DuplicateReport{
		Summary: FileSummary{
			Total:       index.TotalCount(),
			Unique:      index.UniqueCount(),
			Diagnostics: index.Diagnostics(),
		},
		Duplicates: []Problem{},
	}

	for _, code := range index.DuplicateCodes() {
		report.Duplicates = append(report.Duplicates, Problem{
			Type:              "ДУБЛИКАТ",
			Code:              code,
			Description:       fmt.Sprintf("код встречается %d раза", index.Count(code)),
			Count:             index.Count(code),
			ReturnedLocations: stringifyLocations(index.Locations(code)),
		})
	}

	return report, nil
}

func problemFromMatchResult(result match.Result) Problem {
	problem := Problem{
		Type:              result.Status,
		Code:              result.ReturnedCode,
		MatchPercent:      result.MatchPercent,
		MatchedCode:       result.MatchedCode,
		ReturnedLocations: stringifyLocations(result.ReturnedPlaces),
		IssuedLocations:   stringifyLocations(result.MatchedPlaces),
	}
	if result.Status == match.StatusFuzzy {
		problem.Description = "код похож на выданный, нужен ручной контроль"
	} else {
		problem.Description = "не найден достаточно похожий код среди выданных"
	}
	return problem
}

func stringifyLocations(locations []codes.Location) []string {
	result := make([]string, 0, len(locations))
	for _, location := range locations {
		result = append(result, location.String())
	}
	return result
}
