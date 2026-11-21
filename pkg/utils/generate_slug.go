package utils

import (
	"strings"
	"unicode"
)

func GenerateSlug(value string) string {
	var slug strings.Builder
	separator := false

	for _, char := range strings.ToLower(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			if separator && slug.Len() > 0 {
				slug.WriteByte('-')
			}
			slug.WriteRune(char)
			separator = false
		} else {
			separator = true
		}
	}

	return slug.String()
}
