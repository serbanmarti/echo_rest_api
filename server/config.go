package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"echo_rest_api/internal"
)

func configureEcho(e *echo.Echo, env *internal.Environ) {
	// Remove Echo startup banner
	e.HideBanner = true

	// Configure Echo port
	e.Server.Addr = fmt.Sprintf(":%d", env.ServerPort)

	// Configure Echo to remove all trailing slashes
	e.Pre(middleware.RemoveTrailingSlash())

	// Configure Echo panic recovery
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 4 << 10, // 4 KB
	}))

	// Configure Echo logging
	e.Logger.SetLevel(log.INFO)
	e.Use(middleware.Logger())

	// Configure Echo Prometheus
	prom := internal.NewMetrics()
	e.Use(prom.Handle)

	// Configure Echo validator
	v, err := internal.CreateValidator()
	if err != nil {
		e.Logger.Fatal(err)
	}
	e.Validator = v

	// Configure Echo error handler
	e.HTTPErrorHandler = internal.ErrorHandler

	// Set a schema for the FE endpoint
	var schema string
	if schema = "http://"; env.Secure {
		schema = "https://"
	}
	env.FEndpoint = schema + env.FEndpoint

	// Configure Echo CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{env.FEndpoint},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowCredentials: true,
	}))
}
