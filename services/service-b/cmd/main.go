package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	libConfig "github.com/mrdan4es/bazel-argocd-grpc-example/pkg/config"
	apb "github.com/mrdan4es/bazel-argocd-grpc-example/services/service-a/api/v1"
	"github.com/mrdan4es/bazel-argocd-grpc-example/services/service-b/internal/server"
)

const serviceName = "service-b"

func main() {
	var configFile = pflag.StringP("config", "c", "", "path to service config file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := libConfig.Load[config](*configFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: load service config: %v", serviceName, err)
		os.Exit(1)
	}

	log := zerolog.New(os.Stdout).
		Level(cfg.Logger.Level).
		With().
		Str("service", serviceName).
		Logger()
	ctx = log.WithContext(ctx)

	err = run(ctx, cfg)

	switch {
	case errors.Is(err, context.Canceled):
		log.Info().Msg("gracefully stopped")
	case err != nil:
		log.Fatal().Err(err).Msg("unexpectedly terminated")
	}
}

func run(ctx context.Context, cfg *config) error {
	log := zerolog.Ctx(ctx)

	serviceAConn, err := grpc.NewClient(
		cfg.ServiceAAddr,
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := serviceAConn.Close(); err != nil {
			log.Err(err).Msg("close service A conn")
		}
	}()

	serviceAClient := apb.NewServiceAClient(serviceAConn)

	srv := server.New(ctx, serviceAClient)

	httpServer := &http.Server{
		Addr:    cfg.HTTP.ListenAddr,
		Handler: srv,
	}
	go func() {
		<-ctx.Done()

		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Err(err).Msg("shutting down http server")
		}
	}()

	log.Info().Msgf("serving on addr %s...", cfg.HTTP.ListenAddr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
