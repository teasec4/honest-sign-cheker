package excel

import (
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestReadExcelCountsDuplicates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "codes.xlsx")
	file := excelize.NewFile()
	if err := file.SetCellValue("Sheet1", "A1", "0101"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "A2", "0102"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "A3", "0101"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "A4", "not-a-code"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "A5", "BROKEN010123456789012121ABC93XYZ"); err != nil {
		t.Fatal(err)
	}
	if err := file.SaveAs(path); err != nil {
		t.Fatal(err)
	}

	reader := NewReaderExcel(path)
	codes, err := reader.ReadExcel()
	if err != nil {
		t.Fatal(err)
	}

	if got := codes.Count("0101"); got != 2 {
		t.Fatalf("0101 count = %d, want 2", got)
	}
	if got := codes.Count("0102"); got != 1 {
		t.Fatalf("0102 count = %d, want 1", got)
	}
	if codes.Has("not-a-code") {
		t.Fatal("non-code cell was counted")
	}
	if got := codes.Count("BROKEN010123456789012121ABC93XYZ"); got != 1 {
		t.Fatalf("broken code-like cell count = %d, want 1", got)
	}

	locations := codes.Locations("0101")
	if got := len(locations); got != 2 {
		t.Fatalf("0101 locations = %d, want 2", got)
	}
	if locations[0].Cell != "A1" || locations[1].Cell != "A3" {
		t.Fatalf("0101 cells = %s,%s, want A1,A3", locations[0].Cell, locations[1].Cell)
	}

	diagnostics := codes.Diagnostics()
	if diagnostics == nil {
		t.Fatal("diagnostics is nil")
	}
	if diagnostics.CountedCells != 4 {
		t.Fatalf("counted cells = %d, want 4", diagnostics.CountedCells)
	}
	if diagnostics.IgnoredNonEmptyCells != 1 {
		t.Fatalf("ignored cells = %d, want 1", diagnostics.IgnoredNonEmptyCells)
	}
}

func TestReadExcelUsesDMCodeColumnWhenPresent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "returned.xlsx")
	file := excelize.NewFile()
	if err := file.SetCellValue("Sheet1", "A1", "CARTON.NO"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "B1", "DM CODE"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "C1", "Product Name"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "B2", "0101"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "B3", ">broken-code-without-prefix"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "C3", "TO MY SKIN 0101 PRODUCT NAME"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellValue("Sheet1", "B4", "0101"); err != nil {
		t.Fatal(err)
	}
	if err := file.SaveAs(path); err != nil {
		t.Fatal(err)
	}

	reader := NewReaderExcel(path)
	codes, err := reader.ReadExcel()
	if err != nil {
		t.Fatal(err)
	}

	if got := codes.TotalCount(); got != 3 {
		t.Fatalf("total count = %d, want 3", got)
	}
	if got := codes.Count("0101"); got != 2 {
		t.Fatalf("0101 count = %d, want 2", got)
	}
	if got := codes.Count(">broken-code-without-prefix"); got != 1 {
		t.Fatalf("broken code count = %d, want 1", got)
	}
	if codes.Has("TO MY SKIN 0101 PRODUCT NAME") {
		t.Fatal("product name column was counted")
	}

	diagnostics := codes.Diagnostics()
	if diagnostics == nil {
		t.Fatal("diagnostics is nil")
	}
	if diagnostics.Column != "B" {
		t.Fatalf("diagnostics column = %q, want B", diagnostics.Column)
	}
}
