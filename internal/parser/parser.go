package parser

import (
	"encoding/json"
	"errors"
	"fmt"
)

type KV struct {
	Key string
	V   any
}

type OrdererMap struct {
	Pairs []KV
}

func DecodeJSON(dec *json.Decoder) (OrdererMap, error) {
	dec.UseNumber()
	parsedData, err := ParseJSON(dec)
	if err != nil {
		return OrdererMap{}, err
	}

	data, ok := parsedData.(OrdererMap)
	if !ok {
		return OrdererMap{}, errors.New("cannot decode json data")
	}

	return data, nil
}

func ParseJSON(dec *json.Decoder) (any, error) {
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

				key := keyTok.(string)
				value, err := ParseJSON(dec)
				if err != nil {
					return nil, err
				}

				omap.Pairs = append(omap.Pairs, KV{Key: key, V: value})
			}
			_, _ = dec.Token() // consume '}'
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
			_, _ = dec.Token() // consume ']'
			return arr, nil
		}
	default:
		return tok, nil
	}

	return nil, fmt.Errorf("unexpected token: %v", tok)
}
