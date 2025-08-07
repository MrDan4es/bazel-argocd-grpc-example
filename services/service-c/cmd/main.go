package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	envoyAuth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	libConfig "github.com/mrdan4es/bazel-argocd-grpc-example/pkg/config"
	"github.com/mrdan4es/bazel-argocd-grpc-example/services/service-c/internal/server"
)

const serviceName = "service-c"

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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.ListenPort))
	if err != nil {
		return fmt.Errorf("create listener: %w", err)
	}

	serverRegister := grpc.NewServer()
	go func() {
		<-ctx.Done()

		log.Info().Msgf("gracefully stopping gRPC server")
		serverRegister.GracefulStop()
	}()

	srv := server.New(ctx)
	envoyAuth.RegisterAuthorizationServer(serverRegister, srv)

	log.Info().Msgf("gRPC server started on port %d", cfg.GRPC.ListenPort)
	if err = serverRegister.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}

	return nil
}
