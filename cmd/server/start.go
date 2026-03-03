package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
	"github.com/SantiagoBobrik/agent-pulse/internal/pid"
	"github.com/SantiagoBobrik/agent-pulse/internal/server"
	"github.com/spf13/cobra"
)

var portFlag int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the event bridge server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		port := cfg.Port
		if portFlag != 0 {
			port = portFlag
		}

		if err := pid.Write(os.Getpid()); err != nil {
			return fmt.Errorf("failed to write pid file: %w", err)
		}
		defer pid.Remove()

		broker := server.NewBroker()
		dispatcher := server.NewDispatcher(broker)
		srv := server.NewServer(dispatcher, broker, port, cfg.BindAddress)

		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start()
		}()

		logger.Info("server started", "addr", port)

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
		}

		logger.Info("server shutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", "error", err)
		}

		logger.Info("server stopped")
		return nil
	},
}

func init() {
	startCmd.Flags().IntVarP(&portFlag, "port", "p", 0, "server listen port (overrides config)")
}
