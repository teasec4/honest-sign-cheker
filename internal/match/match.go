// Package match — нечёткое сравнение кодов: находит похожие коды среди выданных.
// Алгоритм: n-граммы для быстрого отбора кандидатов + расстояние Левенштейна для точной оценки.
package match

import (
	"math"
	"onestsignt/internal/codes"
	"sort"
)

const (
	StatusExact   = "ТОЧНОЕ СОВПАДЕНИЕ" // код найден 1-в-1 в выданных
	StatusFuzzy   = "ПОХОЖИЙ КОД"       // код похож на выданный, нужен ручной контроль
	StatusUnknown = "НЕ НАЙДЕН"         // не удалось найти достаточно похожий код
)

type Config struct {
	MinPercent    float64 // порог схожести для StatusFuzzy (по умолчанию 85%)
	GramSize      int     // размер n-граммы (по умолчанию 3 символа)
	MaxCandidates int     // макс. число кандидатов для сравнения Левенштейном
}

func DefaultConfig() Config {
	return Config{
		MinPercent:    85,
		GramSize:      3,
		MaxCandidates: 500,
	}
}

type Result struct {
	Status          string           // Exact / Fuzzy / Unknown
	ReturnedCode    string           // код из возврата
	ReturnedPlaces  []codes.Location // где в возврате
	MatchedCode     string           // лучший похожий код из выдачи
	MatchedPlaces   []codes.Location // где в выдаче
	MatchPercent    float64          // процент схожести (0–100)
	SharedGramCount int              // сколько n-грамм совпало
}

type Matcher struct {
	issued codes.Index
	config Config
	grams  map[string][]string // n-граммный индекс: грамма → коды, где она встречается
}

// NewMatcher — строит n-граммный индекс по всем выданным кодам.
// Для каждого кода нарезает n-граммы (по 3 символа) и складывает в map[грамма] → коды.
// Это позволяет быстро найти кандидатов: вместо перебора 20000 кодов — смотрим только те,
// у которых есть общие n-граммы с искомым.
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

	// Строим n-граммный индекс: для каждого выданного кода нарезаем граммы
	// и добавляем код в списки всех его грамм.
	for _, code := range issued.Codes() {
		for _, gram := range uniqueGrams(code, matcher.config.GramSize) {
			matcher.grams[gram] = append(matcher.grams[gram], code)
		}
	}

	return matcher
}

// MatchReturned — сопоставить все возвращённые коды с выданными.
// Результат сортируется: сначала Fuzzy (нужен контроль), потом Unknown, потом Exact.
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

// MatchCode — для одного возвращённого кода определить его статус.
//  1. Точное совпадение: код есть в issued → StatusExact (100%).
//  2. Нет точного → собираем кандидатов через n-граммы.
//  3. Среди кандидатов считаем SimilarityPercent (Левенштейн).
//  4. Если лучший ≥ MinPercent → StatusFuzzy, иначе StatusUnknown.
func (m Matcher) MatchCode(returnedCode string) Result {
	// Шаг 1: точное совпадение.
	if m.issued.Has(returnedCode) {
		return Result{
			Status:        StatusExact,
			MatchedCode:   returnedCode,
			MatchedPlaces: m.issued.Locations(returnedCode),
			MatchPercent:  100,
		}
	}

	// Шаг 2: отбор кандидатов через n-граммы.
	candidates := m.candidates(returnedCode)
	if len(candidates) == 0 {
		return Result{Status: StatusUnknown}
	}

	// Шаг 3-4: сравнение Левенштейном + проверка порога.
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
	sharedGrams int // сколько n-грамм совпало с искомым кодом
}

// candidates — отобрать коды-кандидаты по n-граммам.
// Для каждой граммы искомого кода достаём все коды из grams[грамма].
// Сортируем по убыванию sharedGrams (чем больше общих грамм — тем вероятнее совпадение).
// Обрезаем до MaxCandidates (чтобы Левенштейн не считался на 10000 кодах).
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

// SimilarityPercent — процент схожести двух строк (0–100).
// Использует расстояние Левенштейна: (1 - расстояние/максДлина) * 100.
// Пример: "ABCD" vs "ABXD" → расстояние 1, максДлина 4 → (1 - 1/4)*100 = 75%.
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

// levenshtein — расстояние Левенштейна (минимальное число вставок/удалений/замен).
// Использует два ряда (previous/current) для экономии памяти — O(min(m,n)).
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
				previous[j]+1,      // удаление
				current[j-1]+1,     // вставка
				previous[j-1]+cost, // замена
			)
		}
		previous, current = current, previous
	}

	return previous[len(b)]
}

// uniqueGrams — уникальные n-граммы строки (подстроки длины size).
// Пример: "ABCD", size=3 → ["ABC", "BCD"].
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

// statusOrder — порядок сортировки результатов: Fuzzy (0) → Unknown (1) → Exact (2).
// Fuzzy первыми — чтобы пользователь видел коды, требующие ручного контроля.
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
