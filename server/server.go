package server

import (
	"context"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"echo_rest_api/database"
	"echo_rest_api/internal"
	"echo_rest_api/security"
	"echo_rest_api/server/handler"
)

// Initialize a Echo server and DB connection
func InitServer() (*echo.Echo, *mongo.Client) {
	// Create a new Echo instance
	e := echo.New()

	// Get configuration from environment
	env, err := internal.GetEnv()
	if err != nil {
		e.Logger.Fatalf("Failed to get environment variables: %s", err)
	}

	// Configure the Echo instance
	configureEcho(e, env)

	// Configure Echo security, if requested
	if !env.SkipChecks {
		security.ConfigureEchoSecurity(e, env)
	}

	// Create the database client & connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dbClient, dbConn := database.NewDBClientAndConnection(ctx, env, e.Logger)

	// Initialize the data cache
	c := ttlcache.NewCache()
	c.SkipTtlExtensionOnHit(true)

	// Initialize the route handler
	h := &handler.Handler{
		JwtSecret: env.JwtSecret,
		JwtExp:    env.JwtExp,
		DB:        dbConn,
		Cache:     c,
		FEndpoint: env.FEndpoint,
		SMTP: handler.SMTP{
			Host: env.SMTPHost,
			Port: env.SMTPPort,
			User: env.SMTPUser,
			Pass: env.SMTPPass,
		},
	}

	// Assign the routes & handlers
	assignRoutesAndHandlers(e, h)

	// Return
	return e, dbClient
}
