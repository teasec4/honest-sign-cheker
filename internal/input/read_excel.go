package input

import (
	"onestsignt/internal/codes"
	"onestsignt/internal/normalize"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ReaderExcel struct {
	path string
}

func NewReaderExcel(path string) *ReaderExcel {
	return &ReaderExcel{
		path: path,
	}
}

func (r ReaderExcel) Path() string {
	return r.path
}

// ReadExcel — читает первый лист xlsx и возвращает Index с кодами.
// Два режима:
//  1. Быстрый: если есть колонка "DM CODE" — читает только её (пропуская заголовок).
//  2. Медленный: иначе сканирует ВСЕ ячейки и фильтрует через isCodeLike().
func (r ReaderExcel) ReadExcel() (codes.Index, error) {
	file, err := excelize.OpenFile(r.path)
	if err != nil {
		return codes.Index{}, err
	}
	defer file.Close()

	sheets := file.GetSheetList()
	sheetName := sheets[0]

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return codes.Index{}, err
	}

	index := codes.NewIndex()
	diagnostics := codes.Diagnostics{
		File:  r.path,
		Sheet: sheetName,
		Rows:  len(rows),
	}

	// Быстрый путь: колонка "DM CODE" найдена в заголовке.
	if columnIndex, ok := findHeaderColumn(rows, "DM CODE"); ok {
		if err := r.readColumn(rows, sheetName, columnIndex, &index, &diagnostics); err != nil {
			return codes.Index{}, err
		}
		index.SetDiagnostics(diagnostics)
		return index, nil
	}

	// Медленный путь: полный скан всех ячеек.
	for rowIndex, row := range rows {
		for colIndex, cell := range row {
			code := normalize.NormalizeCode(cell)
			if code == "" {
				continue
			}
			diagnostics.NonEmptyCells++

			cellName, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				return codes.Index{}, err
			}

			location := codes.Location{
				File:  r.path,
				Sheet: sheetName,
				Cell:  cellName,
			}

			if isCodeLike(code) {
				diagnostics.CountedCells++
				index.Add(code, codes.Location{
					File:  r.path,
					Sheet: sheetName,
					Cell:  cellName,
				})
			} else {
				diagnostics.IgnoredNonEmptyCells++
				addIgnoredSample(&diagnostics, location, code, "ячейка не похожа на код маркировки")
			}
		}
	}

	index.SetDiagnostics(diagnostics)
	return index, nil
}

// readColumn — читает только одну колонку (быстрый путь).
// Пропускает строку 0 (заголовок), начинает со строки 1.
func (r ReaderExcel) readColumn(rows [][]string, sheetName string, columnIndex int, index *codes.Index, diagnostics *codes.Diagnostics) error {
	columnName, err := excelize.ColumnNumberToName(columnIndex + 1)
	if err != nil {
		return err
	}
	diagnostics.Column = columnName

	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		if columnIndex >= len(row) {
			continue
		}

		code := normalize.NormalizeCode(row[columnIndex])
		if code == "" {
			continue
		}

		diagnostics.NonEmptyCells++
		diagnostics.CountedCells++

		cellName, err := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+1)
		if err != nil {
			return err
		}
		index.Add(code, codes.Location{
			File:  r.path,
			Sheet: sheetName,
			Cell:  cellName,
		})
	}

	return nil
}

// findHeaderColumn — ищет колонку по заголовку в первой строке.
// Нормализует имена: trim, upper, схлопывает пробелы.
func findHeaderColumn(rows [][]string, header string) (int, bool) {
	if len(rows) == 0 {
		return 0, false
	}
	header = normalizeHeader(header)
	for index, cell := range rows[0] {
		if normalizeHeader(cell) == header {
			return index, true
		}
	}
	return 0, false
}

func normalizeHeader(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToUpper(value)
	value = strings.Join(strings.Fields(value), " ")
	return value
}

// isCodeLike — эвристика: ячейка похожа на код маркировки?
// Коды DataMatrix всегда начинаются с "01" (GTIN) и содержат разделитель "93".
func isCodeLike(code string) bool {
	if strings.HasPrefix(code, "01") {
		return true
	}
	return len([]rune(code)) >= 20 && strings.Contains(code, "93")
}

func addIgnoredSample(diagnostics *codes.Diagnostics, location codes.Location, value string, reason string) {
	if len(diagnostics.IgnoredSamples) >= 30 {
		return
	}
	diagnostics.IgnoredSamples = append(diagnostics.IgnoredSamples, codes.IgnoredSample{
		Location: location.String(),
		Value:    truncate(value, 80),
		Reason:   reason,
	})
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}
