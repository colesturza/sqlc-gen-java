package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/colesturza/sqlc-gen-java/internal/codegen/opts"
	"github.com/colesturza/sqlc-gen-java/internal/inflection"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func BuildJavaEnums(req *plugin.GenerateRequest, options *opts.Options) []Enum {
	var enums []Enum
	for _, schema := range req.Catalog.Schemas {
		if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
			continue
		}
		for _, enum := range schema.Enums {
			var enumName string
			if schema.Name == req.Catalog.DefaultSchema {
				enumName = enum.Name
			} else {
				enumName = schema.Name + "_" + enum.Name
			}
			e := Enum{
				Name:    JavaClassName(enumName, options),
				Comment: enum.Comment,
			}
			for _, v := range enum.Vals {
				e.Constants = append(e.Constants, Constant{
					Name:  JavaEnumConstantName(v),
					Value: v,
					Type:  e.Name,
				})
			}
			enums = append(enums, e)
		}
	}
	if len(enums) > 0 {
		sort.Slice(enums, func(i, j int) bool { return enums[i].Name < enums[j].Name })
	}
	return enums
}

func BuildJavaClasses(req *plugin.GenerateRequest, options *opts.Options) []Struct {
	var structs []Struct
	for _, schema := range req.Catalog.Schemas {
		if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
			continue
		}
		for _, table := range schema.Tables {
			var tableName string
			if schema.Name == req.Catalog.DefaultSchema {
				tableName = table.Rel.Name
			} else {
				tableName = schema.Name + "_" + table.Rel.Name
			}
			structName := tableName
			if !options.EmitExactTableNames {
				structName = inflection.Singular(inflection.SingularParams{
					Name:       structName,
					Exclusions: options.InflectionExcludeTableNames,
				})
			}
			s := Struct{
				Table:   &plugin.Identifier{Schema: schema.Name, Name: table.Rel.Name},
				Name:    JavaClassName(structName, options),
				Comment: table.Comment,
			}
			for _, column := range table.Columns {
				s.Fields = append(s.Fields, Field{
					Name:    JavaClassMemberName(column.Name, options),
					Type:    makeType(req, options, column),
					Comment: column.Comment,
				})
			}
			structs = append(structs, s)
		}
	}
	if len(structs) > 0 {
		sort.Slice(structs, func(i, j int) bool { return structs[i].Name < structs[j].Name })
	}
	return structs
}

type goColumn struct {
	id int
	*plugin.Column
}

func columnName(c *plugin.Column, pos int) string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("column_%d", pos+1)
}

func paramName(p *plugin.Parameter) string {
	if p.Column.Name != "" {
		return argName(p.Column.Name)
	}
	return fmt.Sprintf("dollar_%d", p.Number)
}

func argName(name string) string {
	out := ""
	for i, p := range strings.Split(name, "_") {
		if i == 0 {
			out += strings.ToLower(p)
		} else {
			out += strings.Title(p)
		}
	}
	return out
}

func BuildQueries(req *plugin.GenerateRequest, options *opts.Options, structs []Struct) ([]Query, error) {
	qs := make([]Query, 0, len(req.Queries))
	for _, query := range req.Queries {
		if query.Name == "" {
			continue
		}
		if query.Cmd == "" {
			continue
		}

		constantName := ConvertCamelToConstant(query.Name)

		queryText, args := jdbcSQL(query.Text, req.Settings.Engine)
		refs, err := parseInts(args)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter reference: %w", err)
		}

		gq := Query{
			Cmd:          query.Cmd,
			ClassName:    strings.Title(query.Name),
			ConstantName: constantName,
			FieldName:    sdk.LowerTitle(query.Name) + "Stmt",
			MethodName:   sdk.LowerTitle(query.Name),
			SourceName:   query.Filename,
			SQL:          queryText,
			Comments:     query.Comments,
		}

		qpl := int(*options.QueryParameterLimit)

		if len(query.Params) == 1 && qpl != 0 {
			p := query.Params[0]
			gq.Arg = QueryValue{
				Name:    EscapeJavaReservedWord(paramName(p)),
				Typ:     makeType(req, options, p.Column),
				Column:  p.Column,
				binding: refs,
			}
		} else if len(query.Params) >= 1 {
			var cols []goColumn
			for _, p := range query.Params {
				cols = append(cols, goColumn{
					id:     int(p.Number),
					Column: p.Column,
				})
			}
			s := columnsToStruct(req, options, gq.ClassName+"Params", cols)
			gq.Arg = QueryValue{
				Emit:    true,
				Name:    "params",
				Struct:  s,
				binding: refs,
			}

			if len(query.Params) <= qpl {
				gq.Arg.Emit = false
			}
		}

		if len(query.Columns) == 1 && query.Columns[0].EmbedTable == nil {
			c := query.Columns[0]
			name := columnName(c, 0)
			name = strings.Replace(name, "$", "_", -1)
			gq.Ret = QueryValue{
				Name: EscapeJavaReservedWord(name),
				Typ:  makeType(req, options, c),
			}
		} else if putOutColumns(query) {
			var gs *Struct
			var emit bool

			for _, s := range structs {
				if len(s.Fields) != len(query.Columns) {
					continue
				}
				same := true
				for i, f := range s.Fields {
					c := query.Columns[i]
					sameName := f.Name == JavaClassMemberName(columnName(c, i), options)
					sameType := f.Type == makeType(req, options, c)
					sameTable := sdk.SameTableName(c.Table, s.Table, req.Catalog.DefaultSchema)
					if !sameName || !sameType || !sameTable {
						same = false
					}
				}
				if same {
					gs = &s
					break
				}
			}

			if gs == nil {
				var columns []goColumn
				for i, c := range query.Columns {
					columns = append(columns, goColumn{
						id:     i,
						Column: c,
					})
				}
				gs = columnsToStruct(req, options, gq.ClassName+"Row", columns)
				emit = true
			}
			gq.Ret = QueryValue{
				Emit:   emit,
				Name:   "result",
				Struct: gs,
			}
		}

		qs = append(qs, gq)
	}
	sort.Slice(qs, func(i, j int) bool { return qs[i].MethodName < qs[j].MethodName })
	return qs, nil
}

var cmdReturnsData = map[string]struct{}{
	metadata.CmdBatchMany: {},
	metadata.CmdBatchOne:  {},
	metadata.CmdMany:      {},
	metadata.CmdOne:       {},
}

func putOutColumns(query *plugin.Query) bool {
	_, found := cmdReturnsData[query.Cmd]
	return found
}

func columnsToStruct(req *plugin.GenerateRequest, options *opts.Options, name string, columns []goColumn) *Struct {
	gs := Struct{
		Name: name,
	}
	idSeen := map[int]Field{}
	nameSeen := map[string]int{}
	for _, c := range columns {
		if _, ok := idSeen[c.id]; ok {
			continue
		}
		fieldName := JavaClassMemberName(columnName(c.Column, c.id), options)
		if v := nameSeen[c.Name]; v > 0 {
			fieldName = fmt.Sprintf("%s_%d", fieldName, v+1)
		}
		field := Field{
			ID:   c.id,
			Name: fieldName,
			Type: makeType(req, options, c.Column),
		}
		gs.Fields = append(gs.Fields, field)
		nameSeen[c.Name]++
		idSeen[c.id] = field
	}
	return &gs
}
