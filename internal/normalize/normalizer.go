// Package normalize — очистка кодов перед добавлением в Index.
// Удаляет пробелы, невидимые спецсимволы (\x10 \x1d \x1f \x1D), заменяет кривые кавычки на '.
package normalize

import "strings"

// NormalizeCode — очистить строку кода от мусора.
// \x1d, \x1f, \x10 — Group Separator, Unit Separator, Data Link Escape (из GS1 DataMatrix).
// \x1D — то же самое, но в другом регистре.
// ? → ' — в некоторых шрифтах апостроф рендерится как ?.
func NormalizeCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.ReplaceAll(code, "\x10", "")
	code = strings.ReplaceAll(code, "\x1f", "")
	code = strings.ReplaceAll(code, "\x1d", "")
	code = strings.ReplaceAll(code, "_x001D_", "")
	code = strings.ReplaceAll(code, "?", "'")
	code = strings.ReplaceAll(code, "\u0092", "'")
	return code
}
