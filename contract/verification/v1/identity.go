package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	digestPattern  = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	uuidPattern    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	profilePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)
	shaPattern     = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
)

// DigestJSON returns a digest over canonical JSON. Object keys are sorted,
// insignificant whitespace is removed, and JSON numbers retain their literal
// precision through decoding.
func DigestJSON(raw []byte) (string, error) {
	canonical, err := canonicalFromRaw(raw)
	if err != nil {
		return "", err
	}
	return shaDigest(canonical), nil
}

// ComputeAppRevision binds an exact desired application revision.
func ComputeAppRevision(in AppRevisionInput) (string, error) {
	if !uuidPattern.MatchString(in.InstanceUID) {
		return "", errors.New("instanceUid must be a UUID")
	}
	if strings.TrimSpace(in.ItemID) == "" || strings.TrimSpace(in.Version) == "" {
		return "", errors.New("itemId and version are required")
	}
	if !digestPattern.MatchString(in.ArtifactDigest) || !digestPattern.MatchString(in.SpecDigest) {
		return "", errors.New("artifactDigest and specDigest must be sha256 digests")
	}
	if in.Generation < 1 {
		return "", errors.New("generation must be positive")
	}
	if in.DesiredCommit != "" && !shaPattern.MatchString(in.DesiredCommit) {
		return "", errors.New("desiredCommit must be empty or a full Git SHA")
	}
	return digestCanonical(in)
}

// ComputeRunIdentity derives deterministic dedupe and public run identities.
func ComputeRunIdentity(instanceUID, appRevision, suiteDigest string, trigger Trigger) (RunIdentity, error) {
	if !uuidPattern.MatchString(instanceUID) {
		return RunIdentity{}, errors.New("instanceUid must be a UUID")
	}
	if !digestPattern.MatchString(appRevision) {
		return RunIdentity{}, errors.New("appRevision must be a sha256 digest")
	}
	if !digestPattern.MatchString(suiteDigest) {
		return RunIdentity{}, errors.New("suiteDigest must be a sha256 digest")
	}
	normalized, err := normalizeTrigger(trigger)
	if err != nil {
		return RunIdentity{}, err
	}
	triggerJSON, err := canonicalJSON(normalized)
	if err != nil {
		return RunIdentity{}, err
	}
	keyInput := struct {
		AppRevision string  `json:"appRevision"`
		InstanceUID string  `json:"instanceUid"`
		SuiteDigest string  `json:"suiteDigest"`
		Trigger     Trigger `json:"trigger"`
	}{appRevision, instanceUID, suiteDigest, normalized}
	keyJSON, err := canonicalJSON(keyInput)
	if err != nil {
		return RunIdentity{}, err
	}
	runKey := shaDigest(keyJSON)
	rawKey, err := hex.DecodeString(strings.TrimPrefix(runKey, "sha256:"))
	if err != nil {
		return RunIdentity{}, err
	}
	id := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(rawKey[:20])
	return RunIdentity{
		AppRevision:        appRevision,
		RunKey:             runKey,
		RunID:              "vr-" + strings.ToLower(id),
		TriggerFingerprint: shaDigest(triggerJSON),
	}, nil
}

// CronTriggerFromChildName reconstructs the scheduled trigger from Argo's
// deterministic <cron-name>-<scheduled-unix-time> child name.
func CronTriggerFromChildName(cronName, childName, profile, scheduleSetID string) (Trigger, error) {
	if strings.TrimSpace(cronName) == "" || strings.TrimSpace(scheduleSetID) == "" {
		return Trigger{}, errors.New("cron name and scheduleSetId are required")
	}
	prefix := cronName + "-"
	if !strings.HasPrefix(childName, prefix) {
		return Trigger{}, errors.New("Workflow does not belong to the expected CronWorkflow")
	}
	suffix := strings.TrimPrefix(childName, prefix)
	if suffix == "" || strings.Contains(suffix, "-") {
		return Trigger{}, errors.New("invalid cron child name")
	}
	for _, r := range suffix {
		if r < '0' || r > '9' {
			return Trigger{}, errors.New("invalid cron scheduled time")
		}
	}
	unix, err := strconv.ParseInt(suffix, 10, 64)
	if err != nil || unix <= 0 {
		return Trigger{}, errors.New("invalid cron scheduled time")
	}
	scheduled := time.Unix(unix, 0).UTC()
	if scheduled.Year() > 9999 {
		return Trigger{}, errors.New("cron scheduled time is out of range")
	}
	return normalizeTrigger(Trigger{
		Kind: TriggerCron, Profile: profile, ScheduleSetID: scheduleSetID,
		ScheduledAt: scheduled.Format(time.RFC3339),
	})
}

func normalizeTrigger(in Trigger) (Trigger, error) {
	if !profilePattern.MatchString(in.Profile) {
		return Trigger{}, errors.New("profile is required and must be a bounded identifier")
	}
	if len(in.OperationID) > 256 || len(in.RequestKey) > 256 || len(in.ScheduleSetID) > 128 {
		return Trigger{}, errors.New("trigger identifier exceeds its size limit")
	}
	switch in.Kind {
	case TriggerOnRevisionChange:
		if in.OperationID != "" || in.RequestKey != "" || in.ScheduleSetID != "" || in.ScheduledAt != "" {
			return Trigger{}, errors.New("onRevisionChange accepts only kind and profile")
		}
	case TriggerOnSync:
		if strings.TrimSpace(in.OperationID) == "" || in.RequestKey != "" || in.ScheduleSetID != "" || in.ScheduledAt != "" {
			return Trigger{}, errors.New("onSync requires only operationId")
		}
	case TriggerManual:
		if strings.TrimSpace(in.RequestKey) == "" || in.OperationID != "" || in.ScheduleSetID != "" || in.ScheduledAt != "" {
			return Trigger{}, errors.New("manual requires only requestKey")
		}
	case TriggerCron:
		if strings.TrimSpace(in.ScheduleSetID) == "" || in.OperationID != "" || in.RequestKey != "" || in.ScheduledAt == "" {
			return Trigger{}, errors.New("cron requires scheduleSetId and scheduledAt")
		}
		parsed, err := time.Parse(time.RFC3339, in.ScheduledAt)
		if err != nil || parsed.Unix() <= 0 {
			return Trigger{}, errors.New("scheduledAt must be a valid RFC3339 timestamp")
		}
		in.ScheduledAt = parsed.UTC().Format(time.RFC3339)
	default:
		return Trigger{}, fmt.Errorf("unsupported trigger kind %q", in.Kind)
	}
	return in, nil
}

func digestCanonical(value any) (string, error) {
	raw, err := canonicalJSON(value)
	if err != nil {
		return "", err
	}
	return shaDigest(raw), nil
}

func canonicalJSON(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicalFromRaw(raw)
}

func canonicalFromRaw(raw []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode canonical JSON: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("canonical JSON contains multiple values")
		}
		return nil, fmt.Errorf("decode trailing JSON: %w", err)
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode canonical JSON: %w", err)
	}
	return canonical, nil
}

func shaDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
