package collector

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/archway-network/validator-exporter/pkg/config"
	"github.com/archway-network/validator-exporter/pkg/grpc"
	log "github.com/archway-network/validator-exporter/pkg/logger"
	"github.com/archway-network/validator-exporter/pkg/types"
)

const (
	missedBlocksMetricName = "cosmos_validator_missed_blocks"
)

var missedBlocks = prometheus.NewDesc(
	missedBlocksMetricName,
	"Returns missed blocks for a validator.",
	[]string{
		"chain_name",
		"chain_id",
		"valcons",
		"valoper",
		"moniker",
		"jailed",
		"tombstoned",
		"bond_status",
	},
	nil,
)

type ValidatorsCollector struct {
	Cfg config.Config
}

func (vc ValidatorsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- missedBlocks
}

func (vc ValidatorsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(vc.Cfg.Timeout)*time.Second,
	)

	defer cancel()

	vals, err := grpc.SigningValidators(ctx, vc.Cfg)
	if err != nil {
		log.Error(fmt.Sprintf("error getting signing validators: %s", err))
	} else {
		log.Debug("Start collecting", zap.String("metric", missedBlocksMetricName))

		for _, m := range vc.missedBlocksMetrics(vals) {
			ch <- m
		}

		log.Debug("Stop collecting", zap.String("metric", missedBlocksMetricName))
	}
}

func (vc ValidatorsCollector) missedBlocksMetrics(vals []types.Validator) []prometheus.Metric {
	metrics := []prometheus.Metric{}

	for _, val := range vals {
		metrics = append(
			metrics,
			prometheus.MustNewConstMetric(
				missedBlocks,
				prometheus.GaugeValue,
				float64(val.MissedBlocks),
				[]string{
					vc.Cfg.ChainName,
					vc.Cfg.ChainID,
					val.ConsAddress,
					val.OperatorAddress,
					val.Moniker,
					strconv.FormatBool(val.Jailed),
					strconv.FormatBool(val.Tombstoned),
					val.BondStatus,
				}...,
			),
		)
	}

	return metrics
}
