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
