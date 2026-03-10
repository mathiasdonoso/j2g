package cli

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/mathiasdonoso/j2g/internal/generator"
	"github.com/mathiasdonoso/j2g/internal/parser"
)

type J2G struct {
	Input      io.Reader
	Output     io.Writer
	StructName string
}

func (j *J2G) Start() error {
	slog.Debug("decoding json")
	result, err := parser.DecodeJSON(j.Input)
	if err != nil {
		return err
	}
	slog.Debug("json decoded successfully")

	slog.Debug("building structs")
	builder := generator.Builder{StructName: j.StructName}
	var s string

	switch v := result.(type) {
	case parser.OrdererMap:
		s, err = builder.BuildStruct(v)
	case []any:
		s, err = builder.BuildFromArray(v)
	default:
		return fmt.Errorf("unsupported root JSON type: %T", result)
	}

	if err != nil {
		return err
	}
	slog.Debug("structs created successfully")

	slog.Debug("Writing generated structs")
	_, err = j.Output.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}
