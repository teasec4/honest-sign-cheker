package match

import (
	"onestsignt/internal/codes"
	"testing"
)

func TestMatchCodeExact(t *testing.T) {
	issued := codes.NewIndex()
	issued.Add("010123456789012121ABC93XYZ", codes.Location{File: "issued.csv", Line: 1})

	matcher := NewMatcher(issued, DefaultConfig())
	result := matcher.MatchCode("010123456789012121ABC93XYZ")

	if result.Status != StatusExact {
		t.Fatalf("status = %s, want %s", result.Status, StatusExact)
	}
	if result.MatchPercent != 100 {
		t.Fatalf("percent = %.2f, want 100", result.MatchPercent)
	}
}

func TestMatchCodeFuzzy(t *testing.T) {
	issued := codes.NewIndex()
	issued.Add("010123456789012121ABC93XYZ", codes.Location{File: "issued.csv", Line: 1})

	config := DefaultConfig()
	config.MinPercent = 80
	matcher := NewMatcher(issued, config)
	result := matcher.MatchCode("010123456789012121ABD93XYZ")

	if result.Status != StatusFuzzy {
		t.Fatalf("status = %s, want %s", result.Status, StatusFuzzy)
	}
	if result.MatchedCode != "010123456789012121ABC93XYZ" {
		t.Fatalf("matched = %s, want issued code", result.MatchedCode)
	}
}

func TestMatchCodeUnknownBelowThreshold(t *testing.T) {
	issued := codes.NewIndex()
	issued.Add("010123456789012121ABC93XYZ", codes.Location{File: "issued.csv", Line: 1})

	config := DefaultConfig()
	config.MinPercent = 99
	matcher := NewMatcher(issued, config)
	result := matcher.MatchCode("010123456789012121ABD93XYZ")

	if result.Status != StatusUnknown {
		t.Fatalf("status = %s, want %s", result.Status, StatusUnknown)
	}
}
