package normalize

import "strings"

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
