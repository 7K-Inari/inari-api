# Verification contracts

`contract/verification/v1` is the dependency-free canonical identity contract shared by server and agent. It defines application revision, trigger/run identity, canonical JSON digests, and Argo CronWorkflow child-name reconstruction. Golden vectors prevent cross-binary drift.

`contract/trustedsource/v1` carries additive runtime evidence eligibility: `targetAdapter`, pointer `fixtureOnly`, `executableProfiles`, `evidenceScope`, and pointer `appEvidenceEligible`. All five fields must be present together for new evidence. Legacy absence remains decodable but is not real-environment eligibility. Explicit fixture/application-ineligible evidence returns stable `suite_ci_only`; missing/partial metadata returns `suite_evidence_eligibility_unknown`; a non-executable requested profile returns `suite_profile_not_executable`.

The additive agent contract uses `EVENT_TYPE_VERIFICATION_REPORT=16` and `inari.agent.verification-report.v1` on the existing `EventStream`. `StatusUpdate` adds observed source revision and sync operation identity. `VerificationReport` binds run/attempt, Workflow UID/resourceVersion, app/suite/target/companion/pack/capability digests, policy/grant revisions, timestamps, and bounded outcome JSON. Org and cluster identity always come from the authenticated stream session. Unknown enum/event/payload handling remains N/N−1 tolerant; there is no public result-ingest REST operation.

REST request/response models originate from the inari-server Huma registry. `openapi/openapi.yaml`, Go OAS types, and TypeScript clients are generated artifacts; do not hand-edit them. Regenerate with the server export followed by `make import-openapi` and `make generate`, then run buf breaking/lint, Go tests, and TypeScript checks.
