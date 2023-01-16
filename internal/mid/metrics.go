package mid

import (
	"context"
	"github.com/ardanlabs/service/internal/platform/web"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/metrics"
)

// Metrics updates program counters.
func Metrics(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx = metrics.Set(ctx)

			err := handler(ctx, w, r)

			metrics.AddRequests(ctx)
			metrics.AddGoroutines(ctx)

			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}

		return h
	}
