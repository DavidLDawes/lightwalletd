// Package walletrpc wraps the verusd RPC endpoint
// Copyright (c) 2019-2020 The Zcash developers
// Forked and modified for the VerusCoin chain
// Copyright 2020 the VerusCoin developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package walletrpc

//go:generate protoc -I . ./compact_formats.proto --go_out=plugins=grpc:.
//go:generate protoc -I . ./service.proto --go_out=plugins=grpc:.
