package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"echo_rest_api/pkg/server"
)

func main() {
	// Instantiate the Echo REST server and database connection
	e, db := server.InitServer()

	// Start the server
	go func() {
		if err := e.Start(e.Server.Addr); err != nil {
			e.Logger.Info(err)
		}
	}()

	// Graceful shutdown of the server with a timeout
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit
	e.Logger.Info("gracefully shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.Disconnect(context.TODO()); err != nil {
		e.Logger.Fatal(err)
	}
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
