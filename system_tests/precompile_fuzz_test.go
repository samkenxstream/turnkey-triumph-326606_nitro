// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package arbtest

import (
	"math/big"
	"testing"

	"github.com/tenderly/nitro/go-ethereum/common"
	"github.com/tenderly/nitro/go-ethereum/core"
	"github.com/tenderly/nitro/go-ethereum/core/rawdb"
	"github.com/tenderly/nitro/go-ethereum/core/state"
	"github.com/tenderly/nitro/go-ethereum/core/types"
	"github.com/tenderly/nitro/go-ethereum/core/vm"
	"github.com/tenderly/nitro/go-ethereum/params"
	"github.com/tenderly/nitro/arbos/arbosState"
	"github.com/tenderly/nitro/arbos/burn"
	"github.com/tenderly/nitro/arbstate"
	"github.com/tenderly/nitro/precompiles"
)

const fuzzGas uint64 = 1200000

func FuzzPrecompiles(f *testing.F) {
	arbstate.RequireHookedGeth()

	f.Fuzz(func(t *testing.T, precompileSelector byte, methodSelector byte, input []byte) {
		// Create a StateDB
		sdb, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		if err != nil {
			panic(err)
		}
		burner := burn.NewSystemBurner(nil, false)
		_, err = arbosState.InitializeArbosState(sdb, burner, params.ArbitrumDevTestChainConfig())
		if err != nil {
			panic(err)
		}

		// Create an EVM
		gp := core.GasPool(fuzzGas)
		txContext := vm.TxContext{
			GasPrice: common.Big1,
		}
		blockContext := vm.BlockContext{
			CanTransfer: core.CanTransfer,
			Transfer:    core.Transfer,
			GetHash:     nil,
			Coinbase:    common.Address{},
			BlockNumber: new(big.Int),
			Time:        new(big.Int),
			Difficulty:  new(big.Int),
			GasLimit:    fuzzGas,
			BaseFee:     common.Big1,
		}
		evm := vm.NewEVM(blockContext, txContext, sdb, params.ArbitrumDevTestChainConfig(), vm.Config{})

		// Pick a precompile address based on the first byte of the input
		var addr common.Address
		addr[19] = precompileSelector

		// Pick a precompile method based on the second byte of the input
		if precompile := precompiles.Precompiles()[addr]; precompile != nil {
			sigs := precompile.Precompile().Get4ByteMethodSignatures()
			if int(methodSelector) < len(sigs) {
				newInput := make([]byte, 4)
				copy(newInput, sigs[methodSelector][:])
				newInput = append(newInput, input...)
				input = newInput
			}
		}

		// Create and apply a message
		msg := types.NewMessage(
			common.Address{},
			&addr,
			0,
			new(big.Int),
			fuzzGas,
			new(big.Int),
			new(big.Int),
			new(big.Int),
			input,
			nil,
			true,
		)
		_, _ = core.ApplyMessage(evm, msg, &gp)
	})
}
