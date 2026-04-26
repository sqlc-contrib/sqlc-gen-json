package sqlc

import (
	"context"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"google.golang.org/protobuf/encoding/protojson"
)

// Generate is the sqlc codegen entry point. It marshals the incoming
// GenerateRequest to a single JSON file via protojson. Output filename
// and formatting are controlled by plugin options.
func Generate(_ context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	opts, err := ParseOptions(req.GetPluginOptions())
	if err != nil {
		return nil, fmt.Errorf("sqlc-gen-json: %w", err)
	}

	m := protojson.MarshalOptions{
		Indent:            opts.ResolvedIndent(),
		UseProtoNames:     opts.UseProtoNames,
		EmitDefaultValues: opts.EmitDefaults,
	}
	data, err := m.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("sqlc-gen-json: marshal: %w", err)
	}

	return &plugin.GenerateResponse{
		Files: []*plugin.File{{Name: opts.Filename, Contents: data}},
	}, nil
}
