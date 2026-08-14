# inari-api

The compatibility contract for Inari: protobuf definitions (agent EventStream protocol, plugin contract) + OpenAPI specs, with generated Go/TS clients (plan §6 #5).

Stack: Buf, protoc (connect/gRPC), oapi-codegen; generated Go + TypeScript clients

Part of the **Inari** multi-tenant Internal Developer Platform (GitHub org `7K-Inari`).
Canonical architecture & development plan: [inari-docs/docs/architecture/inari-platform-plan.md](https://github.com/7K-Inari/inari-docs/blob/main/docs/architecture/inari-platform-plan.md)

## Layout

- `proto/inari/agent/v1/` — agent EventStream protocol (bidirectional gRPC stream, CloudEvents-style envelope)
- `proto/inari/plugin/v1/` — versioned plugin contract (hashicorp go-plugin model)
- `openapi/openapi.yaml` — REST surface for console/CLI (OpenAPI 3.0)
- `gen/go/`, `gen/ts/` — generated Go and TypeScript clients, **committed to the repo** (see below)

## Development

Prerequisites: `buf`, Go, Node.js, `oapi-codegen`.

- `make generate` — regenerate everything (`buf generate`, oapi-codegen, Orval)
- `make lint` — `buf lint` + `redocly lint`
- `make breaking` — `buf breaking` against `main`
- `make build` / `make test` — compile and smoke-test both clients
- `make ci` — the full local gate

Note: `npm ci` must run with dev dependencies (`npm ci --include=dev`, or unset `NODE_ENV=production`); Orval and TypeScript are devDependencies.

## Generated code policy

Generated output under `gen/` is **committed**, not built only in CI. Consumers pin
`go get github.com/7K-Inari/inari-api@<tag>` and the npm package directly at a git tag,
so the exact contract surface must be reviewable in PRs and resolvable without codegen in
consumer CI. A CI drift check (`make generate && git diff --exit-code`) guarantees
committed code always matches the sources.

## Versioning policy

- Releases are SemVer git tags `vX.Y.Z`; commit messages follow Conventional Commits.
- **Protobuf: additive changes only within a package version.** New fields, messages, or
  RPCs may be added to `inari.*.v1`. Any breaking change (removing/renumbering a field,
  changing a type, renaming a package or service, changing request/response types) requires
  a **new package version** (`v1` → `v2`) — never edit a released version in place.
  `buf breaking` runs in CI against `main` to enforce this.
- **OpenAPI: additive changes only** (new endpoints, new optional fields). Breaking changes
  require a new base path (`/api/v2`) and spec version.
- Go modules at v2+ use the `/v2` import-path suffix per Go module semantics; the release
  workflow verifies tag/module-path consistency.

## Consuming

```sh
go get github.com/7K-Inari/inari-api@v0.1.0
npm install @7k-inari/api-client@0.1.0
```

Tagging `vX.Y.Z` runs the release workflow: it re-verifies the full CI gate and publishes
the npm package (requires the `NPM_TOKEN` secret). The Go module is consumable from the
tag itself.
