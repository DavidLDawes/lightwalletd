// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package common

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/asherda/lightwalletd/parser"
	"github.com/asherda/lightwalletd/walletrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// 'make build' will overwrite this string with the output of git-describe (tag)
var Version = "v0.0.0.0-dev"
var GitCommit = ""
var BuildDate = ""
var BuildUser = ""

type Options struct {
	GRPCBindAddr      string `json:"grpc_bind_address,omitempty"`
	HTTPBindAddr      string `json:"http_bind_address,omitempty"`
	TLSCertPath       string `json:"tls_cert_path,omitempty"`
	TLSKeyPath        string `json:"tls_cert_key,omitempty"`
	LogLevel          uint64 `json:"log_level,omitempty"`
	LogFile           string `json:"log_file,omitempty"`
	VerusdConfPath    string `json:"verusd_conf,omitempty"`
	ZcashConfPath     string `json:"zcash_conf,omitempty"`
	NoTLSVeryInsecure bool   `json:"no_tls_very_insecure,omitempty"`
	Redownload        bool   `json:"redownload"`
	DataDir           string `json:"data-dir"`
	Darkside          bool   `json:"darkside"`
}

// RawRequest points to the function to send a an RPC request to zcashd;
// in production, it points to btcsuite/btcd/rpcclient/rawrequest.go:RawRequest();
// in unit tests it points to a function to mock RPCs to zcashd.
var RawRequest func(method string, params []json.RawMessage) (json.RawMessage, error)

// Sleep allows a request to time.Sleep() to be mocked for testing;
// in production, it points to the standard library time.Sleep();
// in unit tests it points to a mock function.
var Sleep func(d time.Duration)

// Log as a global variable simplifies logging
var Log *logrus.Entry

// GetSaplingInfo returns the result of the getblockchaininfo RPC to zcashd
func GetSaplingInfo() (int, int, string, string) {
	// This request must succeed or we can't go on; give zcashd time to start up
	var f interface{}
	retryCount := 0
	for {
		result, rpcErr := RawRequest("getblockchaininfo", []json.RawMessage{})
		if rpcErr == nil {
			if retryCount > 0 {
				Log.Warn("getblockchaininfo RPC successful")
			}
			err := json.Unmarshal(result, &f)
			if err != nil {
				Log.Fatalf("error parsing JSON getblockchaininfo response: %v", err)
			}
			break
		}
		retryCount++
		if retryCount > 10 {
			Log.WithFields(logrus.Fields{
				"timeouts": retryCount,
			}).Fatal("unable to issue getblockchaininfo RPC call to zcashd node")
		}
		Log.WithFields(logrus.Fields{
			"error": rpcErr.Error(),
			"retry": retryCount,
		}).Warn("error with getblockchaininfo rpc, retrying...")
		Sleep(time.Duration(10+retryCount*5) * time.Second) // backoff
	}

	chainName := f.(map[string]interface{})["chain"].(string)

	upgradeJSON := f.(map[string]interface{})["upgrades"]

	// If the sapling consensus branch doesn't exist, it must be regtest
	saplingHeight := float64(0)
	if saplingJSON, ok := upgradeJSON.(map[string]interface{})["76b809bb"]; ok { // Sapling ID
		saplingHeight = saplingJSON.(map[string]interface{})["activationheight"].(float64)
	}

	blockHeight := f.(map[string]interface{})["headers"].(float64)

	consensus := f.(map[string]interface{})["consensus"]

	branchID := consensus.(map[string]interface{})["nextblock"].(string)

	return int(saplingHeight), int(blockHeight), chainName, branchID
}

func getBlockFromRPC(height int) (*walletrpc.CompactBlock, error) {
	params := make([]json.RawMessage, 2)
	params[0] = json.RawMessage("\"" + strconv.Itoa(height) + "\"")
	params[1] = json.RawMessage("0") // non-verbose (raw hex)
	result, rpcErr := RawRequest("getblock", params)

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		// Check to see if we are requesting a height the zcashd doesn't have yet
		if (strings.Split(rpcErr.Error(), ":"))[0] == "-8" {
			return nil, nil
		}
		return nil, errors.Wrap(rpcErr, "error requesting block")
	}

	var blockDataHex string
	err := json.Unmarshal(result, &blockDataHex)
	if err != nil {
		return nil, errors.Wrap(err, "error reading JSON response")
	}

	blockData, err := hex.DecodeString(blockDataHex)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding getblock output")
	}

	block := parser.NewBlock()
	rest, err := block.ParseFromSlice(blockData)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing block")
	}
	if len(rest) != 0 {
		return nil, errors.New("received overlong message")
	}

	// TODO COINBASE-HEIGHT: restore this check after coinbase height is fixed
	if false && block.GetHeight() != height {
		return nil, errors.New("received unexpected height block")
	}

	return block.ToCompact(), nil
}

// BlockIngestor runs as a goroutine and polls zcashd for new blocks, adding
//  them to the cache. If conn is passed in then use it to store a copy of
//  the data as we ingest it. The repetition count, rep, is nonzero only for unit-testing.
func BlockIngestor(c *BlockCache, db *sql.DB, rep int) {
	lastLog := time.Now()
	reorgCount := 0
	lastHeightLogged := 0
	retryCount := 0
	wait := true

	// Start listening for new blocks
	for i := 0; rep == 0 || i < rep; i++ {
		height := c.GetNextHeight()
		block, err := getBlockFromRPC(height)
		if err != nil {
			Log.WithFields(logrus.Fields{
				"height": height,
				"error":  err,
			}).Warn("error zcashd getblock rpc")
			retryCount++
			if retryCount > 10 {
				Log.WithFields(logrus.Fields{
					"timeouts": retryCount,
				}).Fatal("unable to issue RPC call to zcashd node")
			}
			// Delay then retry the same height.
			c.Sync()
			Sleep(10 * time.Second)
			wait = true
			continue
		}
		retryCount = 0
		if block == nil {
			// No block at this height.
			if wait {
				// Wait a bit then retry the same height.
				c.Sync()
				if lastHeightLogged+1 != height {
					Log.Info("Ingestor waiting for block: ", height)
				}
				Sleep(10 * time.Second)
				wait = false
				continue
			}
		}
		if block == nil || c.HashMismatch(block.PrevHash) {
			// This may not be a reorg; it may be we're at the tip
			// and there's no new block yet, but we want to back up
			// so we detect a reorg in which the new chain is the
			// same length or shorter.
			reorgCount++
			if reorgCount > 100 {
				Log.Fatal("Reorg exceeded max of 100 blocks! Help!")
			}
			// Print the hash of the block that is getting reorg-ed away
			// as 'phash', not the prevhash of the block we just received.
			if block != nil {
				Log.WithFields(logrus.Fields{
					"height": height,
					"hash":   displayHash(block.Hash),
					"phash":  displayHash(c.GetLatestHash()),
					"reorg":  reorgCount,
				}).Warn("REORG")
			} else if reorgCount > 1 {
				Log.WithFields(logrus.Fields{
					"height": height,
					"phash":  displayHash(c.GetLatestHash()),
					"reorg":  reorgCount,
				}).Warn("REORG")
			}
			// Try backing up
			c.Reorg(height - 1)
			Sleep(1 * time.Second)
			continue
		}
		// We have a valid block to add.
		wait = true
		reorgCount = 0
		block.GetHeader()
		if err := c.Add(height, block); err != nil {
			Log.Fatal("Cache add failed:", err)
		}

		// Add it to PostgreSQL
		if db != nil {
			result, err := persistToDB(db, block.Height, block.Hash, block.PrevHash, block.Time, block.GetHeader(), block.GetVtx())
			if err != nil {
				Log.Fatal(result, err)
			}
		}

		// Don't log these too often.
		if time.Now().Sub(lastLog).Seconds() >= 4 && c.GetNextHeight() == height+1 && height != lastHeightLogged {
			lastLog = time.Now()
			lastHeightLogged = height
			Log.Info("Ingestor adding block to cache: ", height)
		}
	}
}

// GetBlock returns the compact block at the requested height, first by querying
// the cache, then, if not found, will request the block from zcashd. It returns
// nil if no block exists at this height.
func GetBlock(cache *BlockCache, height int) (*walletrpc.CompactBlock, error) {
	// First, check the cache to see if we have the block
	block := cache.Get(height)
	if block != nil {
		return block, nil
	}

	// Not in the cache, ask zcashd
	block, err := getBlockFromRPC(height)
	if err != nil {
		return nil, err
	}
	if block == nil {
		// Block height is too large
		return nil, errors.New("block requested is newer than latest block")
	}
	return block, nil
}

// GetBlockRange returns a sequence of consecutive blocks in the given range.
func GetBlockRange(cache *BlockCache, blockOut chan<- walletrpc.CompactBlock, errOut chan<- error, start, end int) {
	// Go over [start, end] inclusive
	for i := start; i <= end; i++ {
		block, err := GetBlock(cache, i)
		if err != nil {
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

// TODO move the postgres DB stuff (the remainder of this source file
// (plus bits from cmd.go maybe) to it's own file
var stmtBlockInsert *sql.Stmt = nil
var stmtBlockDelete *sql.Stmt = nil
var stmtTxInsert *sql.Stmt = nil
var stmtSpendInsert *sql.Stmt = nil
var stmtOutputInsert *sql.Stmt = nil

// SetupPreparedStatements adds prepared statements to insert into the
// 4 tables we use. It also includes delete for blocks to
// support reorgs. We set cascading deletes up so we only need to
// delete the block.
func SetupPreparedStatements(db *sql.DB) (string, error) {
	var err error = nil

	// First the insert prepared statement for the outermost record, blocks
	stmtBlockInsert, err = db.Prepare("INSERT INTO blocks(height, hash, prev_hash, time, header) VALUES ($1, $2, $3, $4, $5);")
	if err != nil {
		return "Unable to prepare SQL blocks insert statement", err
	}

	// Also the update prepared statement for the outermost record,
	// blocks, in case of reorg
	stmtBlockDelete, err = db.Prepare("DELETE FROM blocks WHERE height = $1;")
	if err != nil {
		return "Unable to prepare SQL blocks delete statement", err
	}

	stmtTxInsert, err = db.Prepare("INSERT INTO tx(index, height, hash, fee) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return "Unable to prepare SQL tx insert statement", err
	}

	stmtSpendInsert, err = db.Prepare("INSERT INTO spend(tx_hash, nf) VALUES ($1, $2);")
	if err != nil {
		return "Unable to prepare SQL spend insert statement", err
	}

	stmtOutputInsert, err = db.Prepare("INSERT INTO output(tx_hash, cmu, epk, ciphertext) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return "Unable to prepare SQL output insert statement", err
	}
	return "", nil
}

func persistToDB(db *sql.DB, height uint64, hash []byte, prevHash []byte, time uint32, header []byte, vtx []*walletrpc.CompactTx) (string, error) {
	// Add it to PostgreSQL DB
	var err error = nil

	// First add the blocks record
	// until we fix the header, put a fake header in place if needed
	// (it's always needed because we have not fixed the header yet)
	var tempHeader []byte = nil
	if header == nil {
		tempHeader = []byte("Missing a header still")
	} else {
		tempHeader = header
	}
	_, err = stmtBlockInsert.Exec(height, hash, prevHash, time, tempHeader)
	if err != nil {
		// Block already exists, reorg possible so replace it
		// first get rid of it and related stuff
		_, err = stmtBlockDelete.Exec(height)
		if err != nil {
			return "Unable to delete existing record from DB blocks, ", err
		}

		// deleted OK, (cascading through tx and it's outputs and spends, so put
		// it back in now
		_, err = stmtBlockInsert.Exec(height, hash, prevHash, time, tempHeader)
		if err != nil {
			return "Unable to insert record into DB blocks after deleting, ", err
		}
		return "", nil
	}

	// Now handle the TX array - put it in it's own table with a reference
	// to the height of the related block
	for _, tx := range vtx {
		_, err = stmtTxInsert.Exec(tx.Index, height, tx.GetHash(), tx.GetFee())
		if err != nil {
			return "Unable to insert record into DB tx", err
		}

		// Within each tx, handle the spend array - put it in it's own table
		// with a reference to the TX hash of the related tx
		for _, spend := range tx.GetSpends() {
			_, err = stmtSpendInsert.Exec(tx.GetHash(), spend.GetNf())
			if err != nil {
				return "Unable to insert record into DB spend", err
			}
		}

		// Within each tx, handle the output array - put it in it's own table
		// with a reference to the TX hash of the related tx
		for _, output := range tx.GetOutputs() {
			_, err = stmtOutputInsert.Exec(tx.GetHash(), output.GetCmu(), output.GetEpk(), output.GetCiphertext())
			if err != nil {
				return "Unable to insert record into DB output", err
			}
		}
	}
	return "", nil
}
