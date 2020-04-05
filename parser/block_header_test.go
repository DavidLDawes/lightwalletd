package parser

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"math/big"
	"os"
	"testing"
)

// https://bitcoin.org/en/developer-reference#target-nbits
var nbitsTests = []struct {
	bytes  []byte
	target string
}{
	{
		[]byte{0x18, 0x1b, 0xc3, 0x30},
		"1bc330000000000000000000000000000000000000000000",
	},
	{
		[]byte{0x01, 0x00, 0x34, 0x56},
		"00",
	},
	{
		[]byte{0x01, 0x12, 0x34, 0x56},
		"12",
	},
	{
		[]byte{0x02, 0x00, 0x80, 00},
		"80",
	},
	{
		[]byte{0x05, 0x00, 0x92, 0x34},
		"92340000",
	},
	{
		[]byte{0x04, 0x92, 0x34, 0x56},
		"-12345600",
	},
	{
		[]byte{0x04, 0x12, 0x34, 0x56},
		"12345600",
	},
}

func TestParseNBits(t *testing.T) {
	for i, tt := range nbitsTests {
		target := parseNBits(tt.bytes)
		expected, _ := new(big.Int).SetString(tt.target, 16)
		if target.Cmp(expected) != 0 {
			t.Errorf("NBits parsing failed case %d:\nwant: %x\nhave: %x", i, expected, target)
		}
	}
}

func TestBlockHeader(t *testing.T) {
	testBlocks, err := os.Open("../testdata/blocks")
	if err != nil {
		t.Fatal(err)
	}
	defer testBlocks.Close()

	lastBlockTime := uint32(0)

	scan := bufio.NewScanner(testBlocks)
	for scan.Scan() {
		blockDataHex := scan.Text()
		blockData, err := hex.DecodeString(blockDataHex)
		if err != nil {
			t.Error(err)
			continue
		}

		blockHeader := NewBlockHeader()
		_, err = blockHeader.ParseFromSlice(blockData)
		if err != nil {
			t.Error(err)
			continue
		}

		// Some basic sanity checks
		if blockHeader.Version != 4 {
			t.Error("Read wrong version in a test block.")
			break
		}

		if blockHeader.Time < lastBlockTime {
			t.Error("Block times not increasing.")
			break
		}
		lastBlockTime = blockHeader.Time

		if len(blockHeader.Solution) != equihashSizeMainnet {
			t.Error("Got wrong Equihash solution size.")
			break
		}

		// Re-serialize and check for consistency
		serializedHeader, err := blockHeader.MarshalBinary()
		if err != nil {
			t.Errorf("Error serializing header: %v", err)
			break
		}

		if !bytes.Equal(serializedHeader, blockData[:serBlockHeaderMinusEquihashSize+3+equihashSizeMainnet]) {
			offset := 0
			length := 0
			for i := 0; i < len(serializedHeader); i++ {
				if serializedHeader[i] != blockData[i] {
					if offset == 0 {
						offset = i
					}
					length++
				}
			}
			t.Errorf(
				"Block header failed round-trip:\ngot\n%x\nwant\n%x\nfirst diff at %d",
				serializedHeader[offset:offset+length],
				blockData[offset:offset+length],
				offset,
			)
			break
		}

		hash := blockHeader.GetDisplayHash(1)
		// test caching
		if !bytes.Equal(hash, blockHeader.GetDisplayHash(1)) {
			t.Error("caching is broken")
		}

        if bytes.Compare(hash, []byte{0x8e, 0xeb, 0xcb, 0xf6, 0xf6, 0xd1, 0xa8, 0x45, 0x7d, 0x6a, 0x09, 0xf9, 0x28, 0xdb, 0x8d, 0xe4, 0xb9, 0xad, 0xfc, 0x8b, 0x87, 0x05, 0x52, 0x8d, 0x4d, 0xe4, 0x85, 0x10, 0xbc, 0x83, 0x57, 0xc5}) != 0 &&
           bytes.Compare(hash, []byte{0xb4, 0xb8, 0x35, 0xcd, 0x7b, 0xc1, 0xc0, 0x94, 0xae, 0x1d, 0xd3, 0x6b, 0xc1, 0x54, 0xf3, 0x35, 0x11, 0x28, 0x03, 0xac, 0x75, 0xd8, 0x13, 0x08, 0x9a, 0x96, 0x2e, 0xb9, 0x7b, 0x43, 0x41, 0x63}) != 0 &&
           bytes.Compare(hash, []byte{0x62, 0x5f, 0xc2, 0x3e, 0xe2, 0xed, 0x34, 0x2a, 0xfd, 0xa7, 0x95, 0xc9, 0x9a, 0x94, 0x68, 0xd6, 0xc7, 0xdd, 0x96, 0xd6, 0x73, 0xc0, 0x74, 0xd4, 0x5c, 0x99, 0x19, 0xbe, 0x9f, 0x19, 0x37, 0x02}) != 0 &&
           bytes.Compare(hash, []byte{0x2e, 0x9d, 0xdb, 0xea, 0x66, 0x04, 0xd7, 0xdc, 0x6e, 0xaa, 0x24, 0x9e, 0xe3, 0x33, 0x98, 0xf3, 0x10, 0xbf, 0x07, 0xe9, 0xaf, 0x5d, 0x9c, 0x82, 0xc3, 0xe2, 0x88, 0x5b, 0xdd, 0x8e, 0x3d, 0xbe}) != 0 {
            t.Errorf("Hash not among expected values: %x, huh %x", hash, []byte("8eebcbf6f6d1a8457d6a09f928db8de4b9adfc8b8705528d4de48510bc8357c5"))
        }
	}
}

func TestBadBlockHeader(t *testing.T) {
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
	// these bad blocks are short in various ways
	for i := 1; scan.Scan(); i++ {
		blockDataHex := scan.Text()
		blockData, err := hex.DecodeString(blockDataHex)
		if err != nil {
			t.Error(err)
			continue
		}

		blockHeader := NewBlockHeader()
		_, err = blockHeader.ParseFromSlice(blockData)
		if err == nil {
			t.Errorf("unexpected success parsing bad block %d", i)
		}
	}
}

var compactLengthPrefixedLenTests = []struct {
	length       int
	returnLength int
}{
	/* 00 */ {0, 1},
	/* 01 */ {1, 1 + 1},
	/* 02 */ {2, 1 + 2},
	/* 03 */ {252, 1 + 252},
	/* 04 */ {253, 1 + 2 + 253},
	/* 05 */ {0xffff, 1 + 2 + 0xffff},
	/* 06 */ {0x10000, 1 + 4 + 0x10000},
	/* 07 */ {0x10001, 1 + 4 + 0x10001},
	/* 08 */ {0xffffffff, 1 + 4 + 0xffffffff},
	/* 09 */ {0x100000000, 1 + 8 + 0x100000000},
	/* 10 */ {0x100000001, 1 + 8 + 0x100000001},
}

func TestCompactLengthPrefixedLen(t *testing.T) {
	for i, tt := range compactLengthPrefixedLenTests {
		returnLength := CompactLengthPrefixedLen(tt.length)
		if returnLength != tt.returnLength {
			t.Errorf("TestCompactLengthPrefixedLen case %d: want: %v have %v",
				i, tt.returnLength, returnLength)
		}
	}
}

var writeCompactLengthPrefixedTests = []struct {
	argLen       int
	returnLength int
	header       []byte
}{
	/* 00 */ {0, 1, []byte{0}},
	/* 01 */ {1, 1, []byte{1}},
	/* 02 */ {2, 1, []byte{2}},
	/* 03 */ {252, 1, []byte{252}},
	/* 04 */ {253, 1 + 2, []byte{253, 253, 0}},
	/* 05 */ {254, 1 + 2, []byte{253, 254, 0}},
	/* 06 */ {0xffff, 1 + 2, []byte{253, 0xff, 0xff}},
	/* 07 */ {0x10000, 1 + 4, []byte{254, 0x00, 0x00, 0x01, 0x00}},
	/* 08 */ {0x10003, 1 + 4, []byte{254, 0x03, 0x00, 0x01, 0x00}},
	/* 09 */ {0xffffffff, 1 + 4, []byte{254, 0xff, 0xff, 0xff, 0xff}},
	/* 10 */ {0x100000000, 1 + 8, []byte{255, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}},
	/* 11 */ {0x100000007, 1 + 8, []byte{255, 0x07, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}},
}

func TestWriteCompactLengthPrefixedLen(t *testing.T) {
	for i, tt := range writeCompactLengthPrefixedTests {
		var b bytes.Buffer
		WriteCompactLengthPrefixedLen(&b, tt.argLen)
		if b.Len() != tt.returnLength {
			t.Fatalf("TestWriteCompactLengthPrefixed case %d: unexpected length", i)
		}
		// check the header (tag and length)
		r := make([]byte, len(tt.header))
		b.Read(r)
		if !bytes.Equal(r, tt.header) {
			t.Fatalf("TestWriteCompactLengthPrefixed case %d: incorrect header", i)
		}
		if b.Len() > 0 {
			t.Fatalf("TestWriteCompactLengthPrefixed case %d: unexpected data remaining", i)
		}
	}
}

func TestWriteCompactLengthPrefixed(t *testing.T) {
	var b bytes.Buffer
	val := []byte{22, 33, 44}
	WriteCompactLengthPrefixed(&b, val)
	r := make([]byte, 4)
	b.Read(r)
	expected := []byte{3, 22, 33, 44}
	if !bytes.Equal(r, expected) {
		t.Fatal("TestWriteCompactLengthPrefixed incorrect result")
	}
}

func Contains(a []string, x string) bool {
    for _, n := range a {
        if x == n {
            return true
        }
    }
    return false
}
