package csv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCSVCountsDuplicates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "codes.csv")
	data := "0101\n0102\n0101\n\n0103\x1dABC\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	reader := NewReaderCSV(path)
	codes, err := reader.ReadCSV()
	if err != nil {
		t.Fatal(err)
	}

	if got := codes.Count("0101"); got != 2 {
		t.Fatalf("0101 count = %d, want 2", got)
	}
	if got := codes.Count("0102"); got != 1 {
		t.Fatalf("0102 count = %d, want 1", got)
	}
	if got := codes.Count("0103ABC"); got != 1 {
		t.Fatalf("0103ABC count = %d, want 1", got)
	}

	locations := codes.Locations("0101")
	if got := len(locations); got != 2 {
		t.Fatalf("0101 locations = %d, want 2", got)
	}
	if locations[0].Line != 1 || locations[1].Line != 3 {
		t.Fatalf("0101 lines = %d,%d, want 1,3", locations[0].Line, locations[1].Line)
	}
}
