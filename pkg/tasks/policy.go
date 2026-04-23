package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

// EnqueuePolicy describes the supported explicit enqueue-policy subset.
// MaxRetry uses -1 to keep Asynq defaults; 0 disables retries.
type EnqueuePolicy struct {
	Queue     string
	MaxRetry  int
	Timeout   time.Duration
	ProcessIn time.Duration
	Retention time.Duration
}

func DefaultEnqueuePolicy() EnqueuePolicy {
	return EnqueuePolicy{
		MaxRetry: -1,
	}
}

func (p EnqueuePolicy) Validate() error {
	if p.MaxRetry < -1 {
		return fmt.Errorf("tasks.EnqueuePolicy: max retry must be >= -1")
	}
	if p.Timeout < 0 {
		return fmt.Errorf("tasks.EnqueuePolicy: timeout must be >= 0")
	}
	if p.ProcessIn < 0 {
		return fmt.Errorf("tasks.EnqueuePolicy: process_in must be >= 0")
	}
	if p.Retention < 0 {
		return fmt.Errorf("tasks.EnqueuePolicy: retention must be >= 0")
	}
	return nil
}

func (p EnqueuePolicy) Options() ([]asynq.Option, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	opts := make([]asynq.Option, 0, 5)
	if queue := strings.TrimSpace(p.Queue); queue != "" {
		opts = append(opts, asynq.Queue(queue))
	}
	if p.MaxRetry >= 0 {
		opts = append(opts, asynq.MaxRetry(p.MaxRetry))
	}
	if p.Timeout > 0 {
		opts = append(opts, asynq.Timeout(p.Timeout))
	}
	if p.ProcessIn > 0 {
		opts = append(opts, asynq.ProcessIn(p.ProcessIn))
	}
	if p.Retention > 0 {
		opts = append(opts, asynq.Retention(p.Retention))
	}
	return opts, nil
}

func (m *Manager) EnqueueJSONWithPolicy(taskType string, payload any, policy EnqueuePolicy) (*asynq.TaskInfo, error) {
	return m.EnqueueJSONCtxWithPolicy(context.Background(), taskType, payload, policy)
}

func (m *Manager) EnqueueJSONCtxWithPolicy(ctx context.Context, taskType string, payload any, policy EnqueuePolicy) (*asynq.TaskInfo, error) {
	opts, err := policy.Options()
	if err != nil {
		return nil, err
	}
	return m.EnqueueJSONCtx(ctx, taskType, payload, opts...)
}
