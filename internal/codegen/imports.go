package codegen

import (
	"sort"

	"github.com/colesturza/sqlc-gen-java/internal/codegen/opts"
)

type Importer struct {
	options *opts.Options
	enums   []Enum
	classes []Struct
	queries []Query
	cache   map[string][]string
}

func (i *Importer) Imports(name string) []string {
	return i.cache[name]
}

func NewEnumImporter(options *opts.Options, enums []Enum, classes []Struct, queries []Query) Importer {
	i := Importer{
		options: options,
		enums:   enums,
		classes: classes,
		queries: queries,
		cache:   make(map[string][]string),
	}
	for _, e := range i.enums {
		i.cache[e.Name+".java"] = enumImports(i.options, e)
	}
	for _, e := range i.classes {
		i.cache[e.Name+".java"] = classImports(i.options, e)
	}
	i.cache["Queries.java"] = queriesIfaceImports(options, queries)
	i.cache["QueriesImpl.java"] = queriesImports(options, queries)
	return i
}

func enumImports(options *opts.Options, enum Enum) []string {
	importSet := make(map[string]struct{})

	importSet["java.util.HashMap"] = struct{}{}
	importSet["java.util.Map"] = struct{}{}

	importSlice := make([]string, 0, len(importSet))
	for s := range importSet {
		importSlice = append(importSlice, s)
	}

	sort.Strings(importSlice)
	return importSlice
}

func classImports(options *opts.Options, class Struct) []string {
	uses := func(name string) bool {
		for _, f := range class.Fields {
			if f.Type.Name == name {
				return true
			}
		}
		return false
	}

	importSet := make(map[string]struct{})

	std := stdImports(uses)
	for i, _ := range std {
		importSet[i] = struct{}{}
	}

	importSlice := make([]string, 0, len(importSet))
	for s := range importSet {
		importSlice = append(importSlice, s)
	}

	sort.Strings(importSlice)
	return importSlice
}

func queriesIfaceImports(options *opts.Options, queries []Query) []string {
	uses := func(name string) bool {
		for _, q := range queries {
			if !q.Arg.isEmpty() {
				if q.Arg.IsStruct() && !q.Arg.Emit {
					for _, f := range q.Arg.Struct.Fields {
						if f.Type.Name == name {
							return true
						}
					}
				} else if q.Arg.Type() == name {
					return true
				}
			}
		}
		return false
	}

	importSet := make(map[string]struct{})

	for _, query := range queries {
		if query.Cmd == ":many" {
			importSet["java.util.List"] = struct{}{}
		}
		if query.Cmd == ":execresult" {
			importSet["java.util.Map"] = struct{}{}
		}
	}

	std := stdImports(uses)
	for i := range std {
		importSet[i] = struct{}{}
	}

	importSlice := make([]string, 0, len(importSet))
	for s := range importSet {
		importSlice = append(importSlice, s)
	}

	sort.Strings(importSlice)
	return importSlice
}

func queriesImports(options *opts.Options, queries []Query) []string {
	uses := func(name string) bool {
		for _, q := range queries {
			if !q.Ret.isEmpty() {
				if q.Ret.Struct != nil {
					for _, f := range q.Ret.Struct.Fields {
						if f.Type.Name == name {
							return true
						}
					}
				}
				if q.Ret.Type() == name {
					return true
				}
			}
			if !q.Arg.isEmpty() {
				if q.Arg.IsStruct() {
					for _, f := range q.Arg.Struct.Fields {
						if f.Type.Name == name {
							return true
						}
					}
				} else if q.Arg.Type() == name {
					return true
				}
			}
		}
		return false
	}

	hasEnum := func() int {
		res := -1
		for _, q := range queries {
			if !q.Arg.isEmpty() {
				if q.Arg.IsStruct() {
					for _, f := range q.Arg.Struct.Fields {
						if f.Type.IsEnum {
							if f.Type.IsArray {
								res = 2
							} else {
								res = 1
							}
						}
					}
				} else if q.Arg.Typ.IsEnum {
					if q.Arg.Typ.IsArray {
						res = 2
					} else {
						res = 1
					}
				}
			}
			if !q.Ret.isEmpty() {
				if q.Ret.IsStruct() {
					for _, f := range q.Ret.Struct.Fields {
						if f.Type.IsEnum {
							if f.Type.IsArray {
								res = 2
							} else {
								res = 1
							}
						}
					}
				}
			}
		}
		return res
	}

	importSet := make(map[string]struct{})

	importSet["javax.sql.DataSource"] = struct{}{}
	importSet["java.sql.Connection"] = struct{}{}
	importSet["java.sql.PreparedStatement"] = struct{}{}
	importSet["java.sql.SQLException"] = struct{}{}

	for _, query := range queries {
		if query.Cmd == ":one" {
			importSet["java.sql.ResultSet"] = struct{}{}
		}
		if query.Cmd == ":many" {
			importSet["java.util.List"] = struct{}{}
			importSet["java.util.ArrayList"] = struct{}{}
			importSet["java.sql.ResultSet"] = struct{}{}
		}
		if query.Cmd == ":execlastid" {
			importSet["java.sql.ResultSet"] = struct{}{}
		}
		if query.Cmd == ":execresult" {
			importSet["java.util.Map"] = struct{}{}
			importSet["java.sql.ResultSet"] = struct{}{}
		}
	}

	std := stdImports(uses)
	if r := hasEnum(); r > 0 {
		std["java.sql.Types"] = struct{}{}
		if r == 2 {
			std["java.util.Arrays"] = struct{}{}
			std["java.util.stream.Collectors"] = struct{}{}
		}
	}

	for i := range std {
		importSet[i] = struct{}{}
	}

	importSlice := make([]string, 0, len(importSet))
	for s := range importSet {
		importSlice = append(importSlice, s)
	}

	sort.Strings(importSlice)
	return importSlice
}

func stdImports(uses func(name string) bool) map[string]struct{} {
	std := make(map[string]struct{})
	if uses("Instant") {
		std["java.time.Instant"] = struct{}{}
		std["java.sql.Timestamp"] = struct{}{}
	}
	if uses("LocalDate") {
		std["java.time.LocalDate"] = struct{}{}
	}
	if uses("LocalTime") {
		std["java.time.LocalTime"] = struct{}{}
	}
	if uses("LocalDateTime") {
		std["java.time.LocalDateTime"] = struct{}{}
	}
	if uses("OffsetDateTime") {
		std["java.time.OffsetDateTime"] = struct{}{}
	}
	if uses("UUID") {
		std["java.util.UUID"] = struct{}{}
	}
	return std
}
