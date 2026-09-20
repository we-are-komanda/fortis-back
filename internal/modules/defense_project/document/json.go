// Package document preserves uninterpreted project fields at the JSON boundary.
// Authorization and calculations consume the typed domain model, never extensions.
package document

import (
	"bytes"
	"encoding/json"
)

// Encode retains the submitted document and supplies missing legacy defaults.
// Only explicitly authoritative metadata replaces submitted fields.
func Encode(raw string, defaults any, metadata map[string]any) ([]byte, error) {
	b, err := json.Marshal(defaults)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(b, &fields); err != nil {
		return nil, err
	}
	if raw != "" {
		var original map[string]json.RawMessage
		if err = json.Unmarshal([]byte(raw), &original); err != nil {
			return nil, err
		}
		for key, value := range original {
			fields[key] = value
		}
	}
	for key, value := range metadata {
		b, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		fields[key] = b
	}
	// Transport tracing/time fields never participate in the saved business snapshot.
	delete(fields, "requestId")
	delete(fields, "generatedAt")
	return json.Marshal(fields)
}

// Canonical sorts object keys recursively while preserving array order and
// decimal text without routing unknown integers through float64.
func Canonical(raw []byte) ([]byte, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(out.Bytes(), []byte{'\n'}), nil
}
