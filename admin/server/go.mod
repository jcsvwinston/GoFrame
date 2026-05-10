module github.com/jcsvwinston/nucleus/admin/server

go 1.26.3

// Real require blocks land as the server is implemented (Phase 4). The
// proto generator already added connectrpc + protobuf to admin/proto;
// admin/server pulls connectrpc directly here.

replace github.com/jcsvwinston/nucleus/admin/proto => ../proto

require (
	connectrpc.com/connect v1.19.2
	github.com/jcsvwinston/nucleus/admin/proto v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.54.0
	google.golang.org/protobuf v1.36.11
)

require golang.org/x/text v0.37.0 // indirect
