module github.com/jcsvwinston/nucleus/admin/server

go 1.26.3

// Real require blocks (connectrpc, prometheus, golang.org/x/net/http2) are
// added by Phase 4 when the admin server is implemented.

replace (
	github.com/jcsvwinston/nucleus/admin/proto => ../proto
)
