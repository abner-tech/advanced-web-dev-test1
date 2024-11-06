package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (a *applicationDependences) serve() error {
	apiServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.config.port),
		Handler:      a.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(a.logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)                      //detect and receive shutdown signal
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) //got signal
		s := <-quit                                          //dont pass here until signal is received

		a.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		//if everything ok, start tru shutdown process
		shutdownError <- apiServer.Shutdown(ctx)
	}()

	a.logger.Info("Starting Server", "address", apiServer.Addr, "environment", a.config.environment, "rateLimiterConfig", a.config.limiter)

	err := apiServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	a.logger.Info("stopped server", "address", apiServer.Addr)
	return nil
}
