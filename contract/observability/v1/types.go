// Package observability defines the signed producer metadata and target-side
// verification contracts shared by Inari deployment consumers.
package observability

import "time"

const (
	APIVersion           = "inari.dev/v1alpha1"
	CapabilityKind       = "Observability"
	VerificationKind     = "ObservabilityVerificationResult"
	PackRenderContractV2 = "inari.observability-pack/v2"
	ProfileCI            = "ci"
	ProfileStage         = "stage"
	ProfileProd          = "prod"
)

type Evaluation string

const (
	EvaluationCentral Evaluation = "central"
	EvaluationLocal   Evaluation = "local"
)

type Dashboard struct {
	Title string `json:"title" yaml:"title"`
	UID   string `json:"uid" yaml:"uid"`
	File  string `json:"file" yaml:"file"`
}

type Rule struct {
	File       string     `json:"file" yaml:"file"`
	Evaluation Evaluation `json:"evaluation" yaml:"evaluation"`
}

type PackReference struct {
	Package      string `json:"package" yaml:"package"`
	VersionRange string `json:"versionRange" yaml:"versionRange"`
}

type ObservabilityCapability struct {
	APIVersion        string            `json:"apiVersion" yaml:"apiVersion"`
	Kind              string            `json:"kind" yaml:"kind"`
	App               string            `json:"app" yaml:"app"`
	Folder            string            `json:"folder" yaml:"folder"`
	Datasources       map[string]string `json:"datasources" yaml:"datasources"`
	LinkVariables     []string          `json:"linkVariables,omitempty" yaml:"linkVariables,omitempty"`
	Dashboards        []Dashboard       `json:"dashboards" yaml:"dashboards"`
	Rules             []Rule            `json:"rules" yaml:"rules"`
	Targets           []Evaluation      `json:"targets" yaml:"targets"`
	SeverityRouting   map[string]string `json:"severityRouting" yaml:"severityRouting"`
	MinimumAppVersion string            `json:"minimumAppVersion" yaml:"minimumAppVersion"`
	ObservabilityPack PackReference     `json:"observabilityPack" yaml:"observabilityPack"`
}

// Capability is retained as a Go source alias; the canonical schema name is
// ObservabilityCapability to avoid colliding with inventory capabilities.
type Capability = ObservabilityCapability

type CompanionIdentity struct {
	ID         string `json:"id"`
	Generation int64  `json:"generation"`
}

type TargetIdentity struct {
	ID        string `json:"id"`
	Revision  int64  `json:"revision"`
	ClusterID string `json:"clusterId"`
}

type AppIdentity struct {
	PlanID        string `json:"planId"`
	InstanceID    string `json:"instanceId,omitempty"`
	ScaffoldRunID string `json:"scaffoldRunId,omitempty"`
	SourceDigest  string `json:"sourceDigest"`
	Revision      string `json:"revision"`
}

type PackIdentity struct {
	ItemID  string `json:"itemId"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type SuiteIdentity struct {
	Digest  string `json:"digest"`
	Profile string `json:"profile"`
}

type RunIdentity struct {
	ID         string    `json:"id"`
	Attempt    int       `json:"attempt"`
	FinishedAt time.Time `json:"finishedAt"`
}

type Check struct {
	State string `json:"state"`
}

type FolderCheck struct {
	Check
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type DashboardCheck struct {
	Check
	UID   string `json:"uid"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

type DatasourceCheck struct {
	Check
	Logical string `json:"logical"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}

type RuleCheck struct {
	Check
	File       string     `json:"file"`
	Evaluation Evaluation `json:"evaluation"`
}

type VerificationChecks struct {
	Folder      FolderCheck       `json:"folder"`
	Dashboards  []DashboardCheck  `json:"dashboards"`
	Datasources []DatasourceCheck `json:"datasources"`
	Rules       []RuleCheck       `json:"rules"`
}

type VerificationResult struct {
	APIVersion       string             `json:"apiVersion" yaml:"apiVersion"`
	Kind             string             `json:"kind" yaml:"kind"`
	Companion        CompanionIdentity  `json:"companion" yaml:"companion"`
	Target           TargetIdentity     `json:"target" yaml:"target"`
	App              AppIdentity        `json:"app" yaml:"app"`
	Pack             PackIdentity       `json:"pack" yaml:"pack"`
	CapabilityDigest string             `json:"capabilityDigest" yaml:"capabilityDigest"`
	Suite            SuiteIdentity      `json:"suite" yaml:"suite"`
	Run              RunIdentity        `json:"run" yaml:"run"`
	Result           string             `json:"result" yaml:"result"`
	Checks           VerificationChecks `json:"checks" yaml:"checks"`
	ExpiresAt        time.Time          `json:"expiresAt" yaml:"expiresAt"`
}
