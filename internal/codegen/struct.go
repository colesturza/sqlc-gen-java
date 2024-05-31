package codegen

import "github.com/sqlc-dev/plugin-sdk-go/plugin"

type Struct struct {
	Table   *plugin.Identifier
	Name    string
	Fields  []Field
	Comment string
}
