package network

import (
	"context"
	"github.com/habiliai/agentruntime/entity"
	"time"
)

func (s *service) runHealthChecker(ctx context.Context) {
	s.logger.Info("Starting health checker")
	defer s.logger.Info("Health checker stopped")

	ticker := time.NewTicker(60 * time.Second)
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

			for _, runtime := range runtimes {
				if time.Now().Sub(runtime.LastLiveAt).Minutes() <= 2.5 {
					continue
				}

				s.logger.Warn("Agent runtime is not alive", "addr", runtime.Addr, "last_live_at", runtime.LastLiveAt)
				if err := runtime.Delete(s.db); err != nil {
					s.logger.Error("Failed to delete agent runtime", "addr", runtime.Addr, "err", err)
				} else {
					s.logger.Info("Deleted agent runtime", "addr", runtime.Addr)
				}
			}
		}
	}
}
