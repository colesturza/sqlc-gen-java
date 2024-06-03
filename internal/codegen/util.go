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
	var lastWasUpper bool

	for i, r := range input {
		isUpper := unicode.IsUpper(r)
		if i > 0 && isUpper && (!lastWasUpper || (lastWasUpper && !unicode.IsUpper(rune(input[i-1])))) {
			words = append(words, string(currentWord))
			currentWord = []rune{r}
		} else {
			currentWord = append(currentWord, r)
		}
		lastWasUpper = isUpper
	}
	words = append(words, string(currentWord))

	for i := range words {
		words[i] = strings.ToUpper(words[i])
	}

	return strings.Join(words, "_")
}

func toLowerCase(str string) string {
	if str == "" {
		return ""
	}

	return strings.ToLower(str[:1]) + str[1:]
}
