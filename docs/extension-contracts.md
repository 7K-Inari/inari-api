# Extension contract compatibility

Inari extensions declare their implemented backend and UI contracts in `extension.yaml`; runtime discovery is a safety check, not a substitute for declaration.

```yaml
spec:
  contracts:
    backend: ["inari.plugin.v1"]
    ui:
      sdk: ">=0.1.8 <2.0.0"
      hostApi: "inari.ui.v1"
```

`backend` is a set, not a preference list. Registration is valid when at least one declared backend contract intersects the server-supported set. The handshake independently verifies the actual `PluginInfo.api_version`; an unsupported runtime selection fails closed with `contract_unsupported`. UI `hostApi` names the host surface and `sdk` is the npm SemVer range compiled against. The browser checks host API membership and SDK range intersection before exposing slots.

Backend REST registration carries the JSON form of `spec` in `manifest`, so declarations live at `manifest.contracts`. UI registration accepts the same object as its optional `contracts` body field. Extension responses include `contracts` when explicitly declared. Missing metadata is distinct from an empty or malformed declaration.

## Deprecation and skew

During the initial optional phase, legacy manifests without contracts remain accepted and produce deprecation warnings. `INARI_EXTENSION_CONTRACTS_REQUIRED=false` is the initial default. After a published minimum window of 90 days and two stable server releases, a later server release changes the default to required. Missing declarations then fail registration with 422 `contract_required`, and UI remotes without stamped metadata become incompatible rather than mounting.

A future `inari.plugin.v2` follows the same published overlap policy: servers advertise and accept v1+v2 during the overlap, test v1-only, v2-only, and mixed peers, and remove v1 only after the window. Actual v2 support must never be inferred from a declaration alone.

Within v1, protobuf and OpenAPI changes are additive. Consumers ignore unknown additive fields. Buf breaking checks remain mandatory; field removal, renumbering, type changes, and package or service renames require a new major contract.

## Machine-readable support metadata

`GET /api/v1/meta/contracts` is tenant-free and unauthenticated. It returns only non-secret support metadata:

```json
{
  "pluginContracts": ["inari.plugin.v1"],
  "uiHostApis": ["inari.ui.v1"],
  "serverVersion": "4.1.0"
}
```

Arrays are non-empty, unique, and sorted. `serverVersion` identifies the build but is not an SDK compatibility oracle. CLI and CI clients use bounded timeouts and ignore unknown additive response fields.

## Compatibility evidence

The immutable consumer input for reusable compatibility workflows is `7K-Hiroba/workflows@f572c5a244e27e4a77589f2730066749c7b51c6b` (v1.0.0). Its PR lanes run pinned-SDK testkit checks against stable and previous-stable entries and browser mount/slot tests against stable UI. Nightly reruns both against resolved server/UI edge SHAs and opens or updates one early-warning issue on failure.

Every declared UI slot requires a passing Playwright spec named `slot <kind>:<name>`. Reviewed UI compatibility evidence is source-bound and no more than seven days old, with passing stable/previous-stable matrix entries, browser result, and declared slot identifiers. Missing, stale, future-dated, incomplete, or failed evidence fails closed. Comprehensive application E2E and visual regression remain extension-maintainer responsibilities.
