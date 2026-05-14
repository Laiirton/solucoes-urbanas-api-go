package models

import (
	"strings"
	"unicode"
)

func Slugify(text string) string {
	replacements := map[rune]string{
		'á': "a", 'à': "a", 'â': "a", 'ã': "a", 'ä': "a",
		'é': "e", 'è': "e", 'ê': "e", 'ë': "e",
		'í': "i", 'ì': "i", 'î': "i", 'ï': "i",
		'ó': "o", 'ò': "o", 'ô': "o", 'õ': "o", 'ö': "o",
		'ú': "u", 'ù': "u", 'û': "u", 'ü': "u",
		'ç': "c", 'ñ': "n",
		'Á': "a", 'À': "a", 'Â': "a", 'Ã': "a", 'Ä': "a",
		'É': "e", 'È': "e", 'Ê': "e", 'Ë': "e",
		'Í': "i", 'Ì': "i", 'Î': "i", 'Ï': "i",
		'Ó': "o", 'Ò': "o", 'Ô': "o", 'Õ': "o", 'Ö': "o",
		'Ú': "u", 'Ù': "u", 'Û': "u", 'Ü': "u",
		'Ç': "c", 'Ñ': "n",
	}

	var result strings.Builder
	for _, r := range text {
		if replacement, ok := replacements[r]; ok {
			result.WriteString(replacement)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			result.WriteRune('-')
		}
	}

	slug := strings.ToLower(result.String())
	slug = strings.Trim(slug, "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	return slug
}
