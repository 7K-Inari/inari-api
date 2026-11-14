# Extension Agent Gateway contract

`inari.extensions.v1.AgentGateway` is the versioned control-plane contract for an authenticated backend extension to request one imperative operation on a tenant cluster. It reuses `inari.agent.v1.InvokeAction` and `CommandAck`; it does not change the agent EventStream protocol.

**Procedure:** `/inari.extensions.v1.AgentGateway/InvokeAction`

The addition is additive within `inari.extensions.v1`. New fields or RPCs may be added compatibly. Renaming the package, service, RPC, or request/response types requires `inari.extensions.v2`; `buf breaking` gates changes.

## Consumers

Backend extension authors use the typed `host/agentgateway` client in `inari-plugin-sdk`. Do not create a service descriptor, custom protobuf JSON codec, raw gRPC connection, or private copy of this contract. The generated Go Connect client/handler and TypeScript definitions live under `gen/` for platform integrations.

## Identity and routing metadata

| Metadata | Contract |
| --- | --- |
| `Authorization: Bearer <extension-jwt>` | Required short-lived `client_credentials` JWT for the extension's dedicated identity. It is never the end user's token or a shared gateway token. |
| `x-inari-tenant` | Required authenticated organization ID. The SDK derives it from `AuthContext`, never from an action payload. |
| `x-inari-cluster` | Required target cluster. The control plane verifies cluster ownership and active state. |
| `x-inari-user-credential` | Optional downstream user credential on the in-memory extension-to-control-plane hop only. |
| `x-inari-user-credential-ref` | Optional opaque control-plane vault reference. It is mutually exclusive with the raw credential and is not a JWT. |

The typed SDK obtains the tenant and optional downstream credential from the current plugin invocation context. Extension code must not accept a tenant ID or credential reference from caller input.

## Command and credential lifecycle

1. The extension proxy has already authenticated and authorized the end user. The gateway separately authenticates the extension machine identity and checks its organization binding.
2. The control plane stamps a fresh `command_id`, owns the command journal, and discards any self-asserted `InvokeAction.user_credential_ref`.
3. When a raw downstream credential is present, the control plane mints a short-lived, single-use vault reference bound to the command and cluster. Only that reference enters the persisted command. Raw credential material is never logged, audited, persisted, or sent to the agent.
4. The agent receives the command over its existing outbound EventStream. Its normal command ID deduplication and acknowledgement semantics apply. The typed client does not automatically retry an imperative operation.
5. A failed `CommandAck` is a command result, not a transport error. Callers inspect `ack.result` before reporting success.

## Failure contract

| Condition | Connect code |
| --- | --- |
| Missing/invalid extension identity | `UNAUTHENTICATED` |
| Extension invokes outside its organization or cluster org mismatches | `PERMISSION_DENIED` |
| Missing tenant/cluster metadata, malformed credential reference, or mutually exclusive credential inputs | `INVALID_ARGUMENT` |
| Unknown cluster | `NOT_FOUND` |
| Extension or cluster is not in the required state | `FAILED_PRECONDITION` |
| No live agent before enqueue, or the agent disconnects while waiting | `UNAVAILABLE` |
| Connected agent does not acknowledge before the invocation deadline | `DEADLINE_EXCEEDED` |
| Caller cancels the invocation | `CANCELED` |

The disconnected check occurs before command enqueue and audit, so an offline agent cannot cause a delayed mutation. Other control-plane command queues are unchanged.
