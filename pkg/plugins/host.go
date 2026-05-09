package plugins

import (
	"context"
	"time"
)

// Host is the stable plugin runtime contract exposed by Nucleus.
// It allows callers to discover providers, probe capabilities, and execute
// capability requests independently from the underlying transport/runtime.
type Host interface {
	CollectInventory(pathEnv string, builtinMailProviders []string, probeTimeout time.Duration) []Descriptor
	ProbeCapabilities(ctx context.Context, binaryPath string, timeout time.Duration) ([]string, error)
	ExecuteRequest(ctx context.Context, binaryPath string, request RequestEnvelope, timeout time.Duration) (ResponseEnvelope, error)
}

// LocalHost is the default Host implementation backed by local executables.
type LocalHost struct{}

func (LocalHost) CollectInventory(pathEnv string, builtinMailProviders []string, probeTimeout time.Duration) []Descriptor {
	return CollectInventory(pathEnv, builtinMailProviders, probeTimeout)
}

func (LocalHost) ProbeCapabilities(ctx context.Context, binaryPath string, timeout time.Duration) ([]string, error) {
	return ProbeCapabilities(ctx, binaryPath, timeout)
}

func (LocalHost) ExecuteRequest(ctx context.Context, binaryPath string, request RequestEnvelope, timeout time.Duration) (ResponseEnvelope, error) {
	return ExecuteRequest(ctx, binaryPath, request, timeout)
}
