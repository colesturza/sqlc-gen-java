package core

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func dataClassName(name string, settings *plugin.Settings) string {
	out := ""
	caser := cases.Title(language.Und, cases.NoLower)
	for _, p := range strings.Split(name, "_") {
		out += caser.String(p)
	}
	return out
}
