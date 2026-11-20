// Package verification defines the dependency-free contract shared by the
// control plane, agents, and clients for environment verification.
package verification

// TriggerKind identifies the source of a verification run.
type TriggerKind string

const (
	TriggerOnRevisionChange TriggerKind = "onRevisionChange"
	TriggerOnSync           TriggerKind = "onSync"
	TriggerManual           TriggerKind = "manual"
	TriggerCron             TriggerKind = "cron"
)

// Trigger is the canonical execution-trigger snapshot used for deduplication.
// Profile is execution-affecting and therefore belongs to this snapshot.
type Trigger struct {
	Kind          TriggerKind `json:"kind"`
	Profile       string      `json:"profile"`
	OperationID   string      `json:"operationId,omitempty"`
	RequestKey    string      `json:"requestKey,omitempty"`
	ScheduleSetID string      `json:"scheduleSetId,omitempty"`
	ScheduledAt   string      `json:"scheduledAt,omitempty"`
}

// AppRevisionInput identifies one exact desired application revision.
type AppRevisionInput struct {
	InstanceUID    string `json:"instanceUid"`
	ItemID         string `json:"itemId"`
	Version        string `json:"version"`
	ArtifactDigest string `json:"artifactDigest"`
	SpecDigest     string `json:"specDigest"`
	Generation     int    `json:"generation"`
	DesiredCommit  string `json:"desiredCommit"`
}

// RunIdentity is derived from immutable run inputs.
type RunIdentity struct {
	AppRevision        string `json:"appRevision"`
	RunKey             string `json:"runKey"`
	RunID              string `json:"runId"`
	TriggerFingerprint string `json:"triggerFingerprint"`
}
