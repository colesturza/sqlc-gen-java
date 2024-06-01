package codegen

import (
	"bytes"
	"fmt"
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

func jdbcSet(t javaType, idx int, name string) string {
	if t.IsEnum && t.IsArray {
		return fmt.Sprintf(`stmt.setArray(%d, conn.createArrayOf("%s", %s.map { v -> v.value }.toTypedArray()))`, idx, t.DataType, name)
	}
	if t.IsEnum {
		if t.Engine == "postgresql" {
			return fmt.Sprintf("stmt.setObject(%d, %s.value, %s)", idx, name, "Types.OTHER")
		} else {
			return fmt.Sprintf("stmt.setString(%d, %s.value)", idx, name)
		}
	}
	if t.IsArray {
		return fmt.Sprintf(`stmt.setArray(%d, conn.createArrayOf("%s", %s.toTypedArray()))`, idx, t.DataType, name)
	}
	if t.IsTime() {
		return fmt.Sprintf("stmt.setObject(%d, %s)", idx, name)
	}
	if t.IsInstant() {
		return fmt.Sprintf("stmt.setTimestamp(%d, Timestamp.from(%s))", idx, name)
	}
	if t.IsUUID() {
		return fmt.Sprintf("stmt.setObject(%d, %s)", idx, name)
	}
	return fmt.Sprintf("stmt.set%s(%d, %s)", t.Name, idx, name)
}

func jdbcGet(t javaType, idx int) string {
	if t.IsEnum && t.IsArray {
		return fmt.Sprintf(`(results.getArray(%d).array as Array<String>).map { v -> %s.lookup(v)!! }.toList()`, idx, t.Name)
	}
	if t.IsEnum {
		return fmt.Sprintf("%s.lookup(results.getString(%d))!!", t.Name, idx)
	}
	if t.IsArray {
		return fmt.Sprintf(`(results.getArray(%d).array as Array<%s>).toList()`, idx, t.Name)
	}
	if t.IsTime() {
		return fmt.Sprintf(`results.getObject(%d, %s::class.java)`, idx, t.Name)
	}
	if t.IsInstant() {
		return fmt.Sprintf(`results.getTimestamp(%d).toInstant()`, idx)
	}
	if t.IsUUID() {
		var nullCast string
		if t.IsNull {
			nullCast = "?"
		}
		return fmt.Sprintf(`results.getObject(%d) as%s %s`, idx, nullCast, t.Name)
	}
	if t.IsBigDecimal() {
		return fmt.Sprintf(`results.getBigDecimal(%d)`, idx)
	}
	return fmt.Sprintf(`results.get%s(%d)`, t.Name, idx)
}

func (v QueryValue) ResultSet() string {
	var out []string
	if v.Struct == nil {
		return jdbcGet(v.Typ, 1)
	}
	for i, f := range v.Struct.Fields {
		out = append(out, jdbcGet(f.Type, i+1))
	}
	ret := indent(strings.Join(out, ",\n"), 4, -1)
	ret = indent(v.Struct.Name+"(\n"+ret+"\n)", 12, 0)
	return ret
}

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
