// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package precompiles

import (
	"testing"

	"github.com/tenderly/nitro/go-ethereum/common/math"

	"github.com/tenderly/nitro/arbos/arbosState"
	"github.com/tenderly/nitro/arbos/burn"
	"github.com/tenderly/nitro/util/testhelpers"

	"github.com/tenderly/nitro/go-ethereum/common"
	"github.com/tenderly/nitro/go-ethereum/core/types"
	"github.com/tenderly/nitro/go-ethereum/crypto"
	"github.com/tenderly/nitro/arbos/util"
)

func TestArbOwner(t *testing.T) {
	evm := newMockEVMForTesting()
	caller := common.BytesToAddress(crypto.Keccak256([]byte{})[:20])
	tracer := util.NewTracingInfo(evm, testhelpers.RandomAddress(), types.ArbosAddress, util.TracingDuringEVM)
	state, err := arbosState.OpenArbosState(evm.StateDB, burn.NewSystemBurner(tracer, false))
	Require(t, err)
	Require(t, state.ChainOwners().Add(caller))

	addr1 := common.BytesToAddress(crypto.Keccak256([]byte{1})[:20])
	addr2 := common.BytesToAddress(crypto.Keccak256([]byte{2})[:20])
	addr3 := common.BytesToAddress(crypto.Keccak256([]byte{3})[:20])

	prec := &ArbOwner{}
	gasInfo := &ArbGasInfo{}
	callCtx := testContext(caller, evm)

	// the zero address is an owner by default
	Require(t, prec.RemoveChainOwner(callCtx, evm, common.Address{}))

	Require(t, prec.AddChainOwner(callCtx, evm, addr1))
	Require(t, prec.AddChainOwner(callCtx, evm, addr2))
	Require(t, prec.AddChainOwner(callCtx, evm, addr1))

	member, err := prec.IsChainOwner(callCtx, evm, addr1)
	Require(t, err)
	if !member {
		Fail(t)
	}

	member, err = prec.IsChainOwner(callCtx, evm, addr2)
	Require(t, err)
	if !member {
		Fail(t)
	}

	member, err = prec.IsChainOwner(callCtx, evm, addr3)
	Require(t, err)
	if member {
		Fail(t)
	}

	Require(t, prec.RemoveChainOwner(callCtx, evm, addr1))
	member, err = prec.IsChainOwner(callCtx, evm, addr1)
	Require(t, err)
	if member {
		Fail(t)
	}
	member, err = prec.IsChainOwner(callCtx, evm, addr2)
	Require(t, err)
	if !member {
		Fail(t)
	}

	Require(t, prec.AddChainOwner(callCtx, evm, addr1))
	all, err := prec.GetAllChainOwners(callCtx, evm)
	Require(t, err)
	if len(all) != 3 {
		Fail(t)
	}
	if all[0] == all[1] || all[1] == all[2] || all[0] == all[2] {
		Fail(t)
	}
	if all[0] != addr1 && all[1] != addr1 && all[2] != addr1 {
		Fail(t)
	}
	if all[0] != addr2 && all[1] != addr2 && all[2] != addr2 {
		Fail(t)
	}
	if all[0] != caller && all[1] != caller && all[2] != caller {
		Fail(t)
	}

	costCap, err := gasInfo.GetAmortizedCostCapBips(callCtx, evm)
	Require(t, err)
	if costCap != math.MaxUint64 {
		Fail(t, costCap)
	}
	newCostCap := uint64(77734)
	Require(t, prec.SetAmortizedCostCapBips(callCtx, evm, newCostCap))
	costCap, err = gasInfo.GetAmortizedCostCapBips(callCtx, evm)
	Require(t, err)
	if costCap != newCostCap {
		Fail(t)
	}
}
