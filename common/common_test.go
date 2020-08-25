// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package common

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Asherda/lightwalletd/walletrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

// ------------------------------------------ Setup
//
// This section does some setup things that may (even if not currently)
// be useful across multiple tests.

var (
	testT *testing.T

	// The various stub callbacks need to sequence through states
	step int

	getblockchaininfoReply []byte
	logger                 = logrus.New()

	getsaplinginfo []byte

	blocks [][]byte // four test blocks
)

var db *leveldb.DB = nil

// TestMain does common setup that's shared across multiple tests
func TestMain(m *testing.M) {
	output, err := os.OpenFile("test-log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Cannot open test-log: %v", err))
		os.Exit(1)
	}
	logger.SetOutput(output)
	Log = logger.WithFields(logrus.Fields{
		"app": "test",
	})

	// leveldb instances are safe for concurrent use.
	db, err = leveldb.OpenFile(unitTestPath, nil)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Cannot open unitTestPath: %s, error: %v", unitTestPath, err))
		os.Exit(1)
	}

	getsaplinginfo, err := ioutil.ReadFile("../testdata/getsaplinginfo")
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Cannot open testdata/getsaplinginfo: %v", err))
		os.Exit(1)
	}
	getblockchaininfoReply, _ = hex.DecodeString(string(getsaplinginfo))

	// Several tests need test blocks; read all 4 into memory just once
	// (for efficiency).
	testBlocks, err := os.Open("../testdata/blocks")
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Cannot open testdata/blocks: %v", err))
		os.Exit(1)
	}
	scan := bufio.NewScanner(testBlocks)
	for scan.Scan() { // each line (block)
		block := scan.Bytes()
		// Enclose the hex string in quotes (to make it json, to match what's
		// returned by the RPC)
		block = []byte("\"" + string(block) + "\"")
		blocks = append(blocks, block)
	}

	// Setup is done; run all tests.
	exitcode := m.Run()

	// cleanup
	os.Remove("test-log")

	os.Exit(exitcode)
}

// Allow tests to verify that sleep has been called (for retries)
var sleepCount int
var sleepDuration time.Duration

func sleepStub(d time.Duration) {
	sleepCount++
	sleepDuration += d
}

// ------------------------------------------ GetSaplingInfo()

func getblockchaininfoStub(method string, params []json.RawMessage) (json.RawMessage, error) {
	step++
	// Test retry logic (for the moment, it's very simple, just one retry).
	switch step {
	case 1:
		return getblockchaininfoReply, errors.New("first failure")
	}
	if sleepCount != 1 || sleepDuration != 15*time.Second {
		testT.Error("unexpected sleeps", sleepCount, sleepDuration)
	}
	return getblockchaininfoReply, nil
}

func TestGetSaplingInfo(t *testing.T) {
	testT = t
	RawRequest = getblockchaininfoStub
	Sleep = sleepStub
	saplingHeight, blockHeight, chainName, branchID := GetSaplingInfo()

	// Ensure the retry happened as expected
	logFile, err := ioutil.ReadFile("test-log")
	if err != nil {
		t.Fatal("Cannot read test-log", err)
	}
	logStr := string(logFile)
	if !strings.Contains(logStr, "retrying") {
		t.Fatal("Cannot find retrying in test-log")
	}
	if !strings.Contains(logStr, "retry=1") {
		t.Fatal("Cannot find retry=1 in test-log")
	}

	// Check the success case (second attempt)
	if saplingHeight != 419200 {
		t.Error("unexpected saplingHeight", saplingHeight)
	}
	if blockHeight != 677713 {
		t.Error("unexpected blockHeight", blockHeight)
	}
	if chainName != "main" {
		t.Error("unexpected chainName", chainName)
	}
	if branchID != "2bb40e60" {
		t.Error("unexpected branchID", branchID)
	}

	if sleepCount != 1 || sleepDuration != 15*time.Second {
		t.Error("unexpected sleeps", sleepCount, sleepDuration)
	}
	step = 0
	sleepCount = 0
	sleepDuration = 0
}

// ------------------------------------------ BlockIngestor()

// There are four test blocks, 0..3
func getblockStub(method string, params []json.RawMessage) (json.RawMessage, error) {
	var height string
	err := json.Unmarshal(params[0], &height)
	if err != nil {
		testT.Fatal("could not unmarshal height")
	}

	step++
	switch step {
	case 1:
		if height != "380640" {
			testT.Error("unexpected height")
		}
		// Sunny-day
		return blocks[0], nil
	case 2:
		if height != "380641" {
			testT.Error("unexpected height")
		}
		// Sunny-day
		return blocks[1], nil
	case 3:
		if height != "380642" {
			testT.Error("unexpected height", height)
		}
		// Simulate that we're synced (caught up);
		// this should cause one 10s sleep (then retry).
		return nil, errors.New("-8: Block height out of range")
	case 4:
		if sleepCount != 1 || sleepDuration != 2*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		if height != "380642" {
			testT.Error("unexpected height", height)
		}
		// Simulate that we're still caught up; this should cause a 1s
		// wait then a check for reorg to shorter chain (back up one).
		return nil, errors.New("-8: Block height out of range")
	case 5:
		if sleepCount != 1 || sleepDuration != 2*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		// Back up to 41.
		if height != "380641" {
			testT.Error("unexpected height", height)
		}
		// Return the expected block (as normally happens, no actual reorg),
		// ingestor will immediately re-request the next block (42).
		return blocks[1], nil
	case 6:
		if sleepCount != 1 || sleepDuration != 2*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		if height != "380642" {
			testT.Error("unexpected height", height)
		}
		// Block 42 has now finally appeared, it will immediately ask for 43.
		return blocks[2], nil
	case 7:
		if sleepCount != 1 || sleepDuration != 2*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		if height != "380643" {
			testT.Error("unexpected height", height)
		}
		// Simulate a reorg by modifying the block's hash temporarily,
		// this causes a 1s sleep and then back up one block (to 42).
		blocks[3][9]++ // first byte of the prevhash
		return blocks[3], nil
	case 8:
		blocks[3][9]-- // repair first byte of the prevhash
		if sleepCount != 1 || sleepDuration != 2*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		if height != "380642" {
			testT.Error("unexpected height ", height)
		}
		return blocks[2], nil
	case 9:
		if sleepCount != 1 || sleepDuration != 2*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		if height != "380643" {
			testT.Error("unexpected height ", height)
		}
		// Instead of returning expected (43), simulate block unmarshal
		// failure, should cause 10s sleep, retry
		return nil, nil
	case 10:
		if sleepCount != 2 || sleepDuration != 12*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		if height != "380643" {
			testT.Error("unexpected height ", height)
		}
		// Back to sunny-day
		return blocks[3], nil
	case 11:
		if sleepCount != 2 || sleepDuration != 12*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		if height != "380644" {
			testT.Error("unexpected height ", height)
		}
		// next block not ready
		return nil, nil
	}
	testT.Error("getblockStub called too many times")
	return nil, nil
}

/*
func TestBlockIngestor(t *testing.T) {
	testT = t
	RawRequest = getblockStub
	Sleep = sleepStub
	os.RemoveAll(unitTestPath)
	testcache := NewBlockCache(db, unitTestChain, 380640, false)
	testcache.latestHash = make([]byte, 32)
	_, err := hex.Decode(testcache.latestHash, []byte("8a024cebb99e30ff83d5b9f50cc5303351923da95a8dc7fda3e0160900000000"))
	if err != nil {
		t.Error("failed todecode latestHash value successfully")
	}
	BlockIngestor(testcache, 11)
	if step != 11 {
		t.Error("unexpected final step", step)
	}
	step = 0
	sleepCount = 0
	sleepDuration = 0
	testcache.setDbHeight(380642)
	if testcache.GetNextHeight() != 380642 {
		t.Error("unable to reset height back to 280642")
	}
	cBlock := testcache.Get(380642)
	if cBlock != nil {
		t.Error("found a block at 380642 but we set DB height below that so it should have been flushed")
	}
	cBlock = testcache.Get(380641)
	if cBlock == nil {
		t.Error("failed to find a block at 280641, we reset the DB height to 1 higher so it should still be thre")
	}
	cBlock = testcache.Get(380640)
	if cBlock == nil {
		t.Error("failed to find a block at 280640, we reset the DB height to 2 higher so it should still be thre")
	}

	os.RemoveAll(unitTestPath)
}
*/
func TestGetBlockRange(t *testing.T) {
	testT = t
	RawRequest = getblockStub
	os.RemoveAll(unitTestPath)
	testcache := NewBlockCache(db, unitTestChain, 380640, true)
	blockChan := make(chan *walletrpc.CompactBlock)
	errChan := make(chan error)
	go GetBlockRange(testcache, blockChan, errChan, 380640, 380642)

	// read in block 380640
	select {
	case err := <-errChan:
		// this will also catch context.DeadlineExceeded from the timeout
		t.Fatal("unexpected error:", err)
	case cBlock := <-blockChan:
		if cBlock.Height != 380640 {
			t.Fatal("unexpected Height:", cBlock.Height)
		}
	}

	// read in block 380641
	select {
	case err := <-errChan:
		// this will also catch context.DeadlineExceeded from the timeout
		t.Fatal("unexpected error:", err)
	case cBlock := <-blockChan:
		if cBlock.Height != 380641 {
			t.Fatal("unexpected Height:", cBlock.Height)
		}
	}
}

func TestGenerateCerts(t *testing.T) {
	if GenerateCerts() == nil {
		t.Fatal("GenerateCerts returned nil")
	}
}
