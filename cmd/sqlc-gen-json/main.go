// sqlc-gen-json is a sqlc plugin that writes the incoming GenerateRequest
// to a single JSON file via protojson. Useful for debugging and inspecting
// the data sqlc passes to plugins.
package main

import (
	"github.com/sqlc-dev/plugin-sdk-go/codegen"

	"github.com/sqlc-contrib/sqlc-gen-json/internal/sqlc"
)

func main() {
	codegen.Run(sqlc.Generate)
}
