package client

import (
	"context"
	"log/slog"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
)

func RegistryClient(_ context.Context) *regclient.RegClient {
	//slog.SetLogLoggerLevel(slog.LevelDebug)
	return regclient.New(
		regclient.WithSlog(slog.Default()),
		regclient.WithConfigHost(
			config.Host{
				Name: "localhost",
				TLS:  config.TLSInsecure,
			},
		),
		regclient.WithDockerCreds(),
	)
}
