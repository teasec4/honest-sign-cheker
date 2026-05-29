package compare

import (
	"fmt"
	"onestsignt/internal/codes"
	"onestsignt/internal/csv"
	"onestsignt/internal/excel"
	"onestsignt/internal/output"
	"path/filepath"
)

type problem struct {
	kind        string
	code        string
	description string
	locations   []codes.Location
}

type Comparerer struct {
	readerCSV   csv.ReaderCSV
	readerExcel excel.ReaderExcel
}

func NewComparerer(readerCSV csv.ReaderCSV, readerExcel excel.ReaderExcel) *Comparerer {
	return &Comparerer{
		readerCSV:   readerCSV,
		readerExcel: readerExcel,
	}
}

func (c Comparerer) Compare() error {
	codesCSV, err := c.readerCSV.ReadCSV()
	if err != nil {
		return err
	}

	codesExcel, err := c.readerExcel.ReadExcel()
	if err != nil {
		return err
	}

	baseName := filepath.Base(c.readerCSV.Path())
	baseName = baseName[:len(baseName)-4]

	output.WriteLine("\n=== %s ===", baseName)
	output.WriteLine("Файлы:")
	output.WriteLine("  Выдано: %s", filepath.Base(c.readerCSV.Path()))
	output.WriteLine("  Возврат: %s", filepath.Base(c.readerExcel.Path()))
	output.WriteLine("Итог:")
	output.WriteLine("  CSV: %d кодов, %d уникальных", codesCSV.TotalCount(), codesCSV.UniqueCount())
	output.WriteLine("  Excel: %d кодов, %d уникальных", codesExcel.TotalCount(), codesExcel.UniqueCount())

	problems := collectProblems(codesCSV, codesExcel)
	output.WriteLine("  Проблем: %d", len(problems))

	if len(problems) == 0 {
		output.WriteLine("OK")
	} else {
		output.WriteLine("Проблемы:")
		for index, problem := range problems {
			output.WriteLine("%d. [%s] %s", index+1, problem.kind, problem.code)
			output.WriteLine("   Что: %s", problem.description)
			output.WriteLine("   Где:")
			for _, location := range problem.locations {
				output.WriteLine("     - %s", location.String())
			}
		}
	}

	return nil
}

func collectProblems(codesCSV codes.Index, codesExcel codes.Index) []problem {
	var problems []problem

	for _, code := range codesCSV.DuplicateCodes() {
		problems = append(problems, problem{
			kind:        "ДУБЛИКАТ В ВЫДАЧЕ",
			code:        code,
			description: fmt.Sprintf("код встречается в выданном CSV %d раза", codesCSV.Count(code)),
			locations:   codesCSV.Locations(code),
		})
	}

	for _, code := range codesExcel.DuplicateCodes() {
		problems = append(problems, problem{
			kind:        "ДУБЛИКАТ В ВОЗВРАТЕ",
			code:        code,
			description: fmt.Sprintf("код встречается в размеченном файле фабрики %d раза", codesExcel.Count(code)),
			locations:   codesExcel.Locations(code),
		})
	}

	for _, code := range codesExcel.UnknownCodes(codesCSV) {
		problems = append(problems, problem{
			kind:        "КОД НЕ ВЫДАВАЛСЯ",
			code:        code,
			description: "фабрика вернула код, которого нет в выданном CSV",
			locations:   codesExcel.Locations(code),
		})
	}

	return problems
}
