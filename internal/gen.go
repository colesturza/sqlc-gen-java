package java

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"

	"github.com/colesturza/sqlc-gen-java/internal/core"
)

//go:embed tmpl/javamodels.tmpl
var javaModelsTmpl string

//go:embed tmpl/javasql.tmpl
var javaSqlTmpl string

//go:embed tmpl/javaiface.tmpl
var javaIfaceTmpl string

func Offset(v int) int {
	return v + 1
}

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	var conf core.Config
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &conf); err != nil {
			return nil, err
		}
	}

	enums := core.BuildEnums(req)
	structs := core.BuildDataClasses(conf, req)
	queries, err := core.BuildQueries(req, structs)
	if err != nil {
		return nil, err
	}

	i := &core.Importer{
		Settings:    req.Settings,
		Enums:       enums,
		DataClasses: structs,
		Queries:     queries,
	}

	funcMap := template.FuncMap{
		"lowerTitle": sdk.LowerTitle,
		"comment":    sdk.DoubleSlashComment,
		"imports":    i.Imports,
		"offset":     Offset,
	}

	modelsFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaModelsTmpl))
	sqlFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaSqlTmpl))
	ifaceFile := template.Must(template.New("table").Funcs(funcMap).Parse(javaIfaceTmpl))

	core.DefaultImporter = i

	tctx := core.JavaTmplCtx{
		Settings:    req.Settings,
		Q:           `"""`,
		Package:     conf.Package,
		Queries:     queries,
		Enums:       enums,
		DataClasses: structs,
		SqlcVersion: req.SqlcVersion,
	}

	output := map[string]string{}

	execute := func(name string, t *template.Template) error {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		tctx.SourceName = name
		err := t.Execute(w, tctx)
		w.Flush()
		if err != nil {
			return err
		}
		if !strings.HasSuffix(name, ".java") {
			name += ".java"
		}
		output[name] = core.JavaFormat(b.String())
		return nil
	}

	if err := execute("Models.java", modelsFile); err != nil {
		return nil, err
	}
	if err := execute("Queries.java", ifaceFile); err != nil {
		return nil, err
	}
	if err := execute("QueriesImpl.java", sqlFile); err != nil {
		return nil, err
	}

	resp := plugin.GenerateResponse{}

	for filename, code := range output {
		resp.Files = append(resp.Files, &plugin.File{
			Name:     filename,
			Contents: []byte(code),
		})
	}

	return &resp, nil
}
