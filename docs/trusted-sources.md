# Trusted source and artifact admission contract v1

`contract/trustedsource/v1` is the canonical Go and JSON Schema contract shared by the trusted-publishing producer and the inari-server admission consumer. It covers source configuration, `CatalogIndex`, `TemplateIndex`, SLSA provenance, and signed test evidence. It does not change any protobuf package.

## Source identity and immutable artifacts

- A server-assigned source ID and package name form the only external identity: `<source>/<package>`. Index documents cannot declare or replace that source ID.
- `catalog.sources` and `template.sources` each bind the source ID to an OCI `indexRef`, trust tier, artifact repository prefixes, and exact publisher policy.
- Every deployable artifact is resolved as `repository@sha256:<64-hex-digest>`. `latest` is permitted only on the configured sync index and is never a deploy-time artifact reference. `channelRef` is not part of v1.
- A `(source, package, version)` cannot change digest. Channels are metadata within one source; they never select content from another source.

## Publisher identity

The source configuration distinguishes two repositories and two SHAs:

| Field | Meaning |
| --- | --- |
| `publisher.repository` / `workflowSha` | Reusable signing repository and its resolved workflow-library commit SHA. |
| `publisher.sourceRepository` / index `sourceSha` | Application source repository and the commit being released. |
| `publisher.sourceWorkflow` | Top-level application release workflow. |
| `sourceRef`, `event`, `releaseEnvironment` | Optional exact constraints on protected release claims. Configuring one makes the corresponding evidence mandatory where supported. |

The per-artifact workflow map is mandatory; there is no shared fallback:

| Role | Reusable workflow |
| --- | --- |
| `index` | `.github/workflows/release.yml` |
| `catalogPackage`, `template` | `.github/workflows/release-oci.yml` |
| `helmChart` | `.github/workflows/release-chart.yml` |
| `smokeSuite` | `.github/workflows/release-smoke-suite.yml` |

A consumer verifies the signature first, then requires the verified certificate and evidence to agree on issuer, signing repository/workflow/SHA, source repository/SHA/workflow, selected event/ref/environment, and artifact digest. Repository-only regular expressions are not sufficient. `BuildSignerURI/Digest`, `BuildConfigURI/Digest`, and `SourceRepositoryURI/Digest` have distinct meanings; the real producer certificate fixture is the compatibility authority. A claim mismatch is a hard stop, not a reason for a fallback.

## Index contract

Indexes carry `apiVersion: inari.dev/v1`, their kind and index media type, top-level source SHA/ref and producer run metadata, and all currently selectable entries. The index OCI tag includes the immutable index source SHA before the sync `latest` tag is applied.

Every entry includes its package/deploy version, package release version, channel, type, OCI ref and manifest digest, source SHA, anticipated or repaired release tag, artifact media type, required lane evidence, and optional compatibility/runtime evidence. A `platform-app` also has an `oci://` chart ref and chart digest. An initial release can name an anticipated tag because tag finalization occurs only after all artifacts and indexes verify.

The producer publishes an index only after every current entry passes its required signature, SLSA, and test-evidence checks. A consumer validates every entry again rather than accepting a subset.

## SLSA and test evidence

SLSA uses predicate type `https://slsa.dev/provenance/v1`, an in-toto Statement v1 subject bound to the exact OCI repository and digest, immutable workflow dependencies, and matching builder/invocation metadata. Provenance is attached to the manifest digest and must be available as an OCI registry referrer. An unsigned provenance file is not a substitute.

Signed tests use `https://inari.dev/attestations/test-evidence/v1` and `test-evidence-v1.schema.json`. Required data includes app/component/tier, source SHA/release tag, artifact repository/digest/media type, producer workflow/run/attempt, required lanes and their run/job IDs, compatibility/runtime results where applicable, finish time, and a digest of redacted lane evidence. Every required lane must have passed against the artifact source SHA.

General release evidence remains bound to its immutable release. Reviewed UI compatibility evidence additionally requires passing stable and previous-stable backend results and passing browser results for every declared slot. Its completion time must be no more than seven days old at release and at admission (unless a shorter positive `pluginCompatibilityMaxAge` is configured). Future timestamps beyond the allowed clock skew are invalid.

## Failure and compatibility policy

Official and reviewed artifacts require signature, exact publisher, digest, SLSA provenance, and signed passing tests at deploy/render. Signed experimental artifacts do not require the two predicates. `allowUnsigned: development-only` is valid only for experimental sources and never bypasses digest, schema, media type, namespace, or non-production policy. There is no production unsigned fallback, old-ID alias, mutable channel ref, or legacy environment-variable compatibility layer.
