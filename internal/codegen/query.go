package codegen

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
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

func (v QueryValue) Args() string {
	if v.isEmpty() {
		return ""
	}
	var out []string
	fields := v.Struct.Fields
	for _, f := range fields {
		out = append(out, f.Name+": "+f.Type.String())
	}
	if len(v.binding) > 0 {
		lookup := map[int]int{}
		for i, v := range v.binding {
			lookup[v] = i
		}
		sort.Slice(out, func(i, j int) bool {
			return lookup[fields[i].ID] < lookup[fields[j].ID]
		})
	}
	if len(out) < 3 {
		return strings.Join(out, ", ")
	}
	return "\n" + indent(strings.Join(out, ",\n"), 6, -1)
}

func (v QueryValue) Bindings() string {
	if v.isEmpty() {
		return ""
	}
	var out []string
	if len(v.binding) > 0 {
		for i, idx := range v.binding {
			f := v.Struct.Fields[idx-1]
			out = append(out, jdbcSet(f.Type, i+1, f.Name))
		}
	} else {
		for i, f := range v.Struct.Fields {
			out = append(out, jdbcSet(f.Type, i+1, f.Name))
		}
	}
	return indent(strings.Join(out, "\n"), 10, 0)
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
