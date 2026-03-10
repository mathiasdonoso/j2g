package parser

import (
	"encoding/json"
	"fmt"
	"io"
)

type KV struct {
	Key string
	V   any
}

type OrdererMap struct {
	Pairs []KV
}

func DecodeJSON(r io.Reader) (any, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	parsedData, err := ParseJSON(dec)
	if err != nil {
		return nil, err
	}

	return parsedData, nil
}

// tokenReader is a subset of *json.Decoder used by ParseJSON. It is an
// interface so tests can inject controlled token sequences to exercise
// error branches that json.Decoder itself can never produce.
type tokenReader interface {
	Token() (json.Token, error)
	More() bool
}

func ParseJSON(dec tokenReader) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", err
	}

	switch delim := tok.(type) {
	case json.Delim:
		switch delim {
		case '{':
			omap := OrdererMap{}
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}

				key, ok := keyTok.(string)
				if !ok {
					return nil, fmt.Errorf("expected string key, got %T", keyTok)
				}

				value, err := ParseJSON(dec)
				if err != nil {
					return nil, err
				}

				omap.Pairs = append(omap.Pairs, KV{Key: key, V: value})
			}
			if _, err := dec.Token(); err != nil {
				return nil, fmt.Errorf("reading closing delimiter: %w", err)
			}
			return omap, nil

		case '[':
			var arr []any
			for dec.More() {
				val, err := ParseJSON(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, val)
			}
			if _, err := dec.Token(); err != nil {
				return nil, fmt.Errorf("reading closing delimiter: %w", err)
			}
			return arr, nil

		default:
			// This should never happen for valid token streams: '}' and ']' are
			// consumed by the closing-delimiter reads above and never reach this switch.
			return nil, fmt.Errorf("unexpected delimiter: %v", delim)
		}

	default:
		// All token types from encoding/json (Delim, json.Number, bool, string, nil)
		// are handled above. This branch is unreachable for well-formed JSON.
		return tok, nil
	}
}
