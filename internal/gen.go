package java

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"

	"github.com/colesturza/sqlc-gen-java/internal/codegen"
	"github.com/colesturza/sqlc-gen-java/internal/codegen/opts"
	"github.com/colesturza/sqlc-gen-java/internal/inflection"
)

var version = "dev"

//go:embed codegen/templates/javaenum.tmpl
var javaEnumTmpl string

//go:embed codegen/templates/javapojo.tmpl
var javaPOJOTmpl string

//go:embed codegen/templates/javaiface.tmpl
var javaIfaceTmpl string

//go:embed codegen/templates/javasql.tmpl
var javaSQLTmpl string

func Generate(ctx context.Context, req *pb.GenerateRequest) (*pb.GenerateResponse, error) {
	options, err := opts.Parse(req)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	logHandlerOptions := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	jsonLogHandler := slog.NewJSONHandler(&buf, &logHandlerOptions)
	slog.SetDefault(slog.New(jsonLogHandler))

	enums := codegen.BuildJavaEnums(req, options)
	classes := codegen.BuildJavaClasses(req, options)
	queries, err := codegen.BuildQueries(req, options, classes)
	if err != nil {
		return nil, err
	}

	enumsJSON, _ := json.MarshalIndent(enums, "", "\t")
	classesJSON, _ := json.MarshalIndent(classes, "", "\t")
	queriesJSON, _ := json.MarshalIndent(queries, "", "\t")

	output := map[string]string{}
	output["enums.txt"] = string(enumsJSON)
	output["classes.txt"] = string(classesJSON)
	output["queries.txt"] = string(queriesJSON)

	singular := func(s string) string {
		return inflection.Singular(inflection.SingularParams{
			Name:       s,
			Exclusions: options.InflectionExcludeTableNames,
		})
	}

	importer := codegen.NewEnumImporter(options, enums, classes, queries)

	customFuncMap := template.FuncMap{
		"title":      sdk.Title,
		"lowerTitle": sdk.LowerTitle,
		"comment":    sdk.DoubleSlashComment,
		"singular":   singular,
		"imports":    importer.Imports,
	}

	funcMap := sprig.FuncMap()
	for k, v := range customFuncMap {
		funcMap[k] = v
	}

	enumFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaEnumTmpl))
	classFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaPOJOTmpl))
	ifaceFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaIfaceTmpl))
	sqlFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaSQLTmpl))

	execute := func(name string, t *template.Template, data any) error {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		err := t.Execute(w, data)
		w.Flush()
		if err != nil {
			return err
		}
		if !strings.HasSuffix(name, ".java") {
			name += ".java"
		}
		output[name] = codegen.JavaFormat(b.String())
		return nil
	}

	for _, enum := range enums {
		data := struct {
			SourceName         string
			Package            string
			SqlcVersion        string
			SqlcGenJavaVersion string
			codegen.Enum
		}{
			SourceName:         enum.Name + ".java",
			Package:            options.Package,
			SqlcVersion:        req.SqlcVersion,
			SqlcGenJavaVersion: version,
			Enum:               enum,
		}
		if err := execute(enum.Name, enumFile, data); err != nil {
			return nil, err
		}
	}

	createStructTmplData := func(s codegen.Struct, emitBuilder bool) any {
		return struct {
			SourceName         string
			Package            string
			SqlcVersion        string
			SqlcGenJavaVersion string
			EmitBuilder        bool
			codegen.Struct
		}{
			SourceName:         s.Name + ".java",
			Package:            options.Package,
			SqlcVersion:        req.SqlcVersion,
			SqlcGenJavaVersion: version,
			EmitBuilder:        emitBuilder,
			Struct:             s,
		}
	}

	for _, class := range classes {
		data := createStructTmplData(class, false)
		if err := execute(class.Name, classFile, data); err != nil {
			return nil, err
		}
	}

	for _, query := range queries {
		if query.Arg.IsStruct() && query.Arg.EmitStruct() {
			data := createStructTmplData(*query.Arg.Struct, true)
			if err := execute(query.Arg.Struct.Name, classFile, data); err != nil {
				return nil, err
			}
		}
		if query.Ret.IsStruct() && query.Ret.EmitStruct() {
			data := createStructTmplData(*query.Ret.Struct, false)
			if err := execute(query.Ret.Struct.Name, classFile, data); err != nil {
				return nil, err
			}
		}
	}

	{
		data := struct {
			SourceName         string
			Package            string
			SqlcVersion        string
			SqlcGenJavaVersion string
			Queries            []codegen.Query
		}{
			SourceName:         "Queries.java",
			Package:            options.Package,
			SqlcVersion:        req.SqlcVersion,
			SqlcGenJavaVersion: version,
			Queries:            queries,
		}
		if err := execute("Queries.java", ifaceFile, data); err != nil {
			return nil, err
		}
	}

	{
		data := struct {
			SourceName         string
			Package            string
			SqlcVersion        string
			SqlcGenJavaVersion string
			Queries            []codegen.Query
		}{
			SourceName:         "QueriesImpl.java",
			Package:            options.Package,
			SqlcVersion:        req.SqlcVersion,
			SqlcGenJavaVersion: version,
			Queries:            queries,
		}
		if err := execute("QueriesImpl.java", sqlFile, data); err != nil {
			return nil, err
		}
	}

	output["log"] = buf.String()

	resp := pb.GenerateResponse{}

	for filename, code := range output {
		resp.Files = append(resp.Files, &pb.File{
			Name:     filename,
			Contents: []byte(code),
		})
	}

	return &resp, nil
}
