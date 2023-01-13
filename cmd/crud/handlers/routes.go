package handlers

import (
	"github.com/ardanlabs/service/internal/mid"
	"os"

	"github.com/ardanlabs/service/internal/platform/web"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
}

// API returns a handler for a set of routes.
func APIMux(cfg APIMuxConfig) *web.App {

	//Construct the web.App which holds all routes as well as common Middleware
	app := web.New(cfg.Shutdown, mid.RequestLogger, mid.ErrorHandler)

	v1(app)
	return app
}

func v1(app *web.App) {
	const version = "v1"
	var u User
	app.Handle("GET", version, "/users", u.List)
}
