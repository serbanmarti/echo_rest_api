package handler

import (
	"net/http"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	Handler struct {
		JwtSecret string
		JwtExp    time.Duration
		DB        *mongo.Database
		Cache     *ttlcache.Cache
		FEndpoint string
		SMTP
	}
	SMTP struct {
		Host string
		Port int
		User string
		Pass string
	}
)

// HTTPSuccess returns a formatted HTTP Success response
func HTTPSuccess(c echo.Context, d interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"error":   false,
		"message": "Action completed successfully",
		"data":    d,
	})
}
