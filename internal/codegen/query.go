package codegen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

type QueryValue struct {
	Emit    bool
	Name    string
	Struct  *Struct
	Typ     javaType
	binding []int

	// Column is kept so late in the generation process around to differentiate
	// between mysql slices and pg arrays
	Column *plugin.Column
}

func (v QueryValue) EmitStruct() bool {
	return v.Emit
}

func (v QueryValue) IsStruct() bool {
	return v.Struct != nil
}

func (v QueryValue) isEmpty() bool {
	return v.Typ == (javaType{}) && v.Name == "" && v.Struct == nil
}

func (v QueryValue) Type() string {
	if v.Typ != (javaType{}) {
		return v.Typ.String()
	}
	if v.Struct != nil {
		return v.Struct.Name
	}
	panic("no type for QueryValue: " + v.Name)
}

func (v QueryValue) Pair() string {
	var out []string
	for _, arg := range v.Pairs() {
		out = append(out, "final "+arg.Type+" "+arg.Name)
	}
	return strings.Join(out, ",")
}

type Argument struct {
	Name string
	Type string
}

// Return the argument name and type for query methods. Should only be used in
// the context of method arguments.
func (v QueryValue) Pairs() []Argument {
	if v.isEmpty() {
		return nil
	}
	if !v.EmitStruct() && v.IsStruct() {
		var out []Argument
		for _, f := range v.Struct.Fields {
			out = append(out, Argument{
				Name: EscapeJavaReservedWord(toLowerCase(f.Name)),
				Type: f.Type.Name,
			})
		}
		return out
	}
	return []Argument{
		{
			Name: EscapeJavaReservedWord(v.Name),
			Type: v.Type(),
		},
	}
}

func jdbcSet(t javaType, idx int, name string) string {
	if t.IsEnum && t.IsArray {
		return fmt.Sprintf(`ps.setArray(%d, conn.createArrayOf("%s", %s.toArray(new %s[0])));`, idx, t.DataType, name, t.Name)
	}
	if t.IsEnum {
		if t.Engine == "postgresql" {
			return fmt.Sprintf("ps.setObject(%d, %s, %s);", idx, name, "java.sql.Types.OTHER")
		} else {
			return fmt.Sprintf("ps.setString(%d, %s);", idx, name)
		}
	}
	if t.IsArray {
		return fmt.Sprintf(`ps.setArray(%d, conn.createArrayOf("%s", %s.toArray(new %s[0])));`, idx, t.DataType, name, t.Name)
	}
	if t.IsTime() {
		return fmt.Sprintf("ps.setObject(%d, %s);", idx, name)
	}
	if t.IsInstant() {
		return fmt.Sprintf("ps.setTimestamp(%d, Timestamp.from(%s));", idx, name)
	}
	if t.IsUUID() {
		return fmt.Sprintf("ps.setObject(%d, %s);", idx, name)
	}
	if t.Name == "Integer" {
		return fmt.Sprintf("ps.setInt(%d, %s);", idx, name)
	}
	return fmt.Sprintf("ps.set%s(%d, %s);", t.Name, idx, name)
}

func (v QueryValue) Bindings() string {
	if v.isEmpty() {
		return ""
	}
	var out []string
	if !v.IsStruct() {
		out = append(out, jdbcSet(v.Typ, 1, v.Name))
	} else {
		if len(v.binding) > 0 {
			for i, idx := range v.binding {
				f := v.Struct.Fields[idx-1]
				out = append(out, jdbcSet(f.Type, i+1, v.Name+".get"+sdk.Title(f.Name)+"()"))
			}
		} else {
			for i, f := range v.Struct.Fields {
				out = append(out, jdbcSet(f.Type, i+1, v.Name+".get"+sdk.Title(f.Name)+"()"))
			}
		}
	}
	return strings.Join(out, "\n")
}

func jdbcGet(t javaType, idx int) string {
	if t.IsEnum && t.IsArray {
		return fmt.Sprintf(`java.util.Arrays.stream((String[]) rs.getArray(%d).getArray()).map(label -> %s.valueOfLabel(label)).collect(java.util.stream.Collectors.toList())`, idx, t.Name)
	}
	if t.IsEnum {
		return fmt.Sprintf("%s.valueOfLabel(rs.getString(%d))", t.Name, idx)
	}
	if t.IsArray {
		return fmt.Sprintf(`java.util.Arrays.stream((%s[]) rs.getArray(%d).getArray()).collect(java.util.stream.Collectors.toList())`, t.Name, idx)
	}
	if t.IsTime() {
		return fmt.Sprintf(`rs.getObject(%d, %s.class)`, idx, t.Name)
	}
	if t.IsInstant() {
		return fmt.Sprintf(`rs.getTimestamp(%d).toInstant()`, idx)
	}
	if t.IsUUID() {
		return fmt.Sprintf(`rs.getObject(%d, java.util.UUID.class)`, idx)
	}
	if t.IsBigDecimal() {
		return fmt.Sprintf(`rs.getBigDecimal(%d)`, idx)
	}
	if t.Name == "Integer" {
		return fmt.Sprintf("rs.getInt(%d)", idx)
	}
	return fmt.Sprintf(`rs.get%s(%d)`, t.Name, idx)
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
	ret = indent("new "+v.Struct.Name+"(\n"+ret+")", 24, 0)
	return ret
}

// A struct used to generate methods and fields on the Queries struct
type Query struct {
	ClassName    string
	Cmd          string
	Comments     []string
	MethodName   string
	FieldName    string
	ConstantName string
	SQL          string
	SourceName   string
	Ret          QueryValue
	Arg          QueryValue
}

var postgresPlaceholderRegexp = regexp.MustCompile(`\B\$\d+\b`)

// HACK: jdbc doesn't support numbered parameters, so we need to transform them to question marks...
// But there's no access to the SQL parser here, so we just do a dumb regexp replace instead. This won't work if
// the literal strings contain matching values, but good enough for a prototype.
func jdbcSQL(s, engine string) (string, []string) {
	if engine != "postgresql" {
		return s, nil
	}
	var args []string
	q := postgresPlaceholderRegexp.ReplaceAllStringFunc(s, func(placeholder string) string {
		args = append(args, placeholder)
		return "?"
	})
	return q, args
}

func parseInts(s []string) ([]int, error) {
	if len(s) == 0 {
		return nil, nil
	}
	var refs []int
	for _, v := range s {
		i, err := strconv.Atoi(strings.TrimPrefix(v, "$"))
		if err != nil {
			return nil, err
		}
		refs = append(refs, i)
	}
	return refs, nil
}
