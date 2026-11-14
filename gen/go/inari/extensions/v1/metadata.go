package extensionsv1

// AgentGateway request metadata keys. These are hand-maintained companions to
// the generated contract, like agent/v1/event_type_strings.go; buf does not
// overwrite this file.
const (
	// MetadataTenant carries the authenticated organization ID. Extension
	// clients derive it from the invocation AuthContext, never action payload.
	MetadataTenant = "x-inari-tenant"
	// MetadataCluster identifies the target tenant cluster.
	MetadataCluster = "x-inari-cluster"
	// MetadataUserCredential carries a raw downstream user credential only on
	// the in-memory extension-to-control-plane hop. It is never persisted.
	MetadataUserCredential = "x-inari-user-credential"
	// MetadataUserCredentialRef carries an opaque control-plane credential
	// reference. It is mutually exclusive with MetadataUserCredential and must
	// never contain credential material.
	MetadataUserCredentialRef = "x-inari-user-credential-ref"
)
