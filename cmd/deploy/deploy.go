// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/tenderly/nitro/cmd/genericconf"
	"github.com/tenderly/nitro/util/headerreader"
	"github.com/tenderly/nitro/validator"

	"github.com/tenderly/nitro/go-ethereum/common"
	"github.com/tenderly/nitro/go-ethereum/ethclient"
	"github.com/tenderly/nitro/go-ethereum/log"
	"github.com/tenderly/nitro/go-ethereum/params"
	"github.com/tenderly/nitro/arbnode"
	"github.com/tenderly/nitro/cmd/util"
)

func main() {
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.LvlDebug)
	log.Root().SetHandler(glogger)
	log.Info("deploying rollup")

	ctx := context.Background()

	l1conn := flag.String("l1conn", "", "l1 connection")
	l1keystore := flag.String("l1keystore", "", "l1 private key store")
	deployAccount := flag.String("l1DeployAccount", "", "l1 seq account to use (default is first account in keystore)")
	ownerAddressString := flag.String("ownerAddress", "", "the rollup owner's address")
	sequencerAddressString := flag.String("sequencerAddress", "", "the sequencer's address")
	wasmmoduleroot := flag.String("wasmmoduleroot", "", "WASM module root hash")
	wasmrootpath := flag.String("wasmrootpath", "", "path to machine folders")
	l1passphrase := flag.String("l1passphrase", "passphrase", "l1 private key file passphrase")
	outfile := flag.String("l1deployment", "deploy.json", "deployment output json file")
	l1ChainIdUint := flag.Uint64("l1chainid", 1337, "L1 chain ID")
	l2ChainIdUint := flag.Uint64("l2chainid", params.ArbitrumDevTestChainConfig().ChainID.Uint64(), "L2 chain ID")
	authorizevalidators := flag.Uint64("authorizevalidators", 0, "Number of validators to preemptively authorize")
	txTimeout := flag.Duration("txtimeout", 10*time.Minute, "Timeout when waiting for a transaction to be included in a block")
	flag.Parse()
	l1ChainId := new(big.Int).SetUint64(*l1ChainIdUint)
	l2ChainId := new(big.Int).SetUint64(*l2ChainIdUint)

	wallet := genericconf.WalletConfig{
		Pathname:     *l1keystore,
		Account:      *deployAccount,
		PasswordImpl: *l1passphrase,
	}
	l1TransactionOpts, err := util.GetTransactOptsFromWallet(&wallet, l1ChainId)
	if err != nil {
		flag.Usage()
		log.Error("error reading keystore")
		panic(err)
	}

	l1client, err := ethclient.Dial(*l1conn)
	if err != nil {
		flag.Usage()
		log.Error("error creating l1client")
		panic(err)
	}

	if !common.IsHexAddress(*sequencerAddressString) && len(*sequencerAddressString) > 0 {
		panic("specified sequencer address is invalid")
	}
	if !common.IsHexAddress(*ownerAddressString) {
		panic("please specify a valid rollup owner address")
	}
	sequencerAddress := common.HexToAddress(*sequencerAddressString)
	ownerAddress := common.HexToAddress(*ownerAddressString)
	if sequencerAddress != (common.Address{}) && ownerAddress != l1TransactionOpts.From {
		panic("cannot specify sequencer address if owner is not deployer")
	}

	machineConfig := validator.DefaultNitroMachineConfig
	machineConfig.RootPath = *wasmrootpath

	headerReaderConfig := headerreader.DefaultConfig
	headerReaderConfig.TxTimeout = *txTimeout

	deployPtr, err := arbnode.DeployOnL1(
		ctx,
		l1client,
		l1TransactionOpts,
		sequencerAddress,
		ownerAddress,
		*authorizevalidators,
		common.HexToHash(*wasmmoduleroot),
		l2ChainId,
		headerReaderConfig,
		machineConfig,
	)
	if err != nil {
		flag.Usage()
		log.Error("error deploying on l1")
		panic(err)
	}
	deployData, err := json.Marshal(deployPtr)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(*outfile, deployData, 0600); err != nil {
		panic(err)
	}
}
