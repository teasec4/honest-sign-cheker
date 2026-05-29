package excel

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

	if columnIndex, ok := findHeaderColumn(rows, "DM CODE"); ok {
		if err := r.readColumn(rows, sheetName, columnIndex, &index, &diagnostics); err != nil {
			return codes.Index{}, err
		}
		index.SetDiagnostics(diagnostics)
		return index, nil
	}

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
