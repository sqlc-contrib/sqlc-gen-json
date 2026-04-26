# sqlc-gen-json

[![CI](https://github.com/sqlc-contrib/sqlc-gen-json/actions/workflows/ci.yml/badge.svg)](https://github.com/sqlc-contrib/sqlc-gen-json/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/sqlc-contrib/sqlc-gen-json?include_prereleases)](https://github.com/sqlc-contrib/sqlc-gen-json/releases)
[![License](https://img.shields.io/github/license/sqlc-contrib/sqlc-gen-json)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![sqlc](https://img.shields.io/badge/sqlc-compatible-blue)](https://sqlc.dev)

A [sqlc](https://sqlc.dev) plugin that writes the `GenerateRequest` sqlc would
have passed to your codegen plugin out as a single JSON file.

Useful for:

- **Debugging plugin pipelines** — see exactly what schema, queries, settings,
  and plugin options sqlc parsed before invoking your generator.
- **Prototyping new plugins** — capture a real request, then iterate on a
  generator script offline against the saved JSON.
- **Schema snapshots** — keep a versioned JSON view of your database catalog
  alongside the SQL.

## Installation

The plugin ships as a WASM artifact. Reference it from `sqlc.yaml` with the
release URL and its sha256:

```yaml
version: "2"
plugins:
  - name: json
    wasm:
      url: https://github.com/sqlc-contrib/sqlc-gen-json/releases/download/v0.1.0/sqlc-gen-json.wasm
      sha256: <sha256 from the release assets>
```

Both the `.wasm` and a `.wasm.sha256` sidecar are published with each
release.

## Configuration

```yaml
sql:
  - engine: postgresql
    schema: db/schema.sql
    queries: db/queries.sql
    codegen:
      - plugin: json
        out: ./gen
        options:
          filename: generate_request.json # optional
          indent: "  " # optional; "-" disables indent
          use_proto_names: false # optional
          emit_defaults: false # optional
```

| Option            | Type    | Default                  | Description                                                                                                       |
| ----------------- | ------- | ------------------------ | ----------------------------------------------------------------------------------------------------------------- |
| `filename`        | string  | `generate_request.json`  | Name of the emitted file.                                                                                         |
| `indent`          | string  | `"  "` (two spaces)      | Per-level indent string passed to `protojson`. Use the sentinel `-` to emit compact (single-line) JSON.            |
| `use_proto_names` | boolean | `false`                  | When `true`, emit field names from the proto descriptor (snake_case) instead of `protojson`'s default JSON names. |
| `emit_defaults`   | boolean | `false`                  | When `true`, include scalar fields whose value is the zero value.                                                 |

Unknown top-level options are rejected.

## Output

The output is the full `plugin.GenerateRequest` from
[`plugin-sdk-go`](https://pkg.go.dev/github.com/sqlc-dev/plugin-sdk-go/plugin)
serialized via `protojson`. A trimmed example for a single-table schema:

```json
{
  "settings": { "engine": "postgresql" },
  "catalog": {
    "defaultSchema": "public",
    "schemas": [
      {
        "name": "public",
        "tables": [
          {
            "rel": { "schema": "public", "name": "users" },
            "columns": [
              { "name": "id", "type": { "name": "int4" }, "notNull": true },
              { "name": "email", "type": { "name": "text" }, "notNull": true }
            ]
          }
        ]
      }
    ]
  },
  "queries": [{ "name": "GetUser", "cmd": ":one" }],
  "sqlc_version": "v1.30.0"
}
```

Note that `plugin-sdk-go`'s proto descriptors do not consistently set
`json_name`, so default output mixes camelCase (e.g. `defaultSchema`,
`notNull`) and snake_case (e.g. `sqlc_version`). Set `use_proto_names: true`
to force snake_case across the board.

## Caveats

- **Output is not byte-stable across runs.** `protojson.Marshal`
  intentionally injects random whitespace to discourage byte-stable
  diffing. For diffs, use a JSON-aware tool (`jq -S`, `diff <(jq -S . a)
  <(jq -S . b)`, etc.) rather than direct text comparison.
- **No `formatter_cmd`.** sqlc plugins cannot spawn subprocesses. If you
  want canonical formatting, run `jq -S . generate_request.json | sponge
  generate_request.json` (or similar) as a post-step.

## Development

```bash
nix develop
go tool ginkgo run -r -coverprofile=coverage.out -covermode=atomic ./...
```

Build the WASM artifact locally:

```bash
nix build .#wasm
sha256sum result/bin/sqlc-gen-json.wasm
```

A native binary build is also exposed for convenience (used by the
integration tests via `gexec.Build`):

```bash
nix build .       # default == native build
./result/bin/sqlc-gen-json --help
```

## License

[MIT](LICENSE)
