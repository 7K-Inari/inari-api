# Generic provisioning contract

`inari.provisioning.v1.ProvisioningService` is a unary Connect RPC contract for
machine-to-machine tenant provisioning. The generated Go and TypeScript clients
are committed under `gen/`. Connect JSON permits HTTP clients to use the same
versioned contract without a second hand-maintained REST schema.

The service is deliberately not part of `openapi/openapi.yaml`: that document
describes the public inari-server REST/console surface and does not imply that
an OSS server mounts this contract.

## Procedures

| Operation | Procedure |
| --- | --- |
| Create tenant | `/inari.provisioning.v1.ProvisioningService/CreateTenant` |
| Get tenant | `/inari.provisioning.v1.ProvisioningService/GetTenant` |
| Put plan limits | `/inari.provisioning.v1.ProvisioningService/PutPlanLimits` |
| Resend first-admin invite | `/inari.provisioning.v1.ProvisioningService/ResendFirstAdminInvite` |

`CreateTenantRequest` carries `slug`, optional `display_name`, `admin_email`,
and an opaque `plan`. `PutPlanLimitsRequest` carries `tenant_id`, `plan`, and a
complete desired `inari.planlimits.v1.PlanLimits` document.

## Identity metadata

| Metadata | Contract |
| --- | --- |
| `Authorization: Bearer <jwt>` | Short-lived client-credentials token issued for a dedicated machine identity. |
| OAuth scope | `provisioning` is required. No user or tenant-claim context is accepted. |
| `Idempotency-Key` | Required on every mutation; 1-255 visible ASCII bytes. |

Implementations should bind authorization to an explicit machine client and
service-account subject, not only to claims that can also appear in user tokens.
Tokens, credentials, and actionable invite links must not be logged or returned.

## Idempotency and replay

An idempotency record is namespaced by authenticated machine subject, RPC, and
key. The same key and semantically equivalent request returns the stored result
without repeating completed external mutations. The same key and different
request returns `ALREADY_EXISTS`. An incomplete matching operation resumes;
an actively leased concurrent operation returns `UNAVAILABLE` and may be retried.

Implementations persist an audit intent before external side effects and make
each side effect independently recoverable. Numeric limit presence survives
serialization: absent means unlimited, while present zero means zero allowed.
`all_features=true` enables current and future gates; otherwise gates are enabled
only by an explicit true map entry.

## First-admin flow

The implementation ensures the admin user and required organization/team
memberships before sending an identity-provider action email. New users normally
receive email verification and password setup; existing verified users do not
receive a forced password reset. Responses expose only an invite status, never
an action URL. An uncertain send becomes `ACTION_REQUIRED`; an explicitly keyed
resend creates a new invite generation and audit record.

## Errors and compatibility

| Condition | Connect code |
| --- | --- |
| Missing or invalid machine identity | `UNAUTHENTICATED` |
| Wrong client/subject or missing scope | `PERMISSION_DENIED` |
| Rate limit exceeded | `RESOURCE_EXHAUSTED` |
| Malformed request or limits | `INVALID_ARGUMENT` |
| Unknown tenant | `NOT_FOUND` |
| Idempotency-key or natural-key conflict | `ALREADY_EXISTS` |
| Non-retryable lifecycle/invite state | `FAILED_PRECONDITION` |
| Retryable upstream or active lease | `UNAVAILABLE` |

Changes must remain additive within `v1`. Renumbering fields, changing message
or RPC types, or renaming packages requires a new package version. `buf lint`
and `buf breaking` enforce the source compatibility rules.
