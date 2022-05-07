package car

import (
	"context"
	"github.com/go-kit/kit/log"
	"time"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) PostCar(ctx context.Context, car Car) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "PostCar", "id", car.ID, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.PostCar(ctx, car)
}

func (mw loggingMiddleware) GetCar(ctx context.Context, id string) (car Car, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetCar", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetCar(ctx, id)
}

func (mw loggingMiddleware) DeleteCar(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "DeleteCar", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.DeleteCar(ctx, id)
}
