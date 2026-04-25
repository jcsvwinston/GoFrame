package tasks

import "strings"

const (
	QueueActionPause         = "pause"
	QueueActionUnpause       = "unpause"
	QueueActionRetry         = "retry"
	QueueActionArchiveRetry  = "archive-retry"
	QueueActionRetryArchived = "retry-archived"
	QueueActionPurgeArchived = "purge-archived"
)

var supportedQueueActions = []string{
	QueueActionPause,
	QueueActionUnpause,
	QueueActionRetry,
	QueueActionArchiveRetry,
	QueueActionRetryArchived,
	QueueActionPurgeArchived,
}

// RuntimeSnapshot describes queue/worker state discoverable from the provider runtime.
type RuntimeSnapshot struct {
	Enabled           bool                      `json:"enabled"`
	GeneratedAt       string                    `json:"generated_at"`
	Reason            string                    `json:"reason,omitempty"`
	Queues            []RuntimeQueueSnapshot    `json:"queues"`
	Schedules         []RuntimeScheduleSnapshot `json:"schedules"`
	Servers           []RuntimeServerSnapshot   `json:"servers"`
	Workers           []RuntimeWorkerSnapshot   `json:"workers"`
	TotalSchedules    int                       `json:"total_schedules"`
	TotalQueues       int                       `json:"total_queues"`
	TotalServers      int                       `json:"total_servers"`
	TotalWorkers      int                       `json:"total_workers"`
	TotalSize         int                       `json:"total_size"`
	TotalPending      int                       `json:"total_pending"`
	TotalActive       int                       `json:"total_active"`
	TotalScheduled    int                       `json:"total_scheduled"`
	TotalRetry        int                       `json:"total_retry"`
	TotalArchived     int                       `json:"total_archived"`
	TotalCompleted    int                       `json:"total_completed"`
	TotalAggregating  int                       `json:"total_aggregating"`
	TotalProcessed    int                       `json:"total_processed_today"`
	TotalFailed       int                       `json:"total_failed_today"`
	TotalProcessedAll int                       `json:"total_processed_all"`
	TotalFailedAll    int                       `json:"total_failed_all"`
}

// RuntimeQueueSnapshot holds one queue aggregate.
type RuntimeQueueSnapshot struct {
	Name           string `json:"name"`
	Paused         bool   `json:"paused"`
	LatencyMS      int64  `json:"latency_ms"`
	Size           int    `json:"size"`
	Pending        int    `json:"pending"`
	Active         int    `json:"active"`
	Scheduled      int    `json:"scheduled"`
	Retry          int    `json:"retry"`
	Archived       int    `json:"archived"`
	Completed      int    `json:"completed"`
	Aggregating    int    `json:"aggregating"`
	ProcessedToday int    `json:"processed_today"`
	FailedToday    int    `json:"failed_today"`
	ProcessedAll   int    `json:"processed_all"`
	FailedAll      int    `json:"failed_all"`
}

// RuntimeScheduleSnapshot holds one registered periodic task entry.
type RuntimeScheduleSnapshot struct {
	ID            string `json:"id"`
	Spec          string `json:"spec"`
	TaskType      string `json:"task_type"`
	NextEnqueueAt string `json:"next_enqueue_at,omitempty"`
	PrevEnqueueAt string `json:"prev_enqueue_at,omitempty"`
}

// RuntimeServerSnapshot holds one server aggregate.
type RuntimeServerSnapshot struct {
	ID             string         `json:"id"`
	Host           string         `json:"host"`
	PID            int            `json:"pid"`
	Status         string         `json:"status"`
	StartedAt      string         `json:"started_at,omitempty"`
	Concurrency    int            `json:"concurrency"`
	StrictPriority bool           `json:"strict_priority"`
	Queues         map[string]int `json:"queues"`
	ActiveWorkers  int            `json:"active_workers"`
}

// RuntimeWorkerSnapshot describes one active worker task.
type RuntimeWorkerSnapshot struct {
	ServerID  string `json:"server_id"`
	Host      string `json:"host"`
	PID       int    `json:"pid"`
	Queue     string `json:"queue"`
	TaskID    string `json:"task_id"`
	TaskType  string `json:"task_type"`
	StartedAt string `json:"started_at,omitempty"`
	Deadline  string `json:"deadline,omitempty"`
}

// QueueActionResult is the result of one operational queue action.
type QueueActionResult struct {
	Enabled     bool   `json:"enabled"`
	GeneratedAt string `json:"generated_at"`
	Queue       string `json:"queue"`
	Action      string `json:"action"`
	Applied     bool   `json:"applied"`
	Affected    int    `json:"affected,omitempty"`
	Message     string `json:"message,omitempty"`
}

func NormalizeQueueAction(raw string) (string, bool) {
	action := strings.ToLower(strings.TrimSpace(raw))
	for _, candidate := range supportedQueueActions {
		if action == candidate {
			return action, true
		}
	}
	return action, false
}

func SupportedQueueActions() []string {
	out := make([]string, len(supportedQueueActions))
	copy(out, supportedQueueActions)
	return out
}
