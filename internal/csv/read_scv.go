package csv

import (
	"onestsignt/internal/codes"
	"onestsignt/internal/normalize"
	"os"
	"strings"
)

type ReaderCSV struct {
	path string
}

func NewReaderCSV(path string) *ReaderCSV {
	return &ReaderCSV{
		path: path,
	}
}

func (r ReaderCSV) Path() string {
	return r.path
}

func (r ReaderCSV) ReadCSV() (codes.Index, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return codes.Index{}, err
	}

	index := codes.NewIndex()
	dataText := strings.ReplaceAll(string(data), "\r\n", "\n")
	dataText = strings.ReplaceAll(dataText, "\r", "\n")
	lines := strings.Split(dataText, "\n")
	diagnostics := codes.Diagnostics{
		File: r.path,
		Rows: len(lines),
	}

	for lineIndex, line := range lines {
		line = normalize.NormalizeCode(line)
		if line == "" {
			continue
		}
		diagnostics.NonEmptyCells++
		diagnostics.CountedCells++
		index.Add(line, codes.Location{
			File: r.path,
			Line: lineIndex + 1,
		})
	}

	index.SetDiagnostics(diagnostics)
	return index, nil
}
