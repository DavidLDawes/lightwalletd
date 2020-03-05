package common

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"


	"github.com/davidldawes/lightwalletd/parser"
	"github.com/davidldawes/lightwalletd/walletrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// RawRequest points to the function to send a an RPC request to verusd;
// in production, it points to btcsuite/btcd/rpcclient/rawrequest.go:RawRequest();
// in unit tests it points to a function to mock RPCs to verusd (pending).
var RawRequest func(method string, params []json.RawMessage) (json.RawMessage, error)

// Sleep allows a request to time.Sleep() to be mocked for testing;
// in production, it points to the standard library time.Sleep();
// in unit tests it points to a mock function.
var Sleep func(d time.Duration)

// Log as a global variable simplifies logging
var Log *logrus.Entry

// Metrics per API: GetSaplingInfo
var (
	GetSaplingInfoProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_GetSaplingInfo_processed",
		Help: "The total number of GetLatestBlock calls",
	})
)

var (
	GetSaplingInfoRetries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_GetSaplingInfo_retries",
		Help: "The total number of GetSaplingInfo retries",
	})
)

var (
	GetSaplingInfoErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_GetSaplingInfo_errors",
		Help: "The total number of GetSaplingInfo calls that returned an error",
	})
)

// GetSaplingInfo returns the result of the getblockchaininfo RPC to verusd
func GetSaplingInfo() (int, int, string, string) {
	// This request must succeed or we can't go on; give verusd time to start up
	var f interface{}
	GetSaplingInfoProcessed.Inc()
	retryCount := 0
	for {
		result, rpcErr := RawRequest("getblockchaininfo", []json.RawMessage{})
		if rpcErr == nil {
			if retryCount > 0 {
				Log.Warn("getblockchaininfo RPC successful")
			}
			err := json.Unmarshal(result, &f)
			if err != nil {
				GetSaplingInfoErrors.Inc()
				Log.Fatalf("error parsing JSON getblockchaininfo response: %v", err)
			}
			break
		}
		GetSaplingInfoRetries.Inc()
		retryCount++
		if retryCount > 10 {
			GetSaplingInfoErrors.Inc()
			Log.WithFields(logrus.Fields{
				"timeouts": retryCount,
			}).Fatal("unable to issue getblockchaininfo RPC call to verusd node")
		}
		Log.WithFields(logrus.Fields{
			"error": rpcErr.Error(),
			"retry": retryCount,
		}).Warn("error with getblockchaininfo rpc, retrying...")
		Sleep(time.Duration(10+retryCount*5) * time.Second) // backoff
	}

	chainName := f.(map[string]interface{})["chain"].(string)

	upgradeJSON := f.(map[string]interface{})["upgrades"]
	saplingJSON := upgradeJSON.(map[string]interface{})["76b809bb"] // Sapling ID
	saplingHeight := saplingJSON.(map[string]interface{})["activationheight"].(float64)

	blockHeight := f.(map[string]interface{})["headers"].(float64)

	consensus := f.(map[string]interface{})["consensus"]

	branchID := consensus.(map[string]interface{})["nextblock"].(string)

	return int(saplingHeight), int(blockHeight), chainName, branchID
}

// Metrics per API: getBlockFromRPC
var (
	getBlockFromRPCProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_getBlockFromRPC_processed",
		Help: "The total number of getBlockFromRPC calls",
	})
)

var (
	getBlockFromRPCErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_getBlockFromRPC_Errors",
		Help: "The total number of getBlockFromRPC calls that returned an error",
	})
)

func getBlockFromRPC(height int) (*walletrpc.CompactBlock, error) {
	getBlockFromRPCProcessed.Inc()
	params := make([]json.RawMessage, 2)
	params[0] = json.RawMessage("\"" + strconv.Itoa(height) + "\"")
	params[1] = json.RawMessage("0") // non-verbose (raw hex)
	result, rpcErr := RawRequest("getblock", params)

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		getBlockFromRPCErrors.Inc()
		// Check to see if we are requesting a height the verusd doesn't have yet
		if (strings.Split(rpcErr.Error(), ":"))[0] == "-8" {
			return nil, nil
		}
		return nil, errors.Wrap(rpcErr, "error requesting block")
	}

	var blockDataHex string
	err := json.Unmarshal(result, &blockDataHex)
	if err != nil {
		getBlockFromRPCErrors.Inc()
		return nil, errors.Wrap(err, "error reading JSON response")
	}

	blockData, err := hex.DecodeString(blockDataHex)
	if err != nil {
		getBlockFromRPCErrors.Inc()
		return nil, errors.Wrap(err, "error decoding getblock output")
	}

	block := parser.NewBlock()
	rest, err := block.ParseFromSlice(blockData)
	if err != nil {
		getBlockFromRPCErrors.Inc()
		return nil, errors.Wrap(err, "error parsing block")
	}
	if len(rest) != 0 {
		getBlockFromRPCErrors.Inc()
		return nil, errors.New("received overlong message")
	}
	if block.GetHeight() != height {
		getBlockFromRPCErrors.Inc()
		return nil, errors.New("received unexpected height block")
	}

	return block.ToCompact(), nil
}

// BlockIngestor Metrics
var (
	BlockIngestorProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_BlockIngestor_processed",
		Help: "The total number of BlockIngestor calls",
	})
)

var (
	BlockIngestorRetries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_BlockIngestor_retries",
		Help: "The total number of BlockIngestor retries",
	})
)

var (
	BlockIngestorErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_BlockIngestor_Errors",
		Help: "The total number of BlockIngestor calls that returned an error",
	})
)

// BlockIngestor runs as a goroutine and polls verusd for new blocks, adding them
// to the cache. The repetition count, rep, is nonzero only for unit-testing.
func BlockIngestor(c *BlockCache, height int, rep int) {
	BlockIngestorProcessed.Inc()
	reorgCount := 0

	// Start listening for new blocks
	retryCount := 0
	waiting := false
	for i := 0; rep == 0 || i < rep; i++ {
		block, err := getBlockFromRPC(height)
		if block == nil || err != nil {
			if err != nil {
				Log.WithFields(logrus.Fields{
					"height": height,
					"error":  err,
				}).Warn("error with getblock rpc")
				BlockIngestorRetries.Inc()
				retryCount++
				if retryCount > 10 {
					BlockIngestorErrors.Inc()
					Log.WithFields(logrus.Fields{
						"timeouts": retryCount,
					}).Fatal("unable to issue RPC call to verusd node")
				}
			}
			// We're up to date in our polling; wait for a new block
			c.Sync()
			waiting = true
			Sleep(10 * time.Second)
			continue
		}
		retryCount = 0

		if waiting || (height%100) == 0 {
			Log.Info("Ingestor adding block to cache: ", height)
		}

		// Check for reorgs once we have inital block hash from startup
		if c.LatestHash != nil && !bytes.Equal(block.PrevHash, c.LatestHash) {
			// This must back up at least 1, but it's arbitrary, any value
			// will work; this is probably a good balance.
			height = c.Reorg(height - 2)
			reorgCount += 2
			if reorgCount > 100 {
				BlockIngestorErrors.Inc()
				Log.Fatal("Reorg exceeded max of 100 blocks! Help!")
			}
			Log.WithFields(logrus.Fields{
				"height": height,
				"hash":   displayHash(block.Hash),
				"phash":  displayHash(block.PrevHash),
				"reorg":  reorgCount,
			}).Warn("REORG")
			continue
		}
		if err := c.Add(block); err != nil {
			BlockIngestorErrors.Inc()
			Log.Fatal("Cache add failed:", err)
		}
		reorgCount = 0
		height++
	}
}

// Metrics per API: GetBlock
var (
	GetBlockProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_GetBlockProcessed",
		Help: "The total number of GetBlock calls",
	})
)

var (
	GetBlockTooNewErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_GetBlockTooNewErrors",
		Help: "The total number of GetBlock calls requesting a height above the current block height",
	})
)

var (
	GetBlockErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_GetBlockErrors",
		Help: "The total number of GetBlock calls that returned an error other than TooNew",
	})
)

// GetBlock returns the compact block at the requested height, first by querying
// the cache, then, if not found, will request the block from verusd. It returns
// nil if no block exists at this height.
func GetBlock(cache *BlockCache, height int) (*walletrpc.CompactBlock, error) {
	GetBlockProcessed.Inc()
	// First, check the cache to see if we have the block
	block := cache.Get(height)
	if block != nil {
		return block, nil
	}

	// Not in the cache, ask verusd
	block, err := getBlockFromRPC(height)
	if err != nil {
		GetBlockErrors.Inc()
		return nil, err
	}
	if block == nil {
		// Block height is too large
		GetBlockTooNewErrors.Inc()
		return nil, errors.New("block requested is newer than latest block")
	}
	return block, nil
}

// Metrics per API: GetBlockRange
var (
	GetBlockRangeProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_GetBlockRangeProcessed",
		Help: "The total number of GetBlockRange calls",
	})
)

var (
	GetBlockRangeErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "common_GetBlockRangeErrors",
		Help: "The total number of GetBlockRange calls that returned an error other than TooNew",
	})
)
// GetBlockRange returns a sequence of consecutive blocks in the given range.
func GetBlockRange(cache *BlockCache, blockOut chan<- walletrpc.CompactBlock, errOut chan<- error, start, end int) {
	GetBlockRangeProcessed.Inc()
	// Go over [start, end] inclusive
	for i := start; i <= end; i++ {
		block, err := GetBlock(cache, i)
		if err != nil {
			GetBlockRangeErrors.Inc()
			errOut <- err
			return
		}
		blockOut <- *block
	}
	errOut <- nil
}

func displayHash(hash []byte) string {
	rhash := make([]byte, len(hash))
	copy(rhash, hash)
	// Reverse byte order
	for i := 0; i < len(rhash)/2; i++ {
		j := len(rhash) - 1 - i
		rhash[i], rhash[j] = rhash[j], rhash[i]
	}
	return hex.EncodeToString(rhash)
}
