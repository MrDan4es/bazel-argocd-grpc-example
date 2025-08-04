package config

import (
	"bytes"
	"errors"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	zlog "github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Config interface {
	DefaultConfigBytes() []byte
	EnvPrefix() string
}

func Load[Cfg Config](configFilename string) (*Cfg, error) {
	var cfg Cfg
	v := viper.New()

	if configFilename != "" {
		v.SetConfigFile(configFilename)
	}

	v.AutomaticEnv()
	v.SetEnvPrefix(cfg.EnvPrefix())
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.SetConfigType("yml")

	if err := v.ReadConfig(bytes.NewBuffer(cfg.DefaultConfigBytes())); err != nil {
		return nil, err
	}

	if err := v.MergeInConfig(); err != nil {
		if errors.Is(err, &viper.ConfigParseError{}) {
			return nil, err
		}
	}

	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		StringToZerologLevelHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)

	if err := v.Unmarshal(&cfg, viper.DecodeHook(decodeHooks)); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func StringToZerologLevelHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(zlog.Level(1)) || f.Kind() != reflect.String {
			return data, nil
		}

		return zlog.ParseLevel(data.(string))
	}
}
