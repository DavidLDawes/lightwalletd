// Package frontend handles block header and transaction related functions
// Copyright (c) 2019-2020 The Zcash developers
// Forked and modified for the VerusCoin chain
// Copyright 2020 the VerusCoin developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package frontend

import (
	"github.com/btcsuite/btcd/rpcclient"
)

// NewVRPCFromConf connect to the VerusCoin RPC endpoint
func NewVRPCFromConf(chainName string, bindAddr string, user string, password string) (*rpcclient.Client, error) {

	if len(chainName) < 1 {
		chainName = "VRSC"
	}
	if len(bindAddr) < 1 {
		// probably ought to error out on startup for this case
		bindAddr = "127.0.0.1:27486"
	}
	if len(user) < 1 {
		user = "RPCUSER"
	}

	// Connect to local verusd RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         bindAddr,
		User:         user,
		Pass:         password,
		HTTPPostMode: true, // verusd only supports HTTP POST mode
		DisableTLS:   true, // verusd does not provide TLS by default - wait, what?
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	return rpcclient.New(connCfg, nil)
}
