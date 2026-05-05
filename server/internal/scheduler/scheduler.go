package scheduler

import (
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/logger"
	"kenaito-conf/server/internal/model"
	"time"
)

func Start() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
			affected, err := dao.Engine.Where("last_ping_at < ?", fiveMinutesAgo).Delete(&model.ClientInstance{})
			if err != nil {
				logger.SugaredLogger.Errorf("[scheduler] cleanup expired client instances error: %v", err)
				continue
			}
			if affected > 0 {
				logger.SugaredLogger.Infof("[scheduler] cleaned up %d expired client instance(s)", affected)
			}
		}
	}()
	logger.SugaredLogger.Info("[scheduler] client instance cleanup task started, interval: 1m, expire: 5m")
}
