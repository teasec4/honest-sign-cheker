package compare

import (
	"onestsignt/internal/codes"
	"reflect"
	"testing"
)

func TestCollectProblems(t *testing.T) {
	codesCSV := codes.NewIndex()
	codesCSV.Add("0101", codes.Location{File: "issued.csv", Line: 1})
	codesCSV.Add("0101", codes.Location{File: "issued.csv", Line: 3})
	codesCSV.Add("0102", codes.Location{File: "issued.csv", Line: 2})

	codesExcel := codes.NewIndex()
	codesExcel.Add("0102", codes.Location{File: "returned.xlsx", Sheet: "Sheet1", Cell: "A1"})
	codesExcel.Add("0103", codes.Location{File: "returned.xlsx", Sheet: "Sheet1", Cell: "A2"})
	codesExcel.Add("0103", codes.Location{File: "returned.xlsx", Sheet: "Sheet1", Cell: "A3"})

	got := collectProblems(codesCSV, codesExcel)
	if len(got) != 3 {
		t.Fatalf("collectProblems() returned %d problems, want 3", len(got))
	}

	wantKinds := []string{
		"ДУБЛИКАТ В ВЫДАЧЕ",
		"ДУБЛИКАТ В ВОЗВРАТЕ",
		"КОД НЕ ВЫДАВАЛСЯ",
	}
	gotKinds := []string{got[0].kind, got[1].kind, got[2].kind}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Fatalf("problem kinds = %v, want %v", gotKinds, wantKinds)
	}

	if got[0].locations[0].Line != 1 || got[0].locations[1].Line != 3 {
		t.Fatalf("csv duplicate locations = %v, want lines 1 and 3", got[0].locations)
	}
	if got[1].locations[0].Cell != "A2" || got[1].locations[1].Cell != "A3" {
		t.Fatalf("excel duplicate locations = %v, want cells A2 and A3", got[1].locations)
	}
}
