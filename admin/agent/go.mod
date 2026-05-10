module github.com/jcsvwinston/nucleus/admin/agent

go 1.26.3

// Real require blocks are added by Phase 3 when the agent loop is implemented.
// In Phase 1 the module only declares its identity and ties to go.work.

// Local module wiring is provided by go.work at the repository root.
// These replace directives are present so that tools operating on this module
// in isolation (e.g. `go list -m all` from inside admin/agent/) still resolve
// the local sources correctly.
replace (
	github.com/jcsvwinston/nucleus => ../..
	github.com/jcsvwinston/nucleus/admin/proto => ../proto
)

require (
	connectrpc.com/connect v1.19.2
	github.com/google/uuid v1.6.0
	github.com/jcsvwinston/nucleus v0.0.0-00010101000000-000000000000
	github.com/jcsvwinston/nucleus/admin/proto v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.23.2
	golang.org/x/net v0.53.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)
