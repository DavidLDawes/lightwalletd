// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Asherda/lightwalletd/parser"
	"github.com/Asherda/lightwalletd/walletrpc"
	"github.com/syndtr/goleveldb/leveldb"
)

var compacts []*walletrpc.CompactBlock
var cache *BlockCache

const (
	unitTestPath  = "unittestcache"
	unitTestChain = "unittestnet"
)

func TestCache(t *testing.T) {
	type compactTest struct {
		BlockHeight int    `json:"block"`
		BlockHash   string `json:"hash"`
		PrevHash    string `json:"prev"`
		Full        string `json:"full"`
		Compact     string `json:"compact"`
	}
	var compactTests []compactTest

	blockJSON, err := ioutil.ReadFile("../testdata/compact_blocks.json")
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(blockJSON, &compactTests)
	if err != nil {
		t.Fatal(err)
	}

	// Derive compact blocks from file data (setup, not part of the test).
	for _, test := range compactTests {
		blockData, _ := hex.DecodeString(test.Full)
		block := parser.NewBlock()
		blockData, err = block.ParseFromSlice(blockData)
		if err != nil {
			t.Fatal(err)
		}
		if len(blockData) > 0 {
			t.Error("Extra data remaining")
		}
		compacts = append(compacts, block.ToCompact())
	}

	// Pretend Sapling starts at 289460.
	os.RemoveAll(unitTestPath)

	if err := os.MkdirAll(unitTestPath, 0755); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("\n  ** Can't create data directory: %s\n\n", unitTestPath))
		os.Exit(1)
	}
	// leveldb instances are safe for concurrent use.
	db, err := leveldb.OpenFile(unitTestPath, nil)
	defer db.Close()

	cache = NewBlockCache(db, unitTestChain, 289460, true)

	// Initially cache is empty.
	if cache.GetLatestHeight() != -1 {
		t.Fatal("unexpected GetLatestHeight")
	}
	if cache.firstBlock != 289460 {
		t.Fatal("unexpected initial firstBlock")
	}
	if cache.nextBlock != 289460 {
		t.Fatal("unexpected initial nextBlock")
	}
	// Clean up the test files.
	cache.Close()
	os.RemoveAll(unitTestPath)
}

func reorgCache(t *testing.T) {
	// Simulate a reorg by adding a block whose height is lower than the latest;
	// we're replacing the second block, so there should be only two blocks.
	cache.Reorg(289461)
	err := cache.Add(289461, compacts[1])
	if err != nil {
		t.Fatal(err)
	}
	if cache.firstBlock != 289460 {
		t.Fatal("unexpected firstBlock height")
	}
	if cache.nextBlock != 289462 {
		t.Fatal("unexpected nextBlock height")
	}

	// some "black-box" tests (using exported interfaces)
	if cache.GetLatestHeight() != 289461 {
		t.Fatal("unexpected GetLatestHeight")
	}
	if int(cache.Get(289461).Height) != 289461 {
		t.Fatal("unexpected block contents")
	}

	// Make sure we can go forward from here
	err = cache.Add(289462, compacts[2])
	if err != nil {
		t.Fatal(err)
	}
	if cache.firstBlock != 289460 {
		t.Fatal("unexpected firstBlock height")
	}
	if cache.nextBlock != 289463 {
		t.Fatal("unexpected nextBlock height")
	}

	if cache.GetLatestHeight() != 289462 {
		t.Fatal("unexpected GetLatestHeight")
	}
	if int(cache.Get(289462).Height) != 289462 {
		t.Fatal("unexpected block contents")
	}
}

// Whatever the state of the cache, add 6 blocks starting at the
// pretend Sapling height, 289460 (this could cause a reorg).
func fillCache(t *testing.T) {
	next := 289460
	cache.Reorg(next)
	for i, compact := range compacts {
		err := cache.Add(next, compact)
		if err != nil {
			t.Fatal(err)
		}
		next++

		// some "white-box" checks
		if cache.firstBlock != 289460 {
			t.Fatal("unexpected firstBlock height")
		}
		if cache.nextBlock != 289460+i+1 {
			t.Fatal("unexpected nextBlock height")
		}

		// some "black-box" tests (using exported interfaces)
		if cache.GetLatestHeight() != 289460+i {
			t.Fatal("unexpected GetLatestHeight")
		}
		b := cache.Get(289460 + i)
		if b == nil {
			t.Fatal("unexpected Get failure")
		}
		if int(b.Height) != 289460+i {
			t.Fatal("unexpected block contents")
		}
	}
}
