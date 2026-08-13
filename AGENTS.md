# inari-api — Agent Guide

The compatibility contract for Inari: protobuf definitions (agent EventStream protocol, plugin contract) + OpenAPI specs, with generated Go/TS clients (plan §6 #5).

Stack: Buf, protoc (connect/gRPC), oapi-codegen; generated Go + TypeScript clients

## Key architecture constraints
- This repo is the **cross-repo contract**: server, agent, cli, and plugins pin its versioned packages; CI runs contract/breaking-change tests (buf breaking) (§6, §10).
- Agent protocol: bidirectional gRPC stream, CloudEvents-style envelope (eventid/resourceid, type, payload) (§4.3).
- Plugin contract: versioned gRPC (hashicorp go-plugin model) (§5.8).
- Conventional Commits + SemVer tags; breaking proto changes require a new package version (v1 → v2), never in-place edits.

## Conventions
- Conventional Commits; SemVer releases; container images/artifacts cosign-signed (once CI exists).
- Write tests for new behavior; keep changes minimal and focused.
- Canonical architecture & development plan: https://github.com/7K-Inari/inari-docs/blob/main/docs/architecture/inari-platform-plan.md (section references below point into it).

## Platform design principles (apply everywhere)
1. Tenant-aware to the core — every object carries a tenant ID; every API decision is tenant-scoped.
2. Zero tenant credentials on the hub — no tenant kubeconfigs or cloud keys in the control plane.
3. Pull, never push — agents dial out; the control plane never initiates connections into tenant networks.
4. Desired state, eventually reconciled — GitOps/CR-based mutations, not imperative RPCs.
5. The catalog is a projection of reality — capabilities are discovered, not declared.
6. Small kernel, everything else extension.
7. Modular monolith first — strict internal module boundaries.
