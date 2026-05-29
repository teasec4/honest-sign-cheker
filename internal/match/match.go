package match

import (
	"math"
	"onestsignt/internal/codes"
	"sort"
)

const (
	StatusExact   = "ТОЧНОЕ СОВПАДЕНИЕ"
	StatusFuzzy   = "ПОХОЖИЙ КОД"
	StatusUnknown = "НЕ НАЙДЕН"
)

type Config struct {
	MinPercent    float64
	GramSize      int
	MaxCandidates int
}

func DefaultConfig() Config {
	return Config{
		MinPercent:    85,
		GramSize:      3,
		MaxCandidates: 500,
	}
}

type Result struct {
	Status          string
	ReturnedCode    string
	ReturnedPlaces  []codes.Location
	MatchedCode     string
	MatchedPlaces   []codes.Location
	MatchPercent    float64
	SharedGramCount int
}

type Matcher struct {
	issued codes.Index
	config Config
	grams  map[string][]string
}

func NewMatcher(issued codes.Index, config Config) Matcher {
	if config.MinPercent == 0 {
		config.MinPercent = DefaultConfig().MinPercent
	}
	if config.GramSize == 0 {
		config.GramSize = DefaultConfig().GramSize
	}
	if config.MaxCandidates == 0 {
		config.MaxCandidates = DefaultConfig().MaxCandidates
	}

	matcher := Matcher{
		issued: issued,
		config: config,
		grams:  make(map[string][]string),
	}

	for _, code := range issued.Codes() {
		for _, gram := range uniqueGrams(code, matcher.config.GramSize) {
			matcher.grams[gram] = append(matcher.grams[gram], code)
		}
	}

	return matcher
}

func (m Matcher) MatchReturned(returned codes.Index) []Result {
	results := make([]Result, 0, returned.UniqueCount())
	for _, returnedCode := range returned.Codes() {
		result := m.MatchCode(returnedCode)
		result.ReturnedCode = returnedCode
		result.ReturnedPlaces = returned.Locations(returnedCode)
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Status != results[j].Status {
			return statusOrder(results[i].Status) < statusOrder(results[j].Status)
		}
		if results[i].MatchPercent != results[j].MatchPercent {
			return results[i].MatchPercent > results[j].MatchPercent
		}
		return results[i].ReturnedCode < results[j].ReturnedCode
	})
	return results
}

func (m Matcher) MatchCode(returnedCode string) Result {
	if m.issued.Has(returnedCode) {
		return Result{
			Status:        StatusExact,
			MatchedCode:   returnedCode,
			MatchedPlaces: m.issued.Locations(returnedCode),
			MatchPercent:  100,
		}
	}

	candidates := m.candidates(returnedCode)
	if len(candidates) == 0 {
		return Result{Status: StatusUnknown}
	}

	var best Result
	for _, candidate := range candidates {
		percent := SimilarityPercent(returnedCode, candidate.code)
		if percent > best.MatchPercent {
			best = Result{
				Status:          StatusUnknown,
				MatchedCode:     candidate.code,
				MatchedPlaces:   m.issued.Locations(candidate.code),
				MatchPercent:    percent,
				SharedGramCount: candidate.sharedGrams,
			}
		}
	}

	if best.MatchPercent >= m.config.MinPercent {
		best.Status = StatusFuzzy
	}

	return best
}

type candidate struct {
	code        string
	sharedGrams int
}

func (m Matcher) candidates(returnedCode string) []candidate {
	counts := make(map[string]int)
	for _, gram := range uniqueGrams(returnedCode, m.config.GramSize) {
		for _, code := range m.grams[gram] {
			counts[code]++
		}
	}

	candidates := make([]candidate, 0, len(counts))
	for code, count := range counts {
		candidates = append(candidates, candidate{code: code, sharedGrams: count})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].sharedGrams != candidates[j].sharedGrams {
			return candidates[i].sharedGrams > candidates[j].sharedGrams
		}
		return candidates[i].code < candidates[j].code
	})

	if len(candidates) > m.config.MaxCandidates {
		candidates = candidates[:m.config.MaxCandidates]
	}

	return candidates
}

func SimilarityPercent(a, b string) float64 {
	ar := []rune(a)
	br := []rune(b)
	maxLen := math.Max(float64(len(ar)), float64(len(br)))
	if maxLen == 0 {
		return 100
	}
	distance := levenshtein(ar, br)
	return (1 - float64(distance)/maxLen) * 100
}

func levenshtein(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}

	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			current[j] = minInt(
				previous[j]+1,
				current[j-1]+1,
				previous[j-1]+cost,
			)
		}
		previous, current = current, previous
	}

	return previous[len(b)]
}

func uniqueGrams(value string, size int) []string {
	runes := []rune(value)
	if len(runes) < size || size <= 0 {
		return nil
	}

	seen := make(map[string]bool)
	grams := make([]string, 0, len(runes)-size+1)
	for i := 0; i <= len(runes)-size; i++ {
		gram := string(runes[i : i+size])
		if seen[gram] {
			continue
		}
		seen[gram] = true
		grams = append(grams, gram)
	}
	return grams
}

func statusOrder(status string) int {
	switch status {
	case StatusFuzzy:
		return 0
	case StatusUnknown:
		return 1
	case StatusExact:
		return 2
	default:
		return 3
	}
}

func minInt(values ...int) int {
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}
