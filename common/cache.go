// Package common includes logging, certs, caching, and the ingestor (in common.go)
// Copyright (c) 2019-2020 The Zcash developers
// Forked and modified for the VerusCoin chain
// Copyright 2020 the VerusCoin developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package common

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"

	"github.com/asherda/lightwalletd/walletrpc"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-redis/redis/v7"
	"github.com/golang/protobuf/proto"
)

type blockCacheEntry struct {
	data []byte
	hash []byte
}

// BlockCache contains a set of recent compact blocks in marshalled form.
type BlockCache struct {
	chainName  string
	maxEntries int
	// m[firstBlock..nextBlock) are valid
	m           map[int]*blockCacheEntry
	firstBlock  int
	nextBlock   int
	blockHeight int
	RpcClient   *rpcclient.Client
	RedisClient *redis.Client
	mutex       sync.RWMutex
}

// NewBlockCache returns an instance of a block cache object.
func NewBlockCache(chainName string, maxEntries int, startHeight int, blockHeight int, rpcClient *rpcclient.Client, redisOptions *redis.Options) (*BlockCache, error) {
	redisClient, err := GetCheckedRedisClient(redisOptions)
	if err != nil {
		return nil, err
	}
	return &BlockCache{
		chainName:   chainName,
		maxEntries:  maxEntries,
		m:           make(map[int]*blockCacheEntry),
		firstBlock:  startHeight,
		nextBlock:   startHeight,
		blockHeight: blockHeight,
		RpcClient:   rpcClient,

		RedisClient: redisClient,
	}, nil
}

// Add adds the given block to the cache at the given height, returning true if a reorg was detected.
func (c *BlockCache) Add(height int, block *walletrpc.CompactBlock) (bool, error) {
	// Invariant: m[firstBlock..nextBlock) are valid.
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if height > c.nextBlock {
		// restarting the cache (never happens currently), or first time
		for i := c.firstBlock; i < c.nextBlock; i++ {
			delete(c.m, i)
		}
		c.firstBlock = height
		c.nextBlock = height
	}
	// Invariant: m[firstBlock..nextBlock) are valid.

	// If we already have this block, a reorg must have occurred;
	// TODO: check for identical blocks. We want to allow redundant ingestors, so it may already exist & that's OK, not an auto reorg.
	// this block (and all higher) must be re-added.
	h := height
	if h < c.firstBlock {
		h = c.firstBlock
	}
	for i := h; i < c.nextBlock; i++ {
		delete(c.m, i)
	}
	c.nextBlock = height
	if c.firstBlock > c.nextBlock {
		c.firstBlock = c.nextBlock
	}
	// Invariant: m[firstBlock..nextBlock) are valid.
	// Detect reorg, ingestor needs to handle it
	if height > c.firstBlock && !bytes.Equal(block.PrevHash, c.m[height-1].hash) {
		return true, nil
	}

	// Add the entry and update the counters
	data, err := proto.Marshal(block)
	if err != nil {
		return false, err
	}
	c.m[height] = &blockCacheEntry{
		data: data,
		hash: block.GetHash(),
	}
	c.nextBlock++
	// Invariant: m[firstBlock..nextBlock) are valid.

	UpdateRedisBlockAndDetails(c.RedisClient, c.chainName, height, data)

	// remove any blocks that are older than the capacity of the cache
	for c.firstBlock < c.nextBlock-c.maxEntries {
		// Invariant: m[firstBlock..nextBlock) are valid.
		delete(c.m, c.firstBlock)
		c.firstBlock++
	}
	// Invariant: m[firstBlock..nextBlock) are valid.

	return false, nil
}

// Get returns the compact block at the requested height if it is
// in the cache, else nil.
func (c *BlockCache) Get(height int) *walletrpc.CompactBlock {

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	triedRedis := false

	// only start with redis if in last 20-ish blocks to catch reorgs
	if height > c.blockHeight-19 {
		triedRedis = true
		compactBlockPtr := GetCompressedBlockFromRedis(c.RedisClient, c.chainName, height)
		if compactBlockPtr != nil {
			updatedHeight := getRedisCachedBlockHeight(c.RedisClient, c.chainName)
			if updatedHeight > c.blockHeight {
				c.blockHeight = updatedHeight
			}
			return compactBlockPtr
		}
		if c.RpcClient == nil {
			fmt.Println("Error getting compact block from redis at height ", strconv.Itoa(height), " with no verusd rpc client; unable to get a result, returning nil")
			return nil
		}
	}

	serialized := &walletrpc.CompactBlock{}
	unmarshalErr := proto.Unmarshal(c.m[height].data, serialized)
	if unmarshalErr == nil {
		return serialized
	}
	fmt.Println("Error unmarshalling compact block")

	if !triedRedis {
		return GetCompressedBlockFromRedis(c.RedisClient, c.chainName, height)
	}
	return nil
}

// GetLatestHeight returns the block with the greatest height, or nil
// if the cache is empty.
func (c *BlockCache) GetLatestHeight() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.firstBlock == c.nextBlock {
		return -1
	}
	return c.nextBlock - 1
}
