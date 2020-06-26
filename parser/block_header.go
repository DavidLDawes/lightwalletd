// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package parser

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"unsafe"

	"github.com/asherda/lightwalletd/parser/internal/bytestring"
	"github.com/asherda/lightwalletd/parser/verushash"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	serBlockHeaderMinusEquihashSize = 140  // size of a serialized block header minus the Equihash solution
	equihashSizeMainnet             = 1344 // size of a mainnet / testnet Equihash solution in bytes
)

// A block header as defined in version 2018.0-beta-29 of the Zcash Protocol Spec.
type rawBlockHeader struct {
	// The block version number indicates which set of block validation rules
	// to follow. The current and only defined block version number for Zcash
	// is 4.
	Version int32

	// A SHA-256d hash in internal byte order of the previous block's header. This
	// ensures no previous block can be changed without also changing this block's
	// header.
	HashPrevBlock []byte

	// A SHA-256d hash in internal byte order. The merkle root is derived from
	// the hashes of all transactions included in this block, ensuring that
	// none of those transactions can be modified without modifying the header.
	HashMerkleRoot []byte

	// [Pre-Sapling] A reserved field which should be ignored.
	// [Sapling onward] The root LEBS2OSP_256(rt) of the Sapling note
	// commitment tree corresponding to the final Sapling treestate of this
	// block.
	HashFinalSaplingRoot []byte

	// The block time is a Unix epoch time (UTC) when the miner started hashing
	// the header (according to the miner).
	Time uint32

	// An encoded version of the target threshold this block's header hash must
	// be less than or equal to, in the same nBits format used by Bitcoin.
	NBitsBytes []byte

	// An arbitrary field that miners can change to modify the header hash in
	// order to produce a hash less than or equal to the target threshold.
	Nonce []byte

	// The Equihash solution. In the wire format, this is a
	// CompactSize-prefixed value.
	Solution []byte
}

// ParserLog - stash the Log pointer here so we can log while debugging
var ParserLog *logrus.Entry

type BlockHeader struct {
	*rawBlockHeader
	cachedHash      []byte
	targetThreshold *big.Int
}

// VerusHash holds the VerusHash object used for the VerusCoin hashing methods
var VerusHash = verushash.NewVerushash()

// CompactLengthPrefixedLen calculates the total number of bytes needed to
// encode 'length' bytes.
func CompactLengthPrefixedLen(length int) int {
	if length < 253 {
		return 1 + length
	} else if length <= 0xffff {
		return 1 + 2 + length
	} else if length <= 0xffffffff {
		return 1 + 4 + length
	} else {
		return 1 + 8 + length
	}
}

// WriteCompactLengthPrefixedLen writes the given length to the stream.
func WriteCompactLengthPrefixedLen(buf *bytes.Buffer, length int) {
	if length < 253 {
		binary.Write(buf, binary.LittleEndian, uint8(length))
	} else if length <= 0xffff {
		binary.Write(buf, binary.LittleEndian, byte(253))
		binary.Write(buf, binary.LittleEndian, uint16(length))
	} else if length <= 0xffffffff {
		binary.Write(buf, binary.LittleEndian, byte(254))
		binary.Write(buf, binary.LittleEndian, uint32(length))
	} else {
		binary.Write(buf, binary.LittleEndian, byte(255))
		binary.Write(buf, binary.LittleEndian, uint64(length))
	}
}

func WriteCompactLengthPrefixed(buf *bytes.Buffer, val []byte) {
	WriteCompactLengthPrefixedLen(buf, len(val))
	binary.Write(buf, binary.LittleEndian, val)
}

func (hdr *rawBlockHeader) GetSize() int {
	return serBlockHeaderMinusEquihashSize + CompactLengthPrefixedLen(len(hdr.Solution))
}

func (hdr *rawBlockHeader) MarshalBinary() ([]byte, error) {
	headerSize := hdr.GetSize()
	backing := make([]byte, 0, headerSize)
	buf := bytes.NewBuffer(backing)
	binary.Write(buf, binary.LittleEndian, hdr.Version)
	binary.Write(buf, binary.LittleEndian, hdr.HashPrevBlock)
	binary.Write(buf, binary.LittleEndian, hdr.HashMerkleRoot)
	binary.Write(buf, binary.LittleEndian, hdr.HashFinalSaplingRoot)
	binary.Write(buf, binary.LittleEndian, hdr.Time)
	binary.Write(buf, binary.LittleEndian, hdr.NBitsBytes)
	binary.Write(buf, binary.LittleEndian, hdr.Nonce)
	WriteCompactLengthPrefixed(buf, hdr.Solution)
	return backing[:headerSize], nil
}

func NewBlockHeader() *BlockHeader {
	return &BlockHeader{
		rawBlockHeader: new(rawBlockHeader),
	}
}

func BlockHeaderFromParts(version int32, prevhash []byte, merkleroot []byte, saplingroot []byte, time uint32, nbitsbytes []byte, nonce []byte, solution []byte) *BlockHeader {
	return &BlockHeader{
		rawBlockHeader: &rawBlockHeader{
			Version:              version,
			HashPrevBlock:        prevhash,
			HashMerkleRoot:       merkleroot,
			HashFinalSaplingRoot: saplingroot,
			Time:                 time,
			NBitsBytes:           nbitsbytes,
			Nonce:                nonce,
			Solution:             solution,
		},
	}
}

// ParseFromSlice parses the block header struct from the provided byte slice,
// advancing over the bytes read. If successful it returns the rest of the
// slice, otherwise it returns the input slice unaltered along with an error.
func (hdr *BlockHeader) ParseFromSlice(in []byte) (rest []byte, err error) {
	s := bytestring.String(in)

	// Primary parsing layer: sort the bytes into things

	if !s.ReadInt32(&hdr.Version) {
		return in, errors.New("could not read header version")
	}

	if !s.ReadBytes(&hdr.HashPrevBlock, 32) {
		return in, errors.New("could not read HashPrevBlock")
	}

	if !s.ReadBytes(&hdr.HashMerkleRoot, 32) {
		return in, errors.New("could not read HashMerkleRoot")
	}

	if !s.ReadBytes(&hdr.HashFinalSaplingRoot, 32) {
		return in, errors.New("could not read HashFinalSaplingRoot")
	}

	if !s.ReadUint32(&hdr.Time) {
		return in, errors.New("could not read timestamp")
	}

	if !s.ReadBytes(&hdr.NBitsBytes, 4) {
		return in, errors.New("could not read NBits bytes")
	}

	if !s.ReadBytes(&hdr.Nonce, 32) {
		return in, errors.New("could not read Nonce bytes")
	}

	if !s.ReadCompactLengthPrefixed((*bytestring.String)(&hdr.Solution)) {
		return in, errors.New("could not read CompactSize-prefixed Equihash solution")
	}

	// TODO: interpret the bytes
	//hdr.targetThreshold = parseNBits(hdr.NBitsBytes)

	return []byte(s), nil
}

func parseNBits(b []byte) *big.Int {
	byteLen := int(b[0])

	targetBytes := make([]byte, byteLen)
	copy(targetBytes, b[1:])

	// If high bit set, return a negative result. This is in the Bitcoin Core
	// test vectors even though Bitcoin itself will never produce or interpret
	// a difficulty lower than zero.
	if b[1]&0x80 != 0 {
		targetBytes[0] &= 0x7F
		target := new(big.Int).SetBytes(targetBytes)
		target.Neg(target)
		return target
	}

	return new(big.Int).SetBytes(targetBytes)
}

// GetDisplayHash returns the bytes of a block hash in big-endian order.
func (hdr *BlockHeader) GetDisplayHash(height int) []byte {
	if hdr.cachedHash != nil {
		return hdr.cachedHash
	}

	serializedHeader, err := hdr.MarshalBinary()
	if err != nil {
		return nil
	}

	// VerusHash
	hash := make([]byte, 32)
	ptrHash := uintptr(unsafe.Pointer(&hash[0]))
	ParserLog.Warn("block_header.GetDisplayHash() calling Anyverushash_reverse_height()", height)
	VerusHash.Anyverushash_reverse_height(string(serializedHeader), len(string(serializedHeader)), ptrHash, height)

	hdr.cachedHash = hash
	return hdr.cachedHash
}

// GetEncodableHash returns the bytes of a block hash in little-endian wire order.
func (hdr *BlockHeader) GetEncodableHash(height int) []byte {
	serializedHeader, err := hdr.MarshalBinary()

	if err != nil {
		return nil
	}

	hash := make([]byte, 32)
	ptrHash := uintptr(unsafe.Pointer(&hash[0]))
	ParserLog.Warn("block_header.GetEncodableHash()  at height ", height)
	VerusHash.Anyverushash_height(string(serializedHeader), len(string(serializedHeader)), ptrHash, height)
	ParserLog.Warn("block_header.GetEncodableHash() called Anyverushash_height() which returned ", hash)
	return hash
}

// GetDisplayPrevHash gets the previous block's hash from this blocks header
func (hdr *BlockHeader) GetDisplayPrevHash() []byte {
	rhash := make([]byte, len(hdr.HashPrevBlock))
	copy(rhash, hdr.HashPrevBlock)
	// Reverse byte order
	for i := 0; i < len(rhash)/2; i++ {
		j := len(rhash) - 1 - i
		rhash[i], rhash[j] = rhash[j], rhash[i]
	}
	return rhash
}
