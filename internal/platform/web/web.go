package web

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux"
	"github.com/pborman/uuid"
)

// TraceIDHeader is the header added to outgoing requests which adds the
// traceID to it.
const TraceIDHeader = "X-Trace-ID"

// KeyValues is how request values or stored/retrieved.
const KeyValues ctxKey = 1

// A Handler is a type that handles an http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw []Middleware
}

// New creates an App value that handle a set of routes for the application.
func New(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown: shutdown,
		mw: mw,

	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// Handle is our mechanism for mounting Handlers for a given HTTP verb and path
// pair, this makes for really easy, convenient routing.
//func (a *App) Handle(verb, path string, handler Handler) {
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {
	// Wrap up the application-wide first, this will call the first function
	// of each middleware which will return a function of type Handler.
	handler = wrapMiddleware(handler, mw)
	//Add the application's general middleware to the handler chain
	handler = wrapMiddleware(handler, a.mw)


	// The function to execute for each request.
	h := func(w http.ResponseWriter, r *http.Request) {

		// Set the context with the required values to
		// process the request.
		v := Values{
			TraceID: uuid.New(),
			Now:     time.Now(),
		}
		ctx := context.WithValue(r.Context(), KeyValues, &v)

		// Set the trace id on the outgoing requests before any other header to
		// ensure that the trace id is ALWAYS added to the request regardless of
		// any error occuring or not.
		w.Header().Set(TraceIDHeader, v.TraceID)

		//// Call the wrapped handler functions.

		if err := handler(ctx, w, r); err != nil {
			if validateShutdown(err) {
				a.SignalShutdown()
				return
			}
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	// Add this handler for the specified verb and route.
	a.ContextMux.Handle(method, finalPath, h)
}

// validateShutdown validates the error for special conditions that do not
// warrant an actual shutdown by the system.
func validateShutdown(err error) bool {

	// Ignore syscall.EPIPE and syscall.ECONNRESET errors which occurs
	// when a write operation happens on the http.ResponseWriter that
	// has simultaneously been disconnected by the client (TCP
	// connections is broken). For instance, when large amounts of
	// data is being written or streamed to the client.
	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// https://gosamples.dev/broken-pipe/
	// https://gosamples.dev/connection-reset-by-peer/

	switch {
	case errors.Is(err, syscall.EPIPE):

		// Usually, you get the broken pipe error when you write to the connection after the
		// RST (TCP RST Flag) is sent.
		// The broken pipe is a TCP/IP error occurring when you write to a stream where the
		// other end (the peer) has closed the underlying connection. The first write to the
		// closed connection causes the peer to reply with an RST packet indicating that the
		// connection should be terminated immediately. The second write to the socket that
		// has already received the RST causes the broken pipe error.
		return false

	case errors.Is(err, syscall.ECONNRESET):

		// Usually, you get connection reset by peer error when you read from the
		// connection after the RST (TCP RST Flag) is sent.
		// The connection reset by peer is a TCP/IP error that occurs when the other end (peer)
		// has unexpectedly closed the connection. It happens when you send a packet from your
		// end, but the other end crashes and forcibly closes the connection with the RST
		// packet instead of the TCP FIN, which is used to close a connection under normal
		// circumstances.
		return false
	}

	return true
}
