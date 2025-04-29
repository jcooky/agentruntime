package genkit

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/sdk/trace"
)

type loggingSpanProcessor struct {
	logger *slog.Logger
}

func (l loggingSpanProcessor) OnStart(ctx context.Context, s trace.ReadWriteSpan) {
	l.logger.Info("span start", "name", s.Name(), "events", s.Events())
}

func (l loggingSpanProcessor) OnEnd(s trace.ReadOnlySpan) {
	l.logger.Info("span end", "name", s.Name(), "events", s.Events())
}

func (l loggingSpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

func (l loggingSpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

var _ trace.SpanProcessor = (*loggingSpanProcessor)(nil)
