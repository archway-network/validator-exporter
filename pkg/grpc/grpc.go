package grpc

import (
	"context"
	"errors"
	"fmt"

	base "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"

	"github.com/archway-network/validator-exporter/pkg/config"
	"github.com/archway-network/validator-exporter/pkg/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc"

	log "github.com/archway-network/validator-exporter/pkg/logger"
)

const valConsStr = "valcons"

var errEndpoint = errors.New("grpc error")

func endpointError(msg string) error {
	return fmt.Errorf("%w: %s", errEndpoint, msg)
}

type Client struct {
	cfg  config.Config
	conn *grpc.ClientConn
}

func NewClient(cfg config.Config) (Client, error) {
	client := Client{
		cfg: cfg,
	}

	conn, err := cfg.GRPCConn()
	if err != nil {
		return Client{}, endpointError(err.Error())
	}

	client.conn = conn

	return client, nil
}

func (c Client) SignigInfos(ctx context.Context) ([]slashing.ValidatorSigningInfo, error) {
	infos := []slashing.ValidatorSigningInfo{}
	key := []byte{}
	client := slashing.NewQueryClient(c.conn)

	for {
		request := &slashing.QuerySigningInfosRequest{Pagination: &query.PageRequest{Key: key}}

		slashRes, err := client.SigningInfos(ctx, request)
		if err != nil {
			return nil, endpointError(err.Error())
		}

		if slashRes == nil {
			return nil, endpointError("got empty response from signing infos endpoint")
		}

		infos = append(infos, slashRes.GetInfo()...)

		page := slashRes.GetPagination()
		if page == nil {
			break
		}

		key = page.GetNextKey()
		if len(key) == 0 {
			break
		}
	}

	log.Debug(fmt.Sprintf("SigningInfos: %d", len(infos)))

	return infos, nil
}

func (c Client) Validators(ctx context.Context) ([]staking.Validator, error) {
	vals := []staking.Validator{}
	key := []byte{}

	// https://github.com/cosmos/cosmos-sdk/issues/8045#issuecomment-829142440
	encCfg := testutil.MakeTestEncodingConfig()
	interfaceRegistry := encCfg.InterfaceRegistry

	client := staking.NewQueryClient(c.conn)

	for {
		request := &staking.QueryValidatorsRequest{Pagination: &query.PageRequest{Key: key}}

		stakingRes, err := client.Validators(ctx, request)
		if err != nil {
			return nil, endpointError(err.Error())
		}

		if stakingRes == nil {
			return nil, endpointError("got empty response from validators endpoint")
		}

		for _, val := range stakingRes.GetValidators() {
			err = val.UnpackInterfaces(interfaceRegistry)
			if err != nil {
				return nil, endpointError(err.Error())
			}

			vals = append(vals, val)
		}

		page := stakingRes.GetPagination()
		if page == nil {
			break
		}

		key = page.GetNextKey()
		if len(key) == 0 {
			break
		}
	}

	log.Debug(fmt.Sprintf("Validators: %d", len(vals)))

	return vals, nil
}

func (c Client) valConsMap(vals []staking.Validator) (map[string]staking.Validator, error) {
	vMap := map[string]staking.Validator{}

	for _, val := range vals {
		addr, err := val.GetConsAddr()
		if err != nil {
			return nil, endpointError(err.Error())
		}

		consAddr, err := bech32.ConvertAndEncode(c.cfg.Prefix+valConsStr, sdk.ConsAddress(addr))
		if err != nil {
			return nil, endpointError(err.Error())
		}

		vMap[consAddr] = val
	}

	return vMap, nil
}

func SigningValidators(ctx context.Context, cfg config.Config) (sVals []types.Validator, err error) {
	client, err := NewClient(cfg)
	if err != nil {
		log.Error(err.Error())

		return []types.Validator{}, endpointError(err.Error())
	}

	sInfos, err := client.SignigInfos(ctx)

	defer func() {
		if tempErr := client.conn.Close(); tempErr != nil {
			if err != nil {
				err = endpointError(fmt.Errorf("%w: %s", err, tempErr).Error())
			} else {
				err = endpointError(tempErr.Error())
			}
		}
	}()

	if err != nil {
		log.Error(err.Error())

		return []types.Validator{}, endpointError(err.Error())
	}

	vals, err := client.Validators(ctx)
	if err != nil {
		log.Error(err.Error())

		return []types.Validator{}, endpointError(err.Error())
	}

	valsMap, err := client.valConsMap(vals)
	if err != nil {
		log.Error(err.Error())

		return []types.Validator{}, endpointError(err.Error())
	}

	for _, info := range sInfos {
		if _, ok := valsMap[info.Address]; !ok {
			log.Debug(fmt.Sprintf("Not in validators: %s", info.Address))
		}

		sVals = append(sVals, types.Validator{
			ConsAddress:     info.Address,
			OperatorAddress: valsMap[info.Address].OperatorAddress,
			Moniker:         valsMap[info.Address].Description.Moniker,
			MissedBlocks:    info.MissedBlocksCounter,
		})
	}

	return sVals, nil
}

func LatestBlockHeight(ctx context.Context, cfg config.Config) (height int64, err error) {
	client, err := NewClient(cfg)
	if err != nil {
		log.Error(err.Error())

		return 0, endpointError(err.Error())
	}

	request := &base.GetLatestBlockRequest{}
	baseClient := base.NewServiceClient(client.conn)

	blockResp, err := baseClient.GetLatestBlock(ctx, request)

	defer func() {
		if tempErr := client.conn.Close(); tempErr != nil {
			if err != nil {
				err = endpointError(fmt.Errorf("%w: %s", err, tempErr).Error())
			} else {
				err = endpointError(tempErr.Error())
			}
		}
	}()

	if err != nil {
		log.Error(err.Error())

		return 0, endpointError(err.Error())
	}

	height = blockResp.GetBlock().Header.Height
	log.Debug(fmt.Sprintf("Latest height: %d", height))

	return height, nil
}
