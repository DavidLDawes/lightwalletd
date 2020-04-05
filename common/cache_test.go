// Package common includes logging, certs, caching, and the ingestor (in common.go)
// Copyright (c) 2019-2020 The Zcash developers
// Forked and modified for the VerusCoin chain
// Copyright 2020 the VerusCoin developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package common

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/asherda/lightwalletd/parser"
	"github.com/asherda/lightwalletd/walletrpc"
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
	var compacts []*walletrpc.CompactBlock

	blockJSON, err := ioutil.ReadFile("../testdata/compact_blocks.json")
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(blockJSON, &compactTests)
	if err != nil {
		t.Fatal(err)
	}
	cache := NewBlockCache(4, 1)

	// derive compact blocks from file data (setup, not part of the test)
	for _, test := range compactTests {
		blockData, _ := hex.DecodeString(test.Full)
		block := parser.NewBlock()
		_, err = block.ParseFromSlice(blockData)
		if err != nil {
			t.Fatal(err)
		}
		compacts = append(compacts, block.ToCompact())
	}

	// initially empty cache
	if cache.GetLatestHeight() != -1 {
		t.Fatal("unexpected GetLatestHeight")
	}

	// Test handling an invalid block (nil will do)
	reorg, err := cache.Add(21, nil)
	if err == nil {
		t.Error("expected error:", err)
	}
	if reorg {
		t.Fatal("unexpected reorg")
	}

	// No entries just before and just after the cache range
	if cache.Get(11) != nil || cache.Get(16) != nil {
		t.Fatal("unexpected Get")
	}

	// We can re-add the last block (with the same data) and
	// that should just replace and not be considered a reorg
	reorg, err = cache.Add(15, compacts[5])
	if err != nil {
		t.Fatal(err)
	}
	if reorg {
		t.Fatal("unexpected reorg")
	}

	// Simulate a reorg by resubmitting as the next block, 16, any block with
	// the wrote prev-hash (let's use the first, just because it's handy)
	reorg, err = cache.Add(16, compacts[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reorg {
		t.Fatal("unexpected non-reorg")
	}
	// The cache shouldn't have changed in any way
	if cache.Get(16) != nil {
		t.Fatal("unexpected block 16 exists")
	}
	if cache.GetLatestHeight() != 15 {
		t.Fatal("unexpected GetLatestHeight")
	}
	if int(cache.Get(15).Height) != 289460+5 {
		t.Fatal("unexpected Get")
	}
}
