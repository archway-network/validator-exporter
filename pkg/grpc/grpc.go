package grpc

import (
	"context"
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

type Client struct {
	cfg       config.Config
	conn      *grpc.ClientConn
	connClose func()
}

func NewClient(cfg config.Config) (Client, error) {
	client := Client{
		cfg: cfg,
	}

	conn, err := cfg.GRPCConn()
	if err != nil {
		return Client{}, err
	}

	client.conn = conn
	client.connClose = func() {
		if err := conn.Close(); err != nil {
			log.Error(fmt.Sprintf("failed to close connection :%s", err))
		}
	}

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
			return nil, err
		}

		if slashRes == nil {
			return nil, fmt.Errorf("got empty response from signing infos endpoint")
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
			return nil, err
		}

		if stakingRes == nil {
			return nil, fmt.Errorf("got empty response from validators endpoint")
		}

		for _, val := range stakingRes.GetValidators() {
			err = val.UnpackInterfaces(interfaceRegistry)
			if err != nil {
				return nil, err
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
			return nil, err
		}

		consAddr, err := bech32.ConvertAndEncode(c.cfg.Prefix+valConsStr, sdk.ConsAddress(addr))
		if err != nil {
			return nil, err
		}

		vMap[consAddr] = val
	}

	return vMap, nil
}

func SigningValidators(ctx context.Context, cfg config.Config) ([]types.Validator, error) {
	sVals := []types.Validator{}

	client, err := NewClient(cfg)
	if err != nil {
		log.Error(err.Error())

		return []types.Validator{}, err
	}

	defer client.connClose()

	sInfos, err := client.SignigInfos(ctx)
	if err != nil {
		log.Error(err.Error())

		return []types.Validator{}, err
	}

	vals, err := client.Validators(ctx)
	if err != nil {
		log.Error(err.Error())

		return []types.Validator{}, err
	}

	valsMap, err := client.valConsMap(vals)
	if err != nil {
		log.Error(err.Error())

		return []types.Validator{}, err
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

func LatestBlockHeight(ctx context.Context, cfg config.Config) (int64, error) {
	client, err := NewClient(cfg)
	if err != nil {
		log.Error(err.Error())

		return 0, err
	}

	defer client.connClose()

	request := &base.GetLatestBlockRequest{}
	baseClient := base.NewServiceClient(client.conn)

	blockResp, err := baseClient.GetLatestBlock(ctx, request)
	if err != nil {
		log.Error(err.Error())

		return 0, err
	}

	height := blockResp.GetBlock().Header.Height
	log.Debug(fmt.Sprintf("Latest height: %d", height))

	return height, nil
}
