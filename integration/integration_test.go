package integration_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// runPlugin invokes the built plugin binary with the given GenerateRequest
// and returns the decoded response. It mirrors how sqlc invokes process
// plugins: proto-marshalled request on stdin, proto-marshalled response
// on stdout.
func runPlugin(req *plugin.GenerateRequest) *plugin.GenerateResponse {
	reqBytes, err := proto.Marshal(req)
	Expect(err).NotTo(HaveOccurred())

	cmd := exec.Command(binaryPath)
	cmd.Stdin = bytes.NewReader(reqBytes)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	Expect(cmd.Run()).To(Succeed(), "plugin failed: %s", stderr.String())

	var resp plugin.GenerateResponse
	Expect(proto.Unmarshal(stdout.Bytes(), &resp)).To(Succeed())
	return &resp
}

// runPluginExpectErr runs the plugin expecting a non-zero exit.
func runPluginExpectErr(req *plugin.GenerateRequest) string {
	reqBytes, err := proto.Marshal(req)
	Expect(err).NotTo(HaveOccurred())

	cmd := exec.Command(binaryPath)
	cmd.Stdin = bytes.NewReader(reqBytes)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	Expect(err).To(HaveOccurred())
	return stderr.String()
}

// fixtureRequest builds a small reusable GenerateRequest mirroring a
// minimal postgres schema; the JSON dump for this request is used by
// every test case below.
func fixtureRequest(opts map[string]any) *plugin.GenerateRequest {
	var raw []byte
	if opts != nil {
		var err error
		raw, err = json.Marshal(opts)
		Expect(err).NotTo(HaveOccurred())
	}
	return &plugin.GenerateRequest{
		SqlcVersion:   "v1.30.0",
		Settings:      &plugin.Settings{Engine: "postgresql"},
		PluginOptions: raw,
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{{
				Name: "public",
				Tables: []*plugin.Table{{
					Rel: &plugin.Identifier{Schema: "public", Name: "users"},
					Columns: []*plugin.Column{
						{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
						{Name: "email", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
					},
				}},
			}},
		},
		Queries: []*plugin.Query{
			{Name: "GetUser", Cmd: ":one"},
		},
	}
}

var _ = Describe("Plugin binary", func() {
	It("emits a valid JSON dump with default options", func() {
		req := fixtureRequest(nil)
		resp := runPlugin(req)

		Expect(resp.Files).To(HaveLen(1))
		Expect(resp.Files[0].Name).To(Equal("generate_request.json"))

		// Round-trip: the JSON must reconstruct the request exactly.
		var got plugin.GenerateRequest
		Expect(protojson.Unmarshal(resp.Files[0].Contents, &got)).To(Succeed())
		Expect(proto.Equal(req, &got)).To(BeTrue())
	})

	It("honors a custom filename", func() {
		resp := runPlugin(fixtureRequest(map[string]any{"filename": "catalog.json"}))
		Expect(resp.Files).To(HaveLen(1))
		Expect(resp.Files[0].Name).To(Equal("catalog.json"))
	})

	It("emits compact JSON when indent is the sentinel", func() {
		resp := runPlugin(fixtureRequest(map[string]any{"indent": "-"}))
		Expect(strings.Contains(string(resp.Files[0].Contents), "\n")).To(BeFalse())
	})

	It("emits proto (snake_case) field names when use_proto_names is set", func() {
		resp := runPlugin(fixtureRequest(map[string]any{"use_proto_names": true}))
		out := string(resp.Files[0].Contents)
		Expect(out).To(ContainSubstring(`"default_schema"`))
		Expect(out).NotTo(ContainSubstring(`"defaultSchema"`))
	})

	It("includes zero-value fields when emit_defaults is set", func() {
		resp := runPlugin(fixtureRequest(map[string]any{"emit_defaults": true}))
		out := string(resp.Files[0].Contents)
		Expect(out).To(ContainSubstring(`"notNull"`))
		Expect(out).To(ContainSubstring(`"isArray"`))
	})

	It("rejects unknown options", func() {
		req := fixtureRequest(nil)
		req.PluginOptions = []byte(`{"nope":1}`)
		out := runPluginExpectErr(req)
		Expect(out).To(ContainSubstring("sqlc-gen-json"))
	})
})
