// Package Common Copyright (c) 2019-2020 The Zcash developers
// Copyright (c) 2020 The VerusCoin Developers and David L. Dawes
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .

package common

import (
	"encoding/json"
	"strconv"

	"github.com/Asherda/lightwalletd/walletrpc"
	"github.com/pkg/errors"
)

// GetIdentity requests info about an identity from verusd and deserialzies the JSON
// result (if any). This is used by the GRPC GetIdentity endpoint.
func GetIdentity(request *walletrpc.GetIdentityRequest) (*walletrpc.GetIdentityResponse, error) {

	params := make([]json.RawMessage, 1)
	params[0] = json.RawMessage("\"" + request.GetIdentity() + "\"")
	result, rpcErr := RawRequest("getidentity", params)

	if rpcErr != nil {
		return nil, rpcErr
	}

	response := &walletrpc.GetIdentityResponse{}
	err := json.Unmarshal(result, &response.Identityinfo)

	if err != nil {
		return nil, errors.Wrap(err, "error reading JSON response")
	}

	return response, nil
}

// VerifyMessage is used to verify that a message hash was created by the specified identity.
func VerifyMessage(request *walletrpc.VerifyMessageRequest) (*walletrpc.VerifyMessageResponse, error) {
	params := make([]json.RawMessage, 4)
	params[0] = json.RawMessage("\"" + request.Signer + "\"")
	params[1] = json.RawMessage("\"" + request.Signature + "\"")
	params[2] = json.RawMessage("\"" + request.Message + "\"")
	params[3] = json.RawMessage("\"" + strconv.FormatBool(request.Checklatest) + "\"")

	result, rpcErr := RawRequest("verifymessage", params)

	if rpcErr != nil {
		return nil, rpcErr
	}

	var signatureisvalid bool
	err := json.Unmarshal(result, &signatureisvalid)

	if err != nil {
		return nil, errors.Wrap(err, "error reading JSON response")
	}

	return &walletrpc.VerifyMessageResponse{
		Signatureisvalid: signatureisvalid,
	}, err
}

// VerifyHash is used to verify that a hash was created by the specified identity.
func VerifyHash(request *walletrpc.VerifyHashRequest) (*walletrpc.VerifyHashResponse, error) {
	params := make([]json.RawMessage, 4)
	params[0] = json.RawMessage("\"" + request.Signer + "\"")
	params[1] = json.RawMessage("\"" + request.Signature + "\"")
	params[2] = json.RawMessage("\"" + request.Hash + "\"")
	params[3] = json.RawMessage("\"" + strconv.FormatBool(request.Checklatest) + "\"")

	result, rpcErr := RawRequest("verifyhash", params)

	if rpcErr != nil {
		return nil, rpcErr
	}

	var signatureisvalid bool
	err := json.Unmarshal(result, &signatureisvalid)

	if err != nil {
		return nil, errors.Wrap(err, "error reading JSON response")
	}

	return &walletrpc.VerifyHashResponse{
		Signatureisvalid: signatureisvalid,
	}, err
}
