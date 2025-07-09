package server

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"cDNS/internal/api"
	"cDNS/internal/config"
	"cDNS/internal/logger"
)

func RunAPI(cmd *cobra.Command, args []string) {
	cfg := config.GetConfigFromFlags(cmd)
	logger.InitLogger(cfg.LogLevel)

	h := api.NewHandler(logger.GetLogger())
	h.SetupRoutes()
	r := h.GetRouter()

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		logger.GetLogger().Fatal("Failed to get port flag", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,

		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create a channel to capture server errors
	serverErrors := make(chan error, 1)

	go func() {
		logger.GetLogger().Info("Starting API server", zap.Int("port", port))
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Set up signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either a signal or server error
	select {
	case <-quit:
		logger.GetLogger().Info("Received shutdown signal")
	case err = <-serverErrors:
		logger.GetLogger().Fatal("Server failed to start", zap.Error(err))
	}

	logger.GetLogger().Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		logger.GetLogger().Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.GetLogger().Info("Server exiting")
}
