package grpc

import (
	"testing"

	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
)

func TestBondStatus(t *testing.T) {
	status := staking.Bonded

	exp := "bonded"
	res := bondStatus(status)

	assert.Equal(t, exp, res)
}
