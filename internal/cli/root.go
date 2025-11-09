package cli

import (
	"encoding/json"
	"io"

	"github.com/mathiasdonoso/j2g/internal/generator"
	"github.com/mathiasdonoso/j2g/internal/parser"
)

type J2G struct {
	Input  io.Reader
	Output io.Writer
}

func (j *J2G) Start() error {
	result, err := parser.DecodeJSON(json.NewDecoder(j.Input))
	if err != nil {
		return err
	}

	builder := generator.Builder{}
	s, err := builder.BuildStruct(result)
	if err != nil {
		return err
	}

	_, err = j.Output.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}
