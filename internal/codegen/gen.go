package codegen

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

type JavaTmplCtx struct {
	Q           string
	Package     string
	Enums       []Enum
	DataClasses []Struct
	Queries     []Query
	SqlcVersion string

	// TODO: Race conditions
	SourceName string

	EmitJSONTags        bool
	EmitPreparedQueries bool
	EmitInterface       bool
}

var javaIdentPattern = regexp.MustCompile("[^a-zA-Z0-9_]+")

func indent(s string, n int, firstIndent int) string {
	lines := strings.Split(s, "\n")
	buf := bytes.NewBuffer(nil)
	for i, l := range lines {
		indent := n
		if i == 0 && firstIndent != -1 {
			indent = firstIndent
		}
		if i != 0 {
			buf.WriteRune('\n')
		}
		for i := 0; i < indent; i++ {
			buf.WriteRune(' ')
		}
		buf.WriteString(l)
	}
	return buf.String()
}

func JavaFormat(s string) string {
	skipNextSpace := false
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		isSpace := len(strings.TrimSpace(l)) == 0
		if !isSpace || !skipNextSpace {
			lines = append(lines, strings.TrimRightFunc(l, unicode.IsSpace))
		}
		skipNextSpace = isSpace
	}
	o := strings.Join(lines, "\n")
	o += "\n"
	return o
}
