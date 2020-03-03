package common

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/davidldawes/lightwalletd/parser"
	"github.com/davidldawes/lightwalletd/walletrpc"
)

var compacts []*walletrpc.CompactBlock
var cache *BlockCache

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
		_, err = block.ParseFromSlice(blockData)
		if err != nil {
			t.Fatal(err)
		}
		compacts = append(compacts, block.ToCompact())
	}

	// Pretend Sapling starts at 289460.
	CacheTestClean("unittestcache")
	cache = NewBlockCache("unittestcache", 289460)

	// Initially cache is empty.
	if cache.GetLatestHeight() != -1 {
		t.Fatal("unexpected GetLatestHeight")
	}
	if cache.FirstBlock != 289460 {
		t.Fatal("unexpected initial FirstBlock")
	}
	if cache.NextBlock != 289460 {
		t.Fatal("unexpected initial NextBlock")
	}
	fillCache(t)
	reorgCache(t)
	fillCache(t)

	// Simulate a restart to ensure the db files are read correctly.
	cache = NewBlockCache("unittestcache", 289460)

	// Should still be 6 blocks.
	if cache.NextBlock != 289460+6 {
		t.Fatal("unexpected NextBlock height")
	}
	reorgCache(t)

	// Reorg to before the first block moves back to only the first block
	if cache.Reorg(0) != 289460 {
		t.Fatal("unexpected reorg height")
	}
	if cache.LatestHash != nil {
		t.Fatal("unexpected LatestHash, should be nil")
	}
	if cache.NextBlock != 289460 {
		t.Fatal("unexpected NextBlock: ", cache.NextBlock)
	}

	// Clean up the test files.
	cache.Close()
	CacheTestClean("unittestcache")
}

func reorgCache(t *testing.T) {
	// Simulate a reorg by adding a block whose height is lower than the latest;
	// we're replacing the second block, so there should be only two blocks.
	if cache.Reorg(289461) != 289461 {
		t.Fatal("unexpected reorg height")
	}
	err := cache.Add(compacts[1])
	if err != nil {
		t.Fatal(err)
	}
	if cache.FirstBlock != 289460 {
		t.Fatal("unexpected FirstBlock height")
	}
	if cache.NextBlock != 289460+2 {
		t.Fatal("unexpected NextBlock height")
	}
	if len(cache.starts) != 3 {
		t.Fatal("unexpected len(cache.starts)")
	}

	// some "black-box" tests (using exported interfaces)
	if cache.GetLatestHeight() != 289461 {
		t.Fatal("unexpected GetLatestHeight")
	}
	if int(cache.Get(289461).Height) != 289461 {
		t.Fatal("unexpected block contents")
	}

	// Make sure we can go forward from here
	err = cache.Add(compacts[2])
	if err != nil {
		t.Fatal(err)
	}
	if cache.FirstBlock != 289460 {
		t.Fatal("unexpected FirstBlock height")
	}
	if cache.NextBlock != 289460+3 {
		t.Fatal("unexpected NextBlock height")
	}
	if len(cache.starts) != 4 {
		t.Fatal("unexpected len(cache.starts)")
	}

	// some "black-box" tests (using exported interfaces)
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
	// Reorg to lower height than FirstBlock returns FirstBlock
	if cache.Reorg(289459) != 289460 {
		t.Fatal("unexpected reorg height")
	}
	for i, compact := range compacts {
		err := cache.Add(compact)
		if err != nil {
			t.Fatal(err)
		}

		// some "white-box" checks
		if cache.FirstBlock != 289460 {
			t.Fatal("unexpected FirstBlock height")
		}
		if cache.NextBlock != 289460+i+1 {
			t.Fatal("unexpected NextBlock height")
		}
		if len(cache.starts) != i+2 {
			t.Fatal("unexpected len(cache.starts)")
		}

		// some "black-box" tests (using exported interfaces)
		if cache.GetLatestHeight() != 289460+i {
			t.Fatal("unexpected GetLatestHeight")
		}
		if int(cache.Get(289460+i).Height) != 289460+i {
			t.Fatal("unexpected block contents")
		}
	}
}
