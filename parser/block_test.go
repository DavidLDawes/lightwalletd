package parser

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"

	protobuf "github.com/golang/protobuf/proto"
)

func TestBlockParser(t *testing.T) {
	// These (valid on testnet) correspond to the transactions in testdata/blocks
	var txhashes = []string{
		"8d8e1184148cae7c87a6cd50cd8fa07caa0811da5013b4652b9b076c84d92f79",
		"08df1dac9c0710606cfd58446169dab0585ca86dac21f4dccda91573f9630372",
		"838ce6b30b1d67eccd1d7625f86a033e20dcd744f5cf0312224e14c3197dc938",
		"542e7788130fac94701ee53c5bed21198018dd8150a08afb77fdcc8cec0d115b",
		"33b57226b991ed0d30f8d53ada7a1ff6cfb80a9b173f1c198d251932f7363668",
		"e770275b22e6b906ceb2bec5754032bea0642a58657668d5d128e09706ba39e0",
	}
	txindex := 0
	testBlocks, err := os.Open("../testdata/blocks")
	if err != nil {
		t.Fatal(err)
	}
	defer testBlocks.Close()

	scan := bufio.NewScanner(testBlocks)
	for i := 0; scan.Scan(); i++ {
		blockDataHex := scan.Text()
		blockData, err := hex.DecodeString(blockDataHex)
		if err != nil {
			t.Error(err)
			continue
		}

		block := NewBlock()
		blockData, err = block.ParseFromSlice(blockData)
		if err != nil {
			t.Error(errors.Wrap(err, fmt.Sprintf("parsing block %d", i)))
			continue
		}

		// Some basic sanity checks
		if block.hdr.Version != 4 {
			t.Error("Read wrong version in a test block.")
			break
		}
		if block.GetVersion() != 4 {
			t.Error("Read wrong version in a test block.")
			break
		}
		if block.GetTxCount() < 1 {
			t.Error("No transactions in block")
			break
		}
		if len(block.Transactions()) != block.GetTxCount() {
			t.Error("Number of transactions mismatch")
			break
		}
		if block.HasSaplingTransactions() {
			t.Error("Unexpected Saping tx")
			break
		}
		for _, tx := range block.Transactions() {
			if tx.HasSaplingTransactions() {
				t.Error("Unexpected Saping tx")
				break
			}
			if hex.EncodeToString(tx.GetDisplayHash(1)) != txhashes[txindex] {
				t.Error("incorrect tx hash. Expected ", hex.EncodeToString(tx.GetDisplayHash(1)))
			}
			txindex++
		}
	}
}

func TestBlockParserFail(t *testing.T) {
	testBlocks, err := os.Open("../testdata/badblocks")
	if err != nil {
		t.Fatal(err)
	}
	defer testBlocks.Close()

	scan := bufio.NewScanner(testBlocks)

	// the first "block" contains an illegal hex character
	{
		scan.Scan()
		blockDataHex := scan.Text()
		_, err := hex.DecodeString(blockDataHex)
		if err == nil {
			t.Error("unexpected success parsing illegal hex bad block")
		}
	}
	for i := 0; scan.Scan(); i++ {
		blockDataHex := scan.Text()
		blockData, err := hex.DecodeString(blockDataHex)
		if err != nil {
			t.Error(err)
			continue
		}

		block := NewBlock()
		blockData, err = block.ParseFromSlice(blockData)
		if err == nil {
			t.Error("unexpected success parsing bad block")
		}
	}
}

// Checks on the first 20 blocks from mainnet genesis.
func TestGenesisBlockParser(t *testing.T) {
	blockFile, err := os.Open("../testdata/mainnet_genesis")
	if err != nil {
		t.Fatal(err)
	}
	defer blockFile.Close()

	scan := bufio.NewScanner(blockFile)
	for i := 0; scan.Scan(); i++ {
		blockDataHex := scan.Text()
		blockData, err := hex.DecodeString(blockDataHex)
		if err != nil {
			t.Error(err)
			continue
		}

		block := NewBlock()
		blockData, err = block.ParseFromSlice(blockData)
		if err != nil {
			t.Error(err)
			continue
		}

		// Some basic sanity checks
		if block.hdr.Version != 4 {
			t.Error("Read wrong version in genesis block.")
			break
		}

		if block.GetHeight() != i {
			t.Errorf("Got wrong height for block %d: %d", i, block.GetHeight())
		}
	}
}

func TestCompactBlocks(t *testing.T) {
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

	for _, test := range compactTests {
		blockData, _ := hex.DecodeString(test.Full)
		block := NewBlock()
		blockData, err = block.ParseFromSlice(blockData)
		if err != nil {
			t.Error(errors.Wrap(err, fmt.Sprintf("parsing testnet block %d", test.BlockHeight)))
			continue
		}
		if block.GetHeight() != test.BlockHeight {
			t.Errorf("incorrect block height in testnet block %d", test.BlockHeight)
			continue
		}
		if hex.EncodeToString(block.GetDisplayHash(1)) != test.BlockHash {
			t.Errorf("incorrect block hash in testnet block %x", block.GetDisplayHash(1))
			continue
		}
		if hex.EncodeToString(block.GetDisplayPrevHash()) != test.PrevHash {
			t.Errorf("incorrect block prevhash in testnet block %x", block.GetDisplayPrevHash())
			continue
		}
		if !bytes.Equal(block.GetPrevHash(), block.hdr.HashPrevBlock) {
			t.Error("block and block header prevhash don't match")
		}

		compact := block.ToCompact()
		marshaled, err := protobuf.Marshal(compact)
		if err != nil {
			t.Errorf("could not marshal compact testnet block %d", test.BlockHeight)
			continue
		}
		encodedCompact := hex.EncodeToString(marshaled)
		if encodedCompact != test.Compact {
			t.Errorf("wrong data for compact testnet block %d\nhave: %s\nwant: %s\n", test.BlockHeight, encodedCompact, test.Compact)
			break
		}
	}

}
