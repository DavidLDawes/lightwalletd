// Package parser handles block header and transaction related functions
// Copyright (c) 2019-2020 The Zcash developers
// Forked and modified for the VerusCoin chain
// Copyright 2020 the VerusCoin developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package parser

import (
	"fmt"

	"github.com/asherda/lightwalletd/parser/internal/bytestring"
	"github.com/asherda/lightwalletd/walletrpc"
	"github.com/pkg/errors"
)

// Block wraps a pointer to the block header and a pointer to the transaction (vtx)
type Block struct {
	hdr    *BlockHeader
	vtx    []*Transaction
	height int
}

// NewBlock gets a Block with height set to -1
func NewBlock() *Block {
	return &Block{height: -1}
}

// GetVersion returns the Versiun field from the header, this is the first bit of the header
func (b *Block) GetVersion() int {
	return int(b.hdr.Version)
}

// GetTxCount returns the number of transactions included in the Block that is pointed to in the input
func (b *Block) GetTxCount() int {
	return len(b.vtx)
}

// Transactions simply gets the vtx (transaction pointer) from the Block structure
func (b *Block) Transactions() []*Transaction {
	// TODO: these should NOT be mutable
	return b.vtx
}

// GetDisplayHash returns the block hash in big-endian display order using VerusHash algorithms
func (b *Block) GetDisplayHash() []byte {
	return b.hdr.GetDisplayHash(b.GetHeight())
}

// TODO: encode hash endianness in a type?

// GetEncodableHash returns the block hash in little-endian wire order using VerusHash algorithms
func (b *Block) GetEncodableHash() []byte {
	return b.hdr.GetEncodableHash(b.GetHeight())
}

// GetDisplayPrevHash returns the previous block's hash in big-endian order
func (b *Block) GetDisplayPrevHash() []byte {
	return b.hdr.GetDisplayPrevHash()
}

// HasSaplingTransactions checks whether the block has tsapling TXs, returning true if so
func (b *Block) HasSaplingTransactions() bool {
	for _, tx := range b.vtx {
		if tx.HasSaplingTransactions() {
			return true
		}
	}
	return false
}

// see https://github.com/zcash/lightwalletd/issues/17#issuecomment-467110828
const genesisTargetDifficulty = 520617983

// GetHeight extracts the block height from the coinbase transaction. See BIP34. Returns block height on success, or -1 on error.
func (b *Block) GetHeight() int {
	if b.height != -1 {
		return b.height
	}
	coinbaseScript := bytestring.String(b.vtx[0].transparentInputs[0].ScriptSig)
	var heightNum int64
	if !coinbaseScript.ReadScriptInt64(&heightNum) {
		return -1
	}
	if heightNum < 0 {
		return -1
	}
	// uint32 should last us a while (Nov 2018)
	if heightNum > int64(^uint32(0)) {
		return -1
	}
	blockHeight := uint32(heightNum)

	if blockHeight == genesisTargetDifficulty {
		blockHeight = 0
	}

	b.height = int(blockHeight)
	return int(blockHeight)
}

// GetPrevHash hash value of the prior block that is stored in this block
func (b *Block) GetPrevHash() []byte {
	return b.hdr.HashPrevBlock
}

// ToCompact returns the compact block representation (served by lightwalletd's frontend), simply this and the prior block's hashes, the height, and the time
func (b *Block) ToCompact() *walletrpc.CompactBlock {
	compactBlock := &walletrpc.CompactBlock{
		//TODO ProtoVersion: 1,
		Height:   uint64(b.GetHeight()),
		PrevHash: b.hdr.HashPrevBlock,
		Hash:     b.GetEncodableHash(),
		Time:     b.hdr.Time,
	}

	// Only Sapling transactions have a meaningful compact encoding
	saplingTxns := make([]*walletrpc.CompactTx, 0, len(b.vtx))
	for idx, tx := range b.vtx {
		if tx.HasSaplingTransactions() {
			saplingTxns = append(saplingTxns, tx.ToCompact(idx, b.GetHeight()))
		}
	}
	compactBlock.Vtx = saplingTxns
	return compactBlock
}

// ParseFromSlice parses the byte array for the block (returned from verusd) including the TX data
func (b *Block) ParseFromSlice(data []byte) (rest []byte, err error) {
	hdr := NewBlockHeader()
	data, err = hdr.ParseFromSlice(data)
	if err != nil {
		return nil, errors.Wrap(err, "parsing block header")
	}

	s := bytestring.String(data)
	var txCount int
	if !s.ReadCompactSize(&txCount) {
		return nil, errors.New("could not read tx_count")
	}
	data = []byte(s)

	vtx := make([]*Transaction, 0, txCount)
	for i := 0; len(data) > 0; i++ {
		tx := NewTransaction()
		data, err = tx.ParseFromSlice(data)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("parsing transaction %d", i))
		}
		vtx = append(vtx, tx)
	}

	b.hdr = hdr
	b.vtx = vtx

	return data, nil
}
