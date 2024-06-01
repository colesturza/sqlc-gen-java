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

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"

	"github.com/colesturza/sqlc-gen-java/internal/codegen"
	"github.com/colesturza/sqlc-gen-java/internal/codegen/opts"
)

//go:embed codegen/templates/javaenum.tmpl
var javaEnumTmpl string

//go:embed codegen/templates/javapojo.tmpl
var javaPOJOTmpl string

func Generate(ctx context.Context, req *pb.GenerateRequest) (*pb.GenerateResponse, error) {

	var options opts.Options
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &options); err != nil {
			return nil, err
		}
	}

	// configure logging
	var buf bytes.Buffer
	logHandlerOptions := slog.HandlerOptions{
		Level: slog.LevelWarn,
	}
	jsonLogHandler := slog.NewJSONHandler(&buf, &logHandlerOptions)
	slog.SetDefault(slog.New(jsonLogHandler))

	slog.Error("An error")
	slog.Warn("A warning")

	enums := codegen.BuildJavaEnums(req, &options)
	classes := codegen.BuildJavaClasses(req, &options)

	enumsJSON, _ := json.MarshalIndent(enums, "", "\t")
	classesJSON, _ := json.MarshalIndent(classes, "", "\t")

	output := map[string]string{}

	output["log"] = buf.String()
	output["enums.txt"] = string(enumsJSON)
	output["classes.txt"] = string(classesJSON)

	funcMap := template.FuncMap{
		"title":      sdk.Title,
		"lowerTitle": sdk.LowerTitle,
		"comment":    sdk.DoubleSlashComment,
	}

	enumFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaEnumTmpl))
	classFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaPOJOTmpl))

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
		output[name] = b.String()
		return nil
	}

	for _, enum := range enums {
		data := struct {
			Package     string
			SqlcVersion string
			codegen.Enum
		}{
			Package:     options.Package,
			SqlcVersion: req.SqlcVersion,
			Enum:        enum,
		}
		if err := execute(enum.Name, enumFile, data); err != nil {
			return nil, err
		}
	}

	for _, class := range classes {
		data := struct {
			Package     string
			SqlcVersion string
			codegen.Struct
		}{
			Package:     options.Package,
			SqlcVersion: req.SqlcVersion,
			Struct:      class,
		}
		if err := execute(class.Name, classFile, data); err != nil {
			return nil, err
		}
	}

	resp := pb.GenerateResponse{}

	for filename, code := range output {
		resp.Files = append(resp.Files, &pb.File{
			Name:     filename,
			Contents: []byte(code),
		})
	}

	return &resp, nil
}
