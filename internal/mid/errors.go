package mid

import (
	"context"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/ardanlabs/service/internal/platform/web"
	"log"
	"net/http"
)

// ErrorHandler for catching and responding errors.
func Errors(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			// Run the next handler and catch any propagated error
			if err := handler(ctx, w, r); err != nil {
				log.Println("ERROR", "trace_id", v.TraceID, "ERROR", err)

			var er validate.ErrorResponse
			var status int
			switch {
			case validate.IsFieldErrors(err):
				fieldErrors := validate.GetFieldErrors(err)
				er = validate.ErrorResponse{
					Error:  "data validation error",
					Fields: fieldErrors.Fields(),
				}
				status = http.StatusBadRequest

			case validate.IsRequestError(err):
				reqErr := validate.GetRequestError(err)
				er = validate.ErrorResponse{
					Error: reqErr.Error(),
				}
				status = reqErr.Status
			default:
				er = validate.ErrorResponse{
					Error: http.StatusText(http.StatusInternalServerError),
				}
				status = http.StatusInternalServerError
			}

			if err:= web.Respond(ctx, w, er, status); err!=nil {
				return err
			}

			if ok := web.IsShutdown(err); ok {
				return err
			}
		}
		return nil
		}
	return h
	}


