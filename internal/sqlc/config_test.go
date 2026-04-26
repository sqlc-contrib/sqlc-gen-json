package sqlc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sqlc-contrib/sqlc-gen-json/internal/sqlc"
)

var _ = Describe("ParseOptions", func() {
	When("the payload is empty", func() {
		It("returns defaults", func() {
			opts, err := sqlc.ParseOptions(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.Filename).To(Equal("generate_request.json"))
			Expect(opts.ResolvedIndent()).To(Equal("  "))
			Expect(opts.UseProtoNames).To(BeFalse())
			Expect(opts.EmitDefaults).To(BeFalse())
		})
	})

	When("the payload sets every option", func() {
		It("decodes them verbatim", func() {
			data := []byte(`{
				"filename":"out.json",
				"indent":"\t",
				"use_proto_names":true,
				"emit_defaults":true
			}`)
			opts, err := sqlc.ParseOptions(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.Filename).To(Equal("out.json"))
			Expect(opts.ResolvedIndent()).To(Equal("\t"))
			Expect(opts.UseProtoNames).To(BeTrue())
			Expect(opts.EmitDefaults).To(BeTrue())
		})
	})

	When(`the indent is the "-" sentinel`, func() {
		It("resolves to no indentation", func() {
			opts, err := sqlc.ParseOptions([]byte(`{"indent":"-"}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.ResolvedIndent()).To(Equal(""))
		})
	})

	When("the payload has an unknown field", func() {
		It("rejects it to catch typos", func() {
			_, err := sqlc.ParseOptions([]byte(`{"nope":1}`))
			Expect(err).To(HaveOccurred())
		})
	})

	When("the payload is not valid JSON", func() {
		It("returns a decode error", func() {
			_, err := sqlc.ParseOptions([]byte(`{not json`))
			Expect(err).To(HaveOccurred())
		})
	})
})
