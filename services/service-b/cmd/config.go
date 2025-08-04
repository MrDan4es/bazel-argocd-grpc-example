package main

import (
	_ "embed"

	"github.com/rs/zerolog"
)

var (
	//go:embed default-config.yml
	defaultConfigBytes []byte
)

type config struct {
	Logger struct {
		Level zerolog.Level
	}

	HTTP struct {
		ListenAddr string
	}

	ServiceAAddr string
}

func (c config) DefaultConfigBytes() []byte {
	return defaultConfigBytes
}

func (c config) EnvPrefix() string {
	return serviceName
}
