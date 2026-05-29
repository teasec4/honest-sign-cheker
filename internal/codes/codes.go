package codes

import (
	"fmt"
	"path/filepath"
	"sort"
)

type Location struct {
	File  string
	Sheet string
	Cell  string
	Line  int
}

func (l Location) String() string {
	fileName := filepath.Base(l.File)
	if l.Sheet != "" && l.Cell != "" {
		return fmt.Sprintf("%s:%s!%s", fileName, l.Sheet, l.Cell)
	}
	if l.Line > 0 {
		return fmt.Sprintf("%s:строка %d", fileName, l.Line)
	}
	return fileName
}

type Index struct {
	locations   map[string][]Location
	total       int
	diagnostics *Diagnostics
}

type Diagnostics struct {
	File                 string          `json:"file"`
	Sheet                string          `json:"sheet,omitempty"`
	Column               string          `json:"column,omitempty"`
	Rows                 int             `json:"rows"`
	NonEmptyCells        int             `json:"nonEmptyCells"`
	CountedCells         int             `json:"countedCells"`
	IgnoredNonEmptyCells int             `json:"ignoredNonEmptyCells"`
	IgnoredSamples       []IgnoredSample `json:"ignoredSamples"`
}

type IgnoredSample struct {
	Location string `json:"location"`
	Value    string `json:"value"`
	Reason   string `json:"reason"`
}

func NewIndex() Index {
	return Index{
		locations: make(map[string][]Location),
	}
}

func (i *Index) Add(code string, location Location) {
	i.locations[code] = append(i.locations[code], location)
	i.total++
}

func (i *Index) SetDiagnostics(diagnostics Diagnostics) {
	i.diagnostics = &diagnostics
}

func (i Index) Diagnostics() *Diagnostics {
	return i.diagnostics
}

func (i Index) TotalCount() int {
	return i.total
}

func (i Index) UniqueCount() int {
	return len(i.locations)
}

func (i Index) Count(code string) int {
	return len(i.locations[code])
}

func (i Index) Has(code string) bool {
	_, ok := i.locations[code]
	return ok
}

func (i Index) Locations(code string) []Location {
	return i.locations[code]
}

func (i Index) Codes() []string {
	result := make([]string, 0, len(i.locations))
	for code := range i.locations {
		result = append(result, code)
	}
	sort.Strings(result)
	return result
}

func (i Index) DuplicateCodes() []string {
	var duplicates []string
	for code, locations := range i.locations {
		if len(locations) > 1 {
			duplicates = append(duplicates, code)
		}
	}
	sort.Strings(duplicates)
	return duplicates
}

func (i Index) UnknownCodes(known Index) []string {
	var unknown []string
	for code := range i.locations {
		if !known.Has(code) {
			unknown = append(unknown, code)
		}
	}
	sort.Strings(unknown)
	return unknown
}
