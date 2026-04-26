// Package sqlc provides the sqlc codegen plugin entry point and
// configuration parsing for sqlc-gen-json.
package sqlc

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// PluginName is the name under which this plugin is typically registered
// in sqlc.yaml.
const PluginName = "json"

const (
	defaultFilename = "generate_request.json"
	defaultIndent   = "  "
	// CompactIndentSentinel disables indentation when used as the value of
	// the "indent" option, producing a single-line JSON document.
	CompactIndentSentinel = "-"
)

// Options holds plugin-specific options decoded from the JSON payload that
// sqlc sends in GenerateRequest.PluginOptions.
type Options struct {
	// Filename is the name of the JSON file emitted in the response.
	// Defaults to "generate_request.json".
	Filename string `json:"filename,omitempty"`
	// Indent is the per-level indent string passed to protojson. Defaults
	// to two spaces. The sentinel "-" disables indentation entirely.
	Indent string `json:"indent,omitempty"`
	// UseProtoNames toggles protojson's UseProtoNames: when true, fields
	// are emitted with their proto (snake_case) names instead of the
	// default lowerCamelCase JSON names.
	UseProtoNames bool `json:"use_proto_names,omitempty"`
	// EmitDefaults toggles protojson's EmitDefaultValues: when true,
	// scalar fields with their zero value are included in the output.
	EmitDefaults bool `json:"emit_defaults,omitempty"`
}

// ParseOptions decodes the JSON plugin options payload. Unknown fields are
// rejected to catch typos in sqlc.yaml. Defaults are applied for any
// option left unset.
func ParseOptions(data []byte) (Options, error) {
	var opts Options
	if len(data) > 0 {
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&opts); err != nil {
			return opts, fmt.Errorf("decode plugin options: %w", err)
		}
	}
	opts.applyDefaults()
	return opts, nil
}

func (o *Options) applyDefaults() {
	if o.Filename == "" {
		o.Filename = defaultFilename
	}
	if o.Indent == "" {
		o.Indent = defaultIndent
	}
}

// ResolvedIndent returns the indent string to pass to protojson.
// CompactIndentSentinel maps to the empty string (no indentation).
func (o Options) ResolvedIndent() string {
	if o.Indent == CompactIndentSentinel {
		return ""
	}
	return o.Indent
}
