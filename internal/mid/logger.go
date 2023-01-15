package mid

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ardanlabs/service/internal/platform/web"
)

// RequestLogger writes some information about the request to the logs in
// the format: TraceID : (200) GET /foo -> IP ADDR (latency)
func RequestLogger(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			//v := ctx.Value(web.KeyValues).(*web.Values)
			v:= web.GetValues(ctx)

			log.Printf("request started trace_id %s, method %s, path %s, remoteaddr %s",
				v.TraceID, r.Method, r.URL.Path, r.RemoteAddr)

			err:= handler(ctx, w, r)

			log.Printf("request completed trace_id %s : (%d) : method %s, path %s, remoteaddr %s, time (%s)",
				v.TraceID, v.StatusCode, r.Method, r.URL.Path, r.RemoteAddr, time.Since(v.Now))

			// This is the top of the food chain. At this point all error
			// handling has been done including logging.
			return err
		}

		return h
	}


