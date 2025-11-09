package cli

import (
	"encoding/json"
	"io"
	"log/slog"

	"github.com/mathiasdonoso/j2g/internal/generator"
	"github.com/mathiasdonoso/j2g/internal/parser"
)

type J2G struct {
	Input  io.Reader
	Output io.Writer
}

func (j *J2G) Start() error {
	slog.Debug("decoding json")
	result, err := parser.DecodeJSON(json.NewDecoder(j.Input))
	if err != nil {
		return err
	}
	slog.Debug("json decoded successfully")

	slog.Debug("building structs")
	builder := generator.Builder{}
	s, err := builder.BuildStruct(result)
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
