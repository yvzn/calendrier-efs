package normalize

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
var trimDash = regexp.MustCompile(`(^-+|-+$)`)

// Name strips accents, lowercases, and replaces any run of non
// alphanumeric characters with a single hyphen. E.g. "Notre-Dame-des-Landes"
// -> "notre-dame-des-landes", "Le Gâvre" -> "le-gavre".
func Name(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		out = s
	}
	out = strings.ToLower(out)
	out = nonAlnum.ReplaceAllString(out, "-")
	out = trimDash.ReplaceAllString(out, "")
	return out
}
