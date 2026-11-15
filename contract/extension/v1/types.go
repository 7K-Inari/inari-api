// Package extensionv1 defines the versioned backend and UI extension contract
// declarations shared by Inari servers, SDKs, and clients.
package extensionv1

import "fmt"

const (
	BackendContractV1 = "inari.plugin.v1"
	UIHostAPIV1       = "inari.ui.v1"
)

type ContractErrorCode string

const (
	CodeContractRequired    ContractErrorCode = "contract_required"
	CodeContractMalformed   ContractErrorCode = "contract_malformed"
	CodeContractUnsupported ContractErrorCode = "contract_unsupported"
	CodeContractMismatch    ContractErrorCode = "declaration_mismatch"
)

// ContractError identifies a machine-readable contract validation failure.
type ContractError struct {
	Code ContractErrorCode
	Path string
	Err  error
}

func (e *ContractError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("extension contract: %s: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("extension contract: %s: %s: %v", e.Code, e.Path, e.Err)
}

func (e *ContractError) Unwrap() error { return e.Err }

func NewContractError(code ContractErrorCode, path string, format string, args ...any) error {
	return &ContractError{Code: code, Path: path, Err: fmt.Errorf(format, args...)}
}

// Contracts is the spec.contracts extension manifest declaration.
type Contracts struct {
	Backend []string    `json:"backend,omitempty" yaml:"backend,omitempty"`
	UI      *UIContract `json:"ui,omitempty" yaml:"ui,omitempty"`
}

// UIContract declares the UI host and SDK compatibility surface.
type UIContract struct {
	SDK     string `json:"sdk" yaml:"sdk"`
	HostAPI string `json:"hostApi" yaml:"hostApi"`
}

// ContractMetadata is GET /api/v1/meta/contracts response body.
type ContractMetadata struct {
	PluginContracts []string `json:"pluginContracts"`
	UIHostAPIs      []string `json:"uiHostApis"`
	ServerVersion   string   `json:"serverVersion"`
}

// ExtensionDegradedReason classifies runtime health failures.
type ExtensionDegradedReason string

const (
	DegradedContractUnsupported ExtensionDegradedReason = "contract_unsupported"
	DegradedHandshakeFailed     ExtensionDegradedReason = "handshake_failed"
	DegradedIdentityMismatch    ExtensionDegradedReason = "identity_mismatch"
	DegradedChecksumMismatch    ExtensionDegradedReason = "checksum_mismatch"
	DegradedProcessExited       ExtensionDegradedReason = "process_exited"
)
