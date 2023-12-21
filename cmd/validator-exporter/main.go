package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/archway-network/validator-exporter/pkg/collector"
	"github.com/archway-network/validator-exporter/pkg/config"
	"github.com/archway-network/validator-exporter/pkg/grpc"
	log "github.com/archway-network/validator-exporter/pkg/logger"
)

const (
	defaultPort = 8008
	timeout     = 10
)

func main() {
	port := flag.Int("p", defaultPort, "Server port")
	logLevel := log.LevelFlag()

	flag.Parse()

	log.SetLevel(*logLevel)

	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err.Error())
	}

	//nolint:govet
	ctx, _ := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.Timeout)*time.Second,
	)

	_, err := grpc.LatestBlockHeight(ctx, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	valsCollector := collector.ValidatorsCollector{
		Cfg: cfg,
	}

	prometheus.MustRegister(valsCollector)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", *port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  timeout * time.Second,
		WriteTimeout: timeout * time.Second,
	}

	log.Info(fmt.Sprintf("Starting server on addr: %s", addr))

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}
