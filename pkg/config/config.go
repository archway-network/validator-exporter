package config

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var errConfig = errors.New("config error")

func configError(msg string) error {
	return fmt.Errorf("%w: %s", errConfig, msg)
}

type Config struct {
	Addr      string `env:"GRPC_ADDR" envDefault:"grpc.constantine.archway.tech:443"`
	TLS       bool   `env:"GRPC_TLS_ENABLED" envDefault:"true"`
	Timeout   int    `env:"GRPC_TIMEOUT_SECONDS" envDefault:"5"`
	Prefix    string `env:"PREFIX" envDefault:"archway"`
	ChainName string `env:"CHAIN_NAME" envDefault:"archway"`
	ChainID   string `env:"CHAIN_ID" envDefault:"constantine-3"`
}

func (c Config) GRPCConn() (*grpc.ClientConn, error) {
	transportCreds := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))

	if !c.TLS {
		transportCreds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.Dial(
		c.Addr,
		transportCreds,
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return nil, configError(err.Error())
	}

	return conn, nil
}
