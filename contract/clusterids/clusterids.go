// Package clusterids pins the canonical cluster-ID charset contract:
// the shapes inari-server issues and every consumer (inari-cli, UI,
// docs) must accept — plus shapes that must stay rejected. It lives in
// the contract repo so consumers validate against one shared source of
// truth instead of drifting private regexes.
//
// IDs must stay embeddable unquoted in YAML names and URL paths (the
// kubeconfig renderer and proxy routes carry them), which is why the
// charset excludes whitespace, slashes, and percent signs.
package clusterids

// ValidIDs are cluster-ID shapes consumers MUST accept. The canonical
// server-issued shape is "cluster:<uuid>" (inari-server
// clusterregistry.CreateCluster). The remaining entries are legacy and
// test-fixture shapes that appear across existing state, docs, and
// scripts and must keep working.
var ValidIDs = []string{
	// canonical server-issued shape
	"cluster:3f8a9c2e-1b4d-4e5f-9a6b-7c8d9e0f1a2b",
	"cluster:00000000-0000-4000-8000-000000000000",
	// legacy / fixture shapes in existing state and docs
	"cluster-1",
	"prod-eu-1",
	"c1",
	"9lives",
	"clu-1",
	"kind.dev_local",
}

// InvalidIDs are shapes consumers MUST reject.
var InvalidIDs = []string{
	"",
	":cluster:x",  // leading separator
	"-prod",       // leading dash
	".prod",       // leading dot
	"cl uster",    // whitespace
	"cluster/1",   // path separator
	"cluster%201", // percent encoding
	"cluster?x=1", // query char
}
