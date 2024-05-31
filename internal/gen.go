package java

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"

	"github.com/colesturza/sqlc-gen-java/internal/codegen"
	"github.com/colesturza/sqlc-gen-java/internal/codegen/opts"
)

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

	resp := pb.GenerateResponse{}

	for filename, code := range output {
		resp.Files = append(resp.Files, &pb.File{
			Name:     filename,
			Contents: []byte(code),
		})
	}

	return &resp, nil
}
