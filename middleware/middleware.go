package middleware

import (
	"time"

	"github.com/BadKid90s/chilix-msg/core"
)

// LoggingMiddleware 日志中间件
func LoggingMiddleware() core.Middleware {
	return func(next core.Handler) core.Handler {
		return func(ctx core.Context) error {
			start := time.Now()
			err := next(ctx)
			duration := time.Since(start)

			ctx.Logger().Infof("Processed message %s in %v", ctx.MessageType(), duration)
			return err
		}
	}
}
