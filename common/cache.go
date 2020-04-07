// Package common includes logging, certs, caching, and the ingestor (in common.go)
// Copyright (c) 2019-2020 The Zcash developers
// Forked and modified for the VerusCoin chain
// Copyright 2020 the VerusCoin developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package common

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/asherda/lightwalletd/walletrpc"
	"github.com/go-redis/redis/v7"
	"github.com/golang/protobuf/proto"
)

type blockCacheEntry struct {
	data []byte
	hash []byte
}

// BlockCache contains a set of recent compact blocks in marshalled form.
type BlockCache struct {
	MaxEntries int

	// m[firstBlock..nextBlock) are valid
	m           map[int]*blockCacheEntry
	firstBlock  int
	nextBlock   int
	RedisClient redis.Client
	mutex       sync.RWMutex
}

// NoVerusd flag if true indicates we do not use verusd, so only Redis will be queried and it is read only, no transactions
var NoVerusd = false

// ChainName string the chain name, for VerusCoin it is VRSC
var ChainName = ""

// RedisURL string if set this is the URL including the port for redis, if empty redis is diasbled; if you run redis locally by default it uses 127.0.0.1:6379 so that should work
var RedisURL = ""

// RedisPassword is the optional password needed to access redis over tcp
var RedisPassword = ""

// RedisDB int the oprional DB number for redis
var RedisDB = 0

// NoRedis flag if true indicates we do not use verusd, so only verusd will be used; all features available and in memory cache only so startup is slowish
var NoRedis = false

// NewBlockCache returns an instance of a block cache object.
func NewBlockCache(maxEntries int, startHeight int) *BlockCache {
	var RedisClient *redis.Client

	if !NoRedis {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     RedisURL,
			Password: RedisPassword, // no password set
			DB:       RedisDB,       // use default DB
		})
		_, err := RedisClient.Ping().Result()
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("\n  ** redis is enabled but lightwalletd is unable to connect to the redis host\n\n"))
			os.Exit(1)
		}
		return &BlockCache{
			MaxEntries:  maxEntries,
			m:           make(map[int]*blockCacheEntry),
			firstBlock:  startHeight,
			nextBlock:   startHeight,
			RedisClient: *RedisClient,
		}
	}
	return &BlockCache{
		MaxEntries: maxEntries,
		m:          make(map[int]*blockCacheEntry),
		firstBlock: startHeight,
		nextBlock:  startHeight,
	}
}

// Add adds the given block to the cache at the given height, returning true
// if a reorg was detected.
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

	if !NoRedis {
		blockBase64 := base64.StdEncoding.EncodeToString(data)
		redisErr := c.RedisClient.Set(ChainName+strconv.Itoa(height), blockBase64, 0).Err()
		if redisErr != nil {
			fmt.Println("Warning: Error writing to redis")
		} else {
			updateCache(c.RedisClient, ChainName+"blockHeight", height)
			updateCache(c.RedisClient, ChainName+"cachedBlockHeight", height)
		}
	}

	// remove any blocks that are older than the capacity of the cache
	for c.firstBlock < c.nextBlock-c.MaxEntries {
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

	if !NoRedis {
		redisCache, err := c.RedisClient.Get(ChainName + strconv.Itoa(height)).Result()
		if err == nil {
			decoded, decodeErr := base64.StdEncoding.DecodeString(redisCache)
			if decodeErr == nil {
				serialized := &walletrpc.CompactBlock{}
				redisUnmarshalErr := proto.Unmarshal(decoded, serialized)
				if redisUnmarshalErr == nil {
					return serialized
				}
			}
		}
		if NoVerusd {
			fmt.Println("Error unmarshalling compact block")
			return nil
		}
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if height < c.firstBlock || height >= c.nextBlock {
		return nil
	}

	serialized := &walletrpc.CompactBlock{}
	unmarshalErr := proto.Unmarshal(c.m[height].data, serialized)
	if unmarshalErr != nil {
		fmt.Println("Error unmarshalling compact block")
		return nil
	}

	return serialized
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

func updateCache(redisC redis.Client, key string, value int) {
	redisCacheValueString, err := redisC.Get(key).Result()
	if err != nil {
		fmt.Println("Warning: Unable to read redis ", key, ", creating it with value  ", strconv.Itoa(value), " using \"0\"")
		redisCacheValueString = "0"
	}
	redisCacheValue, err := strconv.Atoi(redisCacheValueString)
	if err != nil {
		fmt.Println("Warning: Unable to convert cached redis ", key, " with value \"", redisCacheValueString, "\" from string to int, using 0")
		redisCacheValue = 0
	}
	if value > redisCacheValue {
		redisErr := redisC.Set(key, strconv.Itoa(value), 0).Err()
		if redisErr != nil {
			fmt.Println("Warning: Unable to set redis ", key, " to ", strconv.Itoa(value), " stuck at ", redisCacheValue)
		}
	}
}
