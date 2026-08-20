#!/usr/bin/env bash
# Import the full REST surface from inari-server's offline OpenAPI export and
# merge it with the contract-owned meta fragment (health endpoints, info,
# servers) into openapi/openapi.yaml.
#
# Usage:
#   scripts/import-openapi.sh [path-to-exported-openapi.yaml]
#
# If no path is given, inari-server is cloned (INARI_SERVER_REF, default:
# latest tag, falling back to main) and `go run ./cmd/export-openapi` is run.
#
# Requires: yq v4 (mikefarah), git, go (only when exporting from source).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRAGMENT="$REPO_ROOT/openapi/meta.fragment.yaml"
OUT="$REPO_ROOT/openapi/openapi.yaml"
EXPORTED="${1:-}"

if [[ -z "$EXPORTED" ]]; then
  TMP="$(mktemp -d)"
  trap 'rm -rf "$TMP"' EXIT
  REF="${INARI_SERVER_REF:-}"
  if [[ -z "$REF" ]]; then
    REF="$(git ls-remote --tags --sort=-v:refname https://github.com/7K-Inari/inari-server.git 'v*' | head -1 | sed 's/.*\///')"
    REF="${REF:-main}"
  fi
  echo "import-openapi: cloning 7K-Inari/inari-server@$REF"
  git clone --quiet --depth 1 --branch "$REF" https://github.com/7K-Inari/inari-server.git "$TMP/inari-server"
  (cd "$TMP/inari-server" && go run ./cmd/export-openapi "$TMP/exported.yaml")
  EXPORTED="$TMP/exported.yaml"
fi

[[ -f "$EXPORTED" ]] || { echo "import-openapi: exported spec not found: $EXPORTED" >&2; exit 1; }

# Merge: fragment provides info/servers + /healthz + /readyz + HealthStatus;
# the export provides the full path/schema surface. Export wins on collisions
# except for the keys the fragment owns (health paths are not in the export).
# Patch: drop path parameters that do not appear in the path template
# (Huma emits a stray `id` path param on /api/v1/tenants/{org}/git-config GET;
# remove this once fixed in inari-server).
yq eval-all '
  select(fileIndex == 0) as $frag |
  select(fileIndex == 1) as $exp |
  $exp
  | .info = $frag.info
  | .servers = $frag.servers
  | .paths = ($frag.paths + $exp.paths)
  | .components.schemas = ($frag.components.schemas + $exp.components.schemas)
' "$FRAGMENT" "$EXPORTED" > "$OUT.tmp"

yq eval -i '
  .paths."/api/v1/tenants/{org}/git-config".get.parameters
    |= map(select(.name != "id"))
' "$OUT.tmp"

# Downgrade OpenAPI 3.1 -> 3.0.3 for the codegen toolchain (oapi-codegen and
# orval do not fully support 3.1): rewrite `type: [T, "null"]` unions as
# `type: T` + `nullable: true`.
yq eval -i '
  .openapi = "3.0.3" |
  with((.components.schemas | ..) | select(has("examples"));
    .example = .examples[0] | del(.examples)
  ) |
  with((..) | select(.contentEncoding? == "base64");
    .format = "byte" | del(.contentEncoding)
  ) |
  with(.. | select((.type? | tag == "!!seq") and (.type | contains(["null"])));
    .type = (.type | map(select(. != "null")) | .[0]) |
    .nullable = true
  )
' "$OUT.tmp"

# Health endpoints live at the root (outside the /api/v1 prefix of the
# exported operations), so drop the global servers prefix ambiguity: keep
# servers from the fragment but annotate health ops as absolute.
mv "$OUT.tmp" "$OUT"
echo "import-openapi: wrote $OUT"
