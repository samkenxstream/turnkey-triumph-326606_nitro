// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package burn

import (
	glog "github.com/tenderly/nitro/go-ethereum/log"
	"github.com/tenderly/nitro/arbos/util"
)

type Burner interface {
	Burn(amount uint64) error
	Burned() uint64
	Restrict(err error)
	ReadOnly() bool
	TracingInfo() *util.TracingInfo
}

type SystemBurner struct {
	gasBurnt    uint64
	tracingInfo *util.TracingInfo
	readOnly    bool
}

func NewSystemBurner(tracingInfo *util.TracingInfo, readOnly bool) *SystemBurner {
	return &SystemBurner{
		tracingInfo: tracingInfo,
		readOnly:    readOnly,
	}
}

func (burner *SystemBurner) Burn(amount uint64) error {
	burner.gasBurnt += amount
	return nil
}

func (burner *SystemBurner) Burned() uint64 {
	return burner.gasBurnt
}

func (burner *SystemBurner) Restrict(err error) {
	if err != nil {
		glog.Error("Restrict() received an error", "err", err)
	}
}

func (burner *SystemBurner) ReadOnly() bool {
	return burner.readOnly
}

func (burner *SystemBurner) TracingInfo() *util.TracingInfo {
	return burner.tracingInfo
}
