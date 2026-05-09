package signals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jcsvwinston/nucleus/pkg/observe"
	"github.com/redis/go-redis/v9"
)

const defaultRedisChannelPrefix = "goframe:signals:"

var (
	ErrRedisURLRequired = errors.New("signals: redis url is required")
	ErrSignalRequired   = errors.New("signals: signal is required")
	ErrNilHandler       = errors.New("signals: handler is required")
	ErrNilBus           = errors.New("signals: bus is nil")
	ErrNilRelay         = errors.New("signals: relay is nil")
)

type RedisRelayConfig struct {
	RedisURL      string
	ChannelPrefix string
}

type RedisRelay struct {
	client        *redis.Client
	channelPrefix string
	logger        *slog.Logger
	closeOnce     sync.Once
}

type redisEnvelope struct {
	Signal      Signal `json:"signal"`
	ModelName   string `json:"model_name,omitempty"`
	Payload     any    `json:"payload,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	TraceID     string `json:"trace_id,omitempty"`
}

func NewRedisRelay(cfg RedisRelayConfig, logger *slog.Logger) (*RedisRelay, error) {
	redisURL := strings.TrimSpace(cfg.RedisURL)
	if redisURL == "" {
		return nil, ErrRedisURLRequired
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("signals.NewRedisRelay parse redis url: %w", err)
	}
	if logger == nil {
		logger = slog.Default()
	}

	prefix := strings.TrimSpace(cfg.ChannelPrefix)
	if prefix == "" {
		prefix = defaultRedisChannelPrefix
	}

	return &RedisRelay{
		client:        redis.NewClient(options),
		channelPrefix: prefix,
		logger:        logger,
	}, nil
}

func (r *RedisRelay) Close() error {
	if r == nil {
		return nil
	}
	var closeErr error
	r.closeOnce.Do(func() {
		closeErr = r.client.Close()
	})
	return closeErr
}

func (r *RedisRelay) Channel(signal Signal) string {
	if r == nil {
		return ""
	}
	return r.channelPrefix + string(signal)
}

func (r *RedisRelay) Publish(ctx context.Context, event Event) error {
	if r == nil {
		return ErrNilRelay
	}
	if strings.TrimSpace(string(event.Signal)) == "" {
		return ErrSignalRequired
	}

	ctx = normalizeRelayContext(ctx, event.Ctx)
	body, err := json.Marshal(encodeEnvelope(ctx, event))
	if err != nil {
		return fmt.Errorf("signals.RedisRelay.Publish: %w", err)
	}
	if err := r.client.Publish(ctx, r.Channel(event.Signal), body).Err(); err != nil {
		return fmt.Errorf("signals.RedisRelay.Publish: %w", err)
	}
	return nil
}

func (r *RedisRelay) Subscribe(ctx context.Context, signal Signal, handler Handler) error {
	if r == nil {
		return ErrNilRelay
	}
	if strings.TrimSpace(string(signal)) == "" {
		return ErrSignalRequired
	}
	if handler == nil {
		return ErrNilHandler
	}

	ctx = normalizeRelayContext(ctx, nil)
	pubsub := r.client.Subscribe(ctx, r.Channel(signal))
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("signals.RedisRelay.Subscribe: %w", err)
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil
			}

			event, err := decodeEnvelope(msg.Payload)
			if err != nil {
				return fmt.Errorf("signals.RedisRelay.Subscribe decode: %w", err)
			}
			if err := handler(event); err != nil {
				return fmt.Errorf("signals.RedisRelay.Subscribe handler: %w", err)
			}
		}
	}
}

func (r *RedisRelay) ForwardToBus(ctx context.Context, signal Signal, bus *Bus) error {
	if bus == nil {
		return ErrNilBus
	}
	return r.Subscribe(ctx, signal, func(event Event) error {
		return bus.Emit(event)
	})
}

func encodeEnvelope(ctx context.Context, event Event) redisEnvelope {
	out := redisEnvelope{
		Signal:    event.Signal,
		ModelName: event.ModelName,
		Payload:   event.Payload,
	}
	if ctx != nil {
		out.RequestID = observe.RequestIDFromCtx(ctx)
		out.UserID = observe.UserIDFromCtx(ctx)
		out.TraceID = observe.TraceIDFromCtx(ctx)
	}
	out.PublishedAt = time.Now().UTC().Format(time.RFC3339)
	return out
}

func decodeEnvelope(raw string) (Event, error) {
	var env redisEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return Event{}, err
	}
	if strings.TrimSpace(string(env.Signal)) == "" {
		return Event{}, ErrSignalRequired
	}

	ctx := context.Background()
	if env.RequestID != "" {
		ctx = observe.CtxWithRequestID(ctx, env.RequestID)
	}
	if env.UserID != "" {
		ctx = observe.CtxWithUserID(ctx, env.UserID)
	}
	if env.TraceID != "" {
		ctx = observe.CtxWithTraceID(ctx, env.TraceID)
	}

	return Event{
		Signal:    env.Signal,
		ModelName: env.ModelName,
		Payload:   env.Payload,
		Ctx:       ctx,
	}, nil
}

func normalizeRelayContext(primary, fallback context.Context) context.Context {
	if primary != nil {
		return primary
	}
	if fallback != nil {
		return fallback
	}
	return context.Background()
}
