package str

import (
	"strings"
	"unicode"
)

func NormalizeString(input string) string {
	var builder strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) && unicode.IsLower(unicode.ToLower(r)) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToLower(r))
		}
		if builder.Len() == 30 {
			break
		}
	}
	return builder.String()
}
