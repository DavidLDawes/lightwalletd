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

	"github.com/asherda/lightwalletd/walletrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
		// This should cause one sleep (then retry)
		return nil, errors.New("-8: Block height out of range")
	case 4:
		if sleepCount != 1 || sleepDuration != 10*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		// should re-request the same height
		if height != "380642" {
			testT.Error("unexpected height", height)
		}
		// Back to sunny-day
		return blocks[2], nil
	case 5:
		if height != "380643" {
			testT.Error("unexpected height", height)
		}
		// Simulate a reorg by modifying the block's hash temporarily
		blocks[3][9]++ // first byte of the prevhash
		return blocks[3], nil
	case 6:
		// When a reorg occurs, the ingestor backs up 2 blocks
		blocks[3][9]-- // repair first byte of the prevhash
		if height != "380641" {
			testT.Error("unexpected height ", height)
		}
		return blocks[1], nil
	case 7:
		if height != "380642" {
			testT.Error("unexpected height ", height)
		}
		// Should fail to Unmarshal the block, sleep, retry
		return nil, nil
	case 8:
		if sleepCount != 2 || sleepDuration != 20*time.Second {
			testT.Error("unexpected sleeps", sleepCount, sleepDuration)
		}
		if height != "380642" {
			testT.Error("unexpected height ", height)
		}
		// Back to sunny-day
		return blocks[2], nil
	}
	if height != "380643" {
		testT.Error("unexpected height ", height)
	}
	testT.Error("getblockStub called too many times")
	return nil, nil
}

func TestBlockIngestor(t *testing.T) {
	testT = t
	RawRequest = getblockStub
	Sleep = sleepStub
	testcache := NewBlockCache("unittestcache", 380640)
	BlockIngestor(testcache, 380640, 7)
	if step != 7 {
		t.Error("unexpected final step", step)
	}
	step = 0
	sleepCount = 0
	sleepDuration = 0
}

func TestGetBlockRange(t *testing.T) {
	testT = t
	RawRequest = getblockStub
	CacheTestClean("unittestcache")
	testcache := NewBlockCache("unittestcache", 380640)
	blockChan := make(chan walletrpc.CompactBlock)
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

	// try to read in block 380642, but this will fail (see case 3 above)
	select {
	case err := <-errChan:
		// this will also catch context.DeadlineExceeded from the timeout
		if err.Error() != "block requested is newer than latest block" {
			t.Fatal("unexpected error:", err)
		}
	case _ = <-blockChan:
		t.Fatal("reading height 22 should have failed")
	}

	// check goroutine GetBlockRange() reaching the end of the range (and exiting)
	go GetBlockRange(testcache, blockChan, errChan, 1, 0)
	err := <-errChan
	if err != nil {
		t.Fatal("unexpected err return")
	}
}

func TestGenerateCerts(t *testing.T) {
	if GenerateCerts() == nil {
		t.Fatal("GenerateCerts returned nil")
	}
}
