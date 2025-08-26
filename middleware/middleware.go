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

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() core.Middleware {
	return func(next core.Handler) core.Handler {
		return func(ctx core.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = core.ErrHandlerPanic
					ctx.Logger().Errorf("Recovered from panic in handler for %s: %v", ctx.MessageType(), r)
				}
			}()
			return next(ctx)
		}
	}
}
