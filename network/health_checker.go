package network

import (
	"context"
	"github.com/habiliai/agentruntime/entity"
	myerrors "github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/grpcutils"
	"github.com/pkg/errors"
	"google.golang.org/grpc/health/grpc_health_v1"
	"sync"
	"time"
)

func (s *service) runHealthChecker(ctx context.Context) {
	s.logger.Info("Starting health checker")
	defer s.logger.Info("Health checker stopped")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var runtimes []entity.AgentRuntime
			if err := s.db.Find(&runtimes).Error; err != nil {
				s.logger.Error("Failed to find agent runtimes", "err", err)
				continue
			}

			for i := 0; i < len(runtimes); i += 25 {
				end := min(i+25, len(runtimes))

				var wg sync.WaitGroup
				for _, runtime := range runtimes[i:end] {
					wg.Add(1)
					go func() {
						defer wg.Done()
						ctx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, myerrors.ErrTimeout)
						defer cancel()

						if err := func() error {
							conn, err := grpcutils.NewClient(runtime.Addr, runtime.Secure)
							if err != nil {
								return errors.Wrapf(err, "failed to create gRPC client. addr: %s", runtime.Addr)
							}
							healthClient := grpc_health_v1.NewHealthClient(conn)

							if _, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{}); err != nil {
								return errors.Wrapf(err, "failed to check health. addr: %s", runtime.Addr)
							}

							return nil
						}(); err != nil {
							s.logger.Warn("Health check failed", "addr", runtime.Addr, "err", err)
							if errors.Is(err, myerrors.ErrTimeout) {
								if err := runtime.Delete(s.db); err != nil {
									s.logger.Error("failed to delete agent runtime", "addr", runtime.Addr, "err", err)
								}
							}
						}
					}()
				}
				wg.Wait()
			}
		}
	}
}
