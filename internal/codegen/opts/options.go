package opts

import (
	"encoding/json"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Options struct {
	Package string `json:"package"`

	UseOptionalForNullableReturnValues bool `json:"use_optional_for_nullable_return_values"`
	JakartaEERepositories              bool `json:"jakarta_ee_repositories"`

	EmitExactTableNames         bool      `json:"emit_exact_table_names"`
	InflectionExcludeTableNames []string  `json:"inflection_exclude_table_names"`
	EmitSqlAsComment            bool      `json:"emit_sql_as_comment"`
	QueryParameterLimit         *int32    `json:"query_parameter_limit"`
	Initialisms                 *[]string `json:"initialisms,omitempty" yaml:"initialisms"`

	InitialismsMap map[string]struct{} `json:"-" yaml:"-"`

	OutputLogFile bool   `json:"ouptut_log_file"`
	LogLevel      string `json:"log_level"`
}

func Parse(req *plugin.GenerateRequest) (*Options, error) {
	options, err := parseOpts(req)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func parseOpts(req *plugin.GenerateRequest) (*Options, error) {
	var options Options
	if len(req.PluginOptions) == 0 {
		return &options, nil
	}
	if err := json.Unmarshal(req.PluginOptions, &options); err != nil {
		return nil, fmt.Errorf("unmarshalling plugin options: %w", err)
	}

	if options.QueryParameterLimit == nil {
		options.QueryParameterLimit = new(int32)
		*options.QueryParameterLimit = 1
	}

	if options.Initialisms == nil {
		options.Initialisms = new([]string)
		*options.Initialisms = []string{"id"}
	}

	options.InitialismsMap = map[string]struct{}{}
	for _, initial := range *options.Initialisms {
		options.InitialismsMap[initial] = struct{}{}
	}

	return &options, nil
}

func ValidateOpts(opts *Options) error {
	if *opts.QueryParameterLimit < 0 {
		return fmt.Errorf("invalid options: query parameter limit must not be negative")
	}

	return nil
}
