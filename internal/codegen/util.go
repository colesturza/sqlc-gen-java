package codegen

import (
	"strings"

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
