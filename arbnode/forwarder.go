// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package arbnode

import (
	"context"

	"github.com/tenderly/nitro/go-ethereum/core/types"
	"github.com/tenderly/nitro/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type TxForwarder struct {
	target string
	client *ethclient.Client
}

func NewForwarder(target string) *TxForwarder {
	return &TxForwarder{
		target: target,
	}
}

func (f *TxForwarder) PublishTransaction(ctx context.Context, tx *types.Transaction) error {
	if f.client == nil {
		return errors.New("sequencer temporarily unavailable")
	}
	return f.client.SendTransaction(ctx, tx)
}

func (f *TxForwarder) Initialize(ctx context.Context) error {
	if f.target == "" {
		f.client = nil
		return nil
	}
	client, err := ethclient.DialContext(ctx, f.target)
	if err != nil {
		return err
	}
	f.client = client
	return nil
}

func (f *TxForwarder) Start(ctx context.Context) error {
	return nil
}

func (f *TxForwarder) StopAndWait() {}

type TxDropper struct{}

func NewTxDropper() *TxDropper {
	return &TxDropper{}
}

func (f *TxDropper) PublishTransaction(ctx context.Context, tx *types.Transaction) error {
	return errors.New("transactions not supported by this endpoint")
}

func (f *TxDropper) Initialize(ctx context.Context) error { return nil }

func (f *TxDropper) Start(ctx context.Context) error { return nil }

func (f *TxDropper) StopAndWait() {}
