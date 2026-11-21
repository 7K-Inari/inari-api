# Deployment-time observability contracts

`contract/observability/v1` is the shared producer and target-verification contract for Inari DeploymentPlans.

## Capability metadata

Catalog packages, direct Helm charts (`inari/observability.yaml`), and deployment-capable component templates carry the same signed `Observability` document. It declares deterministic dashboard UIDs, rule files and central/local evaluation, supported targets, one primary datasource under the current render contract, explicit link variables, an app SemVer range, and a same-source observability-pack SemVer range.

Versions are resolved to immutable OCI digests before submission. `latest`, mutable tags, and channel resolution are not valid submit-time behavior. Template compatibility uses `TemplateManifest.Version`, not a future generated application version.

The companion pack must support `inari.observability-pack/v2`, which makes `rulesCentral` and `rulesLocal` effective alongside dashboard rendering. Earlier packs that cannot express dashboards plus one rule evaluation are incompatible.

## Target-side verification

`ObservabilityVerificationResult` binds the companion generation, target revision/cluster, application source revision/digest, exact pack, capability digest, approved read-only suite/profile, and run. The suite profile is the actual generic verifier grant profile: `ci`, `stage`, or `prod`, admitted for the exact target environment; no synthetic observability-specific profile is accepted. Passing readiness requires folder, dashboard, datasource, and selected-rule checks. It makes no Grafana identity, team-sync, or folder-access claim. Evidence carries UIDs and other identifiers, never arbitrary URLs or credentials.

The generic verification transport authenticates and submits this result over the existing outbound agent path. Inari does not call Grafana from the hub. Deep links are derived only from current, unexpired passing evidence and the current caller's target view grant; skipped, pending, failed, stale, or revoked states produce no links.
