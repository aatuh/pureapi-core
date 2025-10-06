package querydec

import (
	"net/url"
)

// Decoder turns url.Values into a normalized map tree.
type Decoder interface {
	Decode(values url.Values) (map[string]any, error)
}

// PlainDecoder implements `?x=1&y=a` into flat map.
type PlainDecoder struct{}

// Decode converts URL values to a flat map.
//
// Parameters:
//   - v: The URL values to decode.
//
// Returns:
//   - map[string]any: The decoded query parameters.
//   - error: An error if decoding fails.
func (d PlainDecoder) Decode(v url.Values) (map[string]any, error) {
	out := make(map[string]any, len(v))
	for k := range v {
		vals := v[k]
		if len(vals) == 1 {
			out[k] = vals[0]
			continue
		}
		out[k] = vals
	}
	return out, nil
}
