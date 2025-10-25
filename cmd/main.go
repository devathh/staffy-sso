package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devathh/staffy-sso/internal/app"
)

func main() {
	app, cleanup, err := app.SetupApp()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer cleanup()

	go func() {
		if err := app.Start(); err != nil {
			slog.Error(err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		slog.Error(err.Error())
	}
}
