package opts

type Options struct {
	Package string `json:"package"`

	UseOptionalForNullableReturnValues bool `json:"use_optional_for_nullable_return_values"`
	JakartaEERepositories              bool `json:"jakarta_ee_repositories"`

	EmitExactTableNames         bool     `json:"emit_exact_table_names"`
	InflectionExcludeTableNames []string `json:"inflection_exclude_table_names"`

	OutputLogFile bool   `json:"ouptut_log_file"`
	LogLevel      string `json:"log_level"`
}
