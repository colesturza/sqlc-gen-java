package codegen

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/sqlc-dev/plugin-sdk-go/sdk"

	"github.com/colesturza/sqlc-gen-java/internal/codegen/opts"
)

func JavaClassName(name string, options *opts.Options) string {
	out := ""
	caser := cases.Title(language.Und, cases.NoLower)
	for _, p := range strings.Split(name, "_") {
		out += caser.String(p)
	}
	return out
}

func JavaClassMemberName(name string, options *opts.Options) string {
	return sdk.LowerTitle(JavaClassName(name, options))
}

// ConvertCamelToConstant converts a camel case string to a Java static final constant identifier.
func ConvertCamelToConstant(input string) string {
	if len(input) == 0 {
		return ""
	}

	var words []string
	var currentWord []rune

	for _, r := range input {
		if unicode.IsUpper(r) && len(currentWord) > 0 {
			words = append(words, string(currentWord))
			currentWord = []rune{r}
		} else {
			currentWord = append(currentWord, r)
		}
	}
	words = append(words, string(currentWord))

	for i := range words {
		words[i] = strings.ToUpper(words[i])
	}

	return strings.Join(words, "_")
}
