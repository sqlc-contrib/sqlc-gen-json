package sqlc_test

import (
	"context"
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/sqlc-contrib/sqlc-gen-json/internal/sqlc"
)

func sampleRequest(rawOpts string) *plugin.GenerateRequest {
	return &plugin.GenerateRequest{
		SqlcVersion: "v1.30.0",
		Settings:    &plugin.Settings{Engine: "postgresql"},
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{{
				Name: "public",
				Tables: []*plugin.Table{{
					Rel: &plugin.Identifier{Schema: "public", Name: "users"},
					Columns: []*plugin.Column{
						{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					},
				}},
			}},
		},
		Queries:       []*plugin.Query{{Name: "GetUser", Cmd: ":one"}},
		PluginOptions: []byte(rawOpts),
	}
}

var _ = Describe("Generate", func() {
	It("returns exactly one file with the default name and valid JSON", func() {
		resp, err := sqlc.Generate(context.Background(), sampleRequest(""))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Files).To(HaveLen(1))
		Expect(resp.Files[0].Name).To(Equal("generate_request.json"))

		var parsed map[string]any
		Expect(json.Unmarshal(resp.Files[0].Contents, &parsed)).To(Succeed())
		Expect(parsed).NotTo(BeEmpty())
	})

	It("round-trips the request through protojson", func() {
		req := sampleRequest("")
		resp, err := sqlc.Generate(context.Background(), req)
		Expect(err).NotTo(HaveOccurred())

		var got plugin.GenerateRequest
		Expect(protojson.Unmarshal(resp.Files[0].Contents, &got)).To(Succeed())
		Expect(proto.Equal(req, &got)).To(BeTrue())
	})

	It("honors a custom filename", func() {
		resp, err := sqlc.Generate(context.Background(), sampleRequest(`{"filename":"catalog.json"}`))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Files[0].Name).To(Equal("catalog.json"))
	})

	It("emits compact JSON when indent is the sentinel", func() {
		resp, err := sqlc.Generate(context.Background(), sampleRequest(`{"indent":"-"}`))
		Expect(err).NotTo(HaveOccurred())
		// Compact mode: no embedded newlines.
		Expect(strings.Contains(string(resp.Files[0].Contents), "\n")).To(BeFalse())
	})

	It("emits snake_case field names when use_proto_names is set", func() {
		// default_schema in plugin-sdk-go's proto has json_name=defaultSchema,
		// so it's the cleanest field to assert on for the snake/camel switch.
		defaultResp, err := sqlc.Generate(context.Background(), sampleRequest(""))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(defaultResp.Files[0].Contents)).To(ContainSubstring(`"defaultSchema"`))

		protoResp, err := sqlc.Generate(context.Background(), sampleRequest(`{"use_proto_names":true}`))
		Expect(err).NotTo(HaveOccurred())
		out := string(protoResp.Files[0].Contents)
		Expect(out).To(ContainSubstring(`"default_schema"`))
		Expect(out).NotTo(ContainSubstring(`"defaultSchema"`))
	})

	It("includes zero-value fields when emit_defaults is set", func() {
		// A column without NotNull/IsArray flags. With EmitDefaultValues,
		// those bool fields should appear in the JSON.
		req := &plugin.GenerateRequest{
			Catalog: &plugin.Catalog{
				Schemas: []*plugin.Schema{{
					Name: "public",
					Tables: []*plugin.Table{{
						Rel: &plugin.Identifier{Name: "t"},
						Columns: []*plugin.Column{
							{Name: "c", Type: &plugin.Identifier{Name: "text"}},
						},
					}},
				}},
			},
			PluginOptions: []byte(`{"emit_defaults":true}`),
		}
		resp, err := sqlc.Generate(context.Background(), req)
		Expect(err).NotTo(HaveOccurred())
		out := string(resp.Files[0].Contents)
		Expect(out).To(ContainSubstring(`"notNull"`))
		Expect(out).To(ContainSubstring(`"isArray"`))
	})

	It("returns a wrapped error for invalid options", func() {
		_, err := sqlc.Generate(context.Background(), sampleRequest(`{"nope":1}`))
		Expect(err).To(MatchError(ContainSubstring("sqlc-gen-json")))
	})
})
