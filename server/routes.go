package server

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"echo_rest_api/server/handler"
)

func assignRoutesAndHandlers(e *echo.Echo, h *handler.Handler) {
	// Index
	e.GET("/", h.Index)

	// Tester
	e.GET("/cache_test", h.CacheTest)

	// Metrics management
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Users processes management
	e.POST("/login", h.Login)
	e.POST("/invite", h.Invite)
	e.POST("/validate_invite", h.ValidateInvite)
	e.POST("/signup", h.SignUp)

	// Users management
	users := e.Group("/users")

	users.GET("", h.UserGetAll)

	users.PUT("/:userID", h.UserUpdate)

	users.DELETE("/:userID", h.UserDelete)

	// Stats
	stats := e.Group("/stats")

	stats.GET("", h.StatsGetData)
}
