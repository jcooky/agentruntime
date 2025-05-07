package genkit

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/sdk/trace"
)

type loggingSpanProcessor struct {
	verbose bool
	logger  *slog.Logger
}

func (l *loggingSpanProcessor) OnStart(ctx context.Context, s trace.ReadWriteSpan) {
	l.logger.Info("span start", l.buildArgs(s)...)
}

func (l *loggingSpanProcessor) OnEnd(s trace.ReadOnlySpan) {
	l.logger.Info("span end", l.buildArgs(s)...)
}

func (l *loggingSpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

func (l *loggingSpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

var _ trace.SpanProcessor = (*loggingSpanProcessor)(nil)

func (l *loggingSpanProcessor) buildArgs(s trace.ReadOnlySpan) []any {
	args := []any{
		slog.String("name", s.Name()),
	}
	for _, attr := range s.Attributes() {
		key := string(attr.Key)
		value := attr.Value.Emit()
		if !l.verbose && len(value) > 256 {
			continue
		}
		args = append(args, slog.String(key, value))
	}

	return args
}
