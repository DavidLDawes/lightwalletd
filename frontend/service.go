package frontend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/asherda/lightwalletd/common"
	"github.com/asherda/lightwalletd/walletrpc"
)

var (
	ErrUnspecified = errors.New("request for unspecified identifier")
)

type LwdStreamer struct {
	cache *common.BlockCache
}

func NewLwdStreamer(cache *common.BlockCache) (walletrpc.CompactTxStreamerServer, error) {
	return &LwdStreamer{cache}, nil
}

// Metrics per API: GetLatestBlock
var (
        getLatestBlocksProcessed = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetLatestBlock",
                Help: "The total number of GetLatestBlock calls",
        })
)

var (
        getLatestBlocksErrors = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetLatestBlockErrors",
                Help: "The total number of GetLatestBlock calls that returned an error",
        })
)

// GetLatestBlock returns the height of the best chain, according to zcashd.
func (s *LwdStreamer) GetLatestBlock(ctx context.Context, placeholder *walletrpc.ChainSpec) (*walletrpc.BlockID, error) {
	getLatestBlocksProcessed.Inc()
	latestBlock := s.cache.GetLatestHeight()

	if latestBlock == -1 {
		getLatestBlocksErrors.Inc()
		return nil, errors.New("Cache is empty. Server is probably not yet ready")
	}

	// TODO: also return block hashes here
	return &walletrpc.BlockID{Height: uint64(latestBlock)}, nil
}

// Metrics per API: GetAddressTxids
var (
        getAddressTxidsProcessed = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetAddressTxids",
                Help: "Total number of GetAddressTxids calls",
        })
)

var (
        getAddressTxidsInvalidAddressErrors = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetAddressTxidsInvalidAddressErrors",
                Help: "The total number of GetAddressTxids calls that returned an invalid address error",
        })
)

var (
        getAddressTxidsOtherErrors = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetAddressTxidsErrors",
                Help: "The total number of GetAddressTxids calls that returned an error, excluding invalid address errors",
        })
)

// GetAddressTxids is a streaming RPC that returns transaction IDs that have
// the given transparent address (taddr) as either an input or output.
func (s *LwdStreamer) GetAddressTxids(addressBlockFilter *walletrpc.TransparentAddressBlockFilter, resp walletrpc.CompactTxStreamer_GetAddressTxidsServer) error {
	getAddressTxidsProcessed.Inc()
	// Test to make sure Address is a single t address
	match, err := regexp.Match("\\At[a-zA-Z0-9]{34}\\z", []byte(addressBlockFilter.Address))
	if err != nil || !match {
		getAddressTxidsInvalidAddressErrors.Inc()
		common.Log.Error("Invalid address:", addressBlockFilter.Address)
		return errors.New("Invalid address")
	}

	params := make([]json.RawMessage, 1)
	st := "{\"addresses\": [\"" + addressBlockFilter.Address + "\"]," +
		"\"start\": " + strconv.FormatUint(addressBlockFilter.Range.Start.Height, 10) +
		", \"end\": " + strconv.FormatUint(addressBlockFilter.Range.End.Height, 10) + "}"

	params[0] = json.RawMessage(st)

	result, rpcErr := common.RawRequest("getaddresstxids", params)

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		getAddressTxidsOtherErrors.Inc()
		common.Log.Errorf("GetAddressTxids error: %s", rpcErr.Error())
		return err
	}

	var txids []string
	err = json.Unmarshal(result, &txids)
	if err != nil {
		getAddressTxidsOtherErrors.Inc()
		common.Log.Errorf("GetAddressTxids error: %s", err.Error())
		return err
	}

	timeout, cancel := context.WithTimeout(resp.Context(), 30*time.Second)
	defer cancel()

	for _, txidstr := range txids {
		txid, _ := hex.DecodeString(txidstr)
		// Txid is read as a string, which is in big-endian order. But when converting
		// to bytes, it should be little-endian
		for left, right := 0, len(txid)-1; left < right; left, right = left+1, right-1 {
			txid[left], txid[right] = txid[right], txid[left]
		}
		tx, err := s.GetTransaction(timeout, &walletrpc.TxFilter{Hash: txid})
		if err != nil {
			getAddressTxidsOtherErrors.Inc()
			common.Log.Errorf("GetTransaction error: %s", err.Error())
			return err
		}
		if err = resp.Send(tx); err != nil {
			getAddressTxidsOtherErrors.Inc()
			return err
		}
	}
	return nil
}

// Metrics per API: getBlock
var (
        getBlocks = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetBlock",
                Help: "The total number of GetBlock calls",
        })
)

var (
        getBlockErrors = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetBlockErrors",
                Help: "The total number of GetBlock calls that returned an error",
        })
)

// GetBlock returns the compact block at the requested height. Requesting a
// block by hash is not yet supported.
func (s *LwdStreamer) GetBlock(ctx context.Context, id *walletrpc.BlockID) (*walletrpc.CompactBlock, error) {
	getBlocks.Inc()
	if id.Height == 0 && id.Hash == nil {
		return nil, ErrUnspecified
	}

	// Precedence: a hash is more specific than a height. If we have it, use it first.
	if id.Hash != nil {
		getBlockErrors.Inc()
		// TODO: Get block by hash
		return nil, errors.New("GetBlock by Hash is not yet implemented")
	}
	cBlock, err := common.GetBlock(s.cache, int(id.Height))

	if err != nil {
		getBlockErrors.Inc()
		return nil, err
	}

	return cBlock, err
}

// Metrics per API: GetBlockRange
var (
        getBlockRanges = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetBlockRange",
                Help: "The total number of GetBlockRange calls",
        })
)

var (
        getBlockRangeErrors = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetBlockRangeErrors",
                Help: "The total number of GetBlockRange calls that returned an error",
        })
)

// GetBlockRange is a streaming RPC that returns blocks, in compact form,
// (as also returned by GetBlock) from the block height 'start' to height
// 'end' inclusively.
func (s *LwdStreamer) GetBlockRange(span *walletrpc.BlockRange, resp walletrpc.CompactTxStreamer_GetBlockRangeServer) error {
	getBlockRanges.Inc()
	blockChan := make(chan walletrpc.CompactBlock)
	errChan := make(chan error)

	go common.GetBlockRange(s.cache, blockChan, errChan, int(span.Start.Height), int(span.End.Height))

	for {
		select {
		case err := <-errChan:
			// this will also catch context.DeadlineExceeded from the timeout
			getBlockRangeErrors.Inc()
			return err
		case cBlock := <-blockChan:
			err := resp.Send(&cBlock)
			if err != nil {
				getBlockRangeErrors.Inc()
				return err
			}
		}
	}
}

// Metrics per API: GetTransaction
var (
        getTransactions = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetTransaction",
                Help: "The total number of GetTransaction calls",
        })
)

var (
        getTransactionErrors = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetTransactionErrors",
                Help: "The total number of GetTransaction calls that returned an error",
        })
)

// GetTransaction returns the raw transaction bytes that are returned
// by the zcashd 'getrawtransaction' RPC.
func (s *LwdStreamer) GetTransaction(ctx context.Context, txf *walletrpc.TxFilter) (*walletrpc.RawTransaction, error) {
	getTransactions.Inc()
	if txf.Hash != nil {
		txid := txf.Hash
		for left, right := 0, len(txid)-1; left < right; left, right = left+1, right-1 {
			txid[left], txid[right] = txid[right], txid[left]
		}
		leHashString := hex.EncodeToString(txid)

		params := []json.RawMessage{
			json.RawMessage("\"" + leHashString + "\""),
			json.RawMessage("1"),
		}

		result, rpcErr := common.RawRequest("getrawtransaction", params)

		// For some reason, the error responses are not JSON
		if rpcErr != nil {
			getTransactionErrors.Inc()
			common.Log.Errorf("GetTransaction error: %s", rpcErr.Error())
			return nil, errors.New((strings.Split(rpcErr.Error(), ":"))[0])
		}
		var txinfo interface{}
		err := json.Unmarshal(result, &txinfo)
		if err != nil {
			getTransactionErrors.Inc()
			return nil, err
		}
		txBytes := txinfo.(map[string]interface{})["hex"].(string)
		txHeight := txinfo.(map[string]interface{})["height"].(float64)
		return &walletrpc.RawTransaction{Data: []byte(txBytes), Height: uint64(txHeight)}, nil
	}

	if txf.Block != nil && txf.Block.Hash != nil {
		getTransactionErrors.Inc()
		common.Log.Error("Can't GetTransaction with a blockhash+num. Please call GetTransaction with txid")
		return nil, errors.New("Can't GetTransaction with a blockhash+num. Please call GetTransaction with txid")
	}
	getTransactionErrors.Inc()
	common.Log.Error("Please call GetTransaction with txid")
	return nil, errors.New("Please call GetTransaction with txid")
}

// Metrics per API: getLightdInfo
var (
        getLightdInfos = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_GetLightdInfo",
                Help: "The total number of GetLightdInfo calls",
        })
)

// GetLightdInfo gets the LightWalletD (this server) info, and includes information
// it gets from its backend verusd.
func (s *LwdStreamer) GetLightdInfo(ctx context.Context, in *walletrpc.Empty) (*walletrpc.LightdInfo, error) {
	getLightdInfos.Inc()
	saplingHeight, blockHeight, chainName, consensusBranchId := common.GetSaplingInfo()

	return &walletrpc.LightdInfo{
		Version:                 "0.2.1",
		Vendor:                  "ECC LightWalletD",
		TaddrSupport:            true,
		ChainName:               chainName,
		SaplingActivationHeight: uint64(saplingHeight),
		ConsensusBranchId:       consensusBranchId,
		BlockHeight:             uint64(blockHeight),
	}, nil
}

// Metrics per API: SendTransaction
var (
        sendTransactions = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_SendTransaction",
                Help: "The total number of SendTransaction calls",
        })
)

var (
        sendTransactionSoftErrors = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_SendTransactionErrors",
                Help: "The total number of SendTransaction calls that returned an error",
        })
)

var (
        sendTransactionCriticalErrors = promauto.NewCounter(prometheus.CounterOpts{
                Name: "service_SendTransactionCriticalErrors",
                Help: "The total number of SendTransaction calls that returned a critical error",
        })
)

// SendTransaction forwards raw transaction bytes to a zcashd instance over JSON-RPC
func (s *LwdStreamer) SendTransaction(ctx context.Context, rawtx *walletrpc.RawTransaction) (*walletrpc.SendResponse, error) {
	// sendrawtransaction "hexstring" ( allowhighfees )
	//
	// Submits raw transaction (serialized, hex-encoded) to local node and network.
	//
	// Also see createrawtransaction and signrawtransaction calls.
	//
	// Arguments:
	// 1. "hexstring"    (string, required) The hex string of the raw transaction)
	// 2. allowhighfees    (boolean, optional, default=false) Allow high fees
	//
	// Result:
	// "hex"             (string) The transaction hash in hex

	// Metrics first
	sendTransactions.Inc()

	// Construct raw JSON-RPC params
	params := make([]json.RawMessage, 1)
	txHexString := hex.EncodeToString(rawtx.Data)
	params[0] = json.RawMessage("\"" + txHexString + "\"")
	result, rpcErr := common.RawRequest("sendrawtransaction", params)

	var err error
	var errCode int64
	var errMsg string

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		errParts := strings.SplitN(rpcErr.Error(), ":", 2)
		errMsg = strings.TrimSpace(errParts[1])
		errCode, err = strconv.ParseInt(errParts[0], 10, 32)
		if err != nil {
			sendTransactionCriticalErrors.Inc()
			// This should never happen. We can't panic here, but it's that class of error.
			// This is why we need integration testing to work better than regtest currently does. TODO.
			return nil, errors.New("SendTransaction couldn't parse error code")
		} else {
			sendTransactionSoftErrors.Inc()
		}
	} else {
		sendTransactionSoftErrors.Inc()
		errMsg = string(result)
	}

	// TODO these are called Error but they aren't at the moment.
	// A success will return code 0 and message txhash.
	return &walletrpc.SendResponse{
		ErrorCode:    int32(errCode),
		ErrorMessage: errMsg,
	}, nil
}
