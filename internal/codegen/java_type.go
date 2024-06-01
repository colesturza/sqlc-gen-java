package codegen

import (
	"fmt"

	"github.com/colesturza/sqlc-gen-java/internal/codegen/opts"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

type javaType struct {
	Name     string
	IsEnum   bool
	IsArray  bool
	IsNull   bool
	DataType string
	Engine   string
}

func (t javaType) String() string {
	v := t.Name
	if t.IsArray {
		v = fmt.Sprintf("java.util.List<%s>", v)
	} else if t.IsNull {
		v = fmt.Sprintf("java.util.Optional<%s>", v)
	}
	return v
}

func (t javaType) jdbcSetter() string {
	return "set" + t.jdbcType()
}

func (t javaType) jdbcType() string {
	if t.IsArray {
		return "Array"
	}
	if t.IsEnum || t.IsTime() {
		return "Object"
	}
	if t.IsInstant() {
		return "Timestamp"
	}
	return t.Name
}

func (t javaType) IsTime() bool {
	return t.Name == "LocalDate" || t.Name == "LocalDateTime" || t.Name == "LocalTime" || t.Name == "OffsetDateTime"
}

func (t javaType) IsInstant() bool {
	return t.Name == "Instant"
}

func (t javaType) IsUUID() bool {
	return t.Name == "UUID"
}

func (t javaType) IsBigDecimal() bool {
	return t.Name == "java.math.BigDecimal"
}

func makeType(req *plugin.GenerateRequest, options *opts.Options, col *plugin.Column) javaType {
	typ, isEnum := javaInnerType(req, options, col)
	return javaType{
		Name:     typ,
		IsEnum:   isEnum,
		IsArray:  col.IsArray,
		IsNull:   !col.NotNull,
		DataType: sdk.DataType(col.Type),
		Engine:   req.Settings.Engine,
	}
}

func javaInnerType(req *plugin.GenerateRequest, options *opts.Options, col *plugin.Column) (string, bool) {
	// TODO: Extend the engine interface to handle types
	switch req.Settings.Engine {
	case "mysql":
		return mysqlType(req, col, options)
	case "postgresql":
		return postgresType(req, col, options)
	default:
		return "Any", false
	}
}
