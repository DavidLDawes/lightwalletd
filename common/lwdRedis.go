// Package common includes logging, certs, caching, and the ingestor (in common.go)
// Copyright (c) 2019-2020 The Zcash developers
// Forked and modified for the VerusCoin chain
// Copyright 2020 the VerusCoin developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package common

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/asherda/lightwalletd/walletrpc"
	"github.com/go-redis/redis/v7"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// GetCheckedRedisClient get a redis client using the URL passed in & check ping works or return nil if URL is empty.
func GetCheckedRedisClient(redisOpts *redis.Options) (*redis.Client, error) {
	var redisClient *redis.Client
	if len(redisOpts.Addr) > 0 {
		redisClient = redis.NewClient(redisOpts)
		_, err := redisClient.Ping().Result()
		if err != nil {
			return redisClient, errors.New("unable to ping redis server at" + redisOpts.Addr)
		}
	}
	return redisClient, nil
}

// UpdateRedisBlockAndDetails updates redis for the new block, updates redis chain settings (if redis is enabled)
func UpdateRedisBlockAndDetails(redisClient *redis.Client, chainName string, height int, data []byte) {
	if redisClient != nil {
		blockBase64 := base64.StdEncoding.EncodeToString(data)
		redisErr := redisClient.Set(chainName+"-"+strconv.Itoa(height), blockBase64, 0).Err()
		if redisErr != nil {
			fmt.Println("Warning: Error writing to redis")
		} else {
			updateRedisCache(*redisClient, chainName+"-blockHeight", height)
			updateRedisCache(*redisClient, chainName+"-cachedBlockHeight", height)
		}
	}
}

// GetCompressedBlockFromRedis pulls the requested block from redis, if it is available
func GetCompressedBlockFromRedis(redisClient *redis.Client, chainName string, height int) *walletrpc.CompactBlock {
	redisCache, err := redisClient.Get(chainName + "-" + strconv.Itoa(height)).Result()
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
	return nil
}

func updateRedisCache(redisC redis.Client, key string, value int) {
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

// UpdateRedisValues sets the values on the redis cache for the chain that are not per block: saplingHeight, blockHeight, chainName and branchID
func UpdateRedisValues(redisClient *redis.Client, saplingHeight int, blockHeight int, chainName string, subChainName string, branchID string) int {
	if redisClient != nil {
		redisCacheIntSet(redisClient, chainName+"-saplingHeight", saplingHeight)
		redisCacheIntSet(redisClient, chainName+"-blockHeight", blockHeight)
		redisCacheStringSet(redisClient, chainName+"-branchID", branchID)
		redisCacheStringSet(redisClient, chainName+"-subChain", subChainName)
		return getRedisCachedBlockHeight(redisClient, chainName)
	}
	return 0
}

func getRedisCachedBlockHeight(redisClient *redis.Client, chainName string) int {
	resultString, err := redisClient.Get(chainName + "-cachedBlockHeight").Result()
	if err != nil {
		fmt.Println("Error reading cachedBlockHeight from redis")
		return 0
	}
	resultInt, convErr := strconv.Atoi(resultString)
	if convErr != nil {
		fmt.Println("Error converting cachedBlockHeight string from redis to int")
		return 0
	}
	return resultInt
}

// CheckRedisIntResult gets an int from redis, checking for errors
func CheckRedisIntResult(redisClient *redis.Client, key string) int {
	valueString := CheckRedisStringResult(redisClient, key)
	result, err := strconv.Atoi(valueString)
	if err != nil {
		giveUp(key)
	}
	return result
}

// CheckRedisStringResult gets a string value from redis, checking for errors
func CheckRedisStringResult(redisClient *redis.Client, key string) string {
	result, err := redisClient.Get(key).Result()
	if err != nil {
		giveUp(key)
	}
	return result
}

func redisCacheIntSet(redisClient *redis.Client, key string, value int) {
	redisErr := redisClient.Set(key, value, 0).Err()
	if redisErr != nil {
		fmt.Println("Error writing ", key, " with value ", value, "to redis")
	}
}

func redisCacheStringSet(redisClient *redis.Client, key string, value string) {
	redisErr := redisClient.Set(key, value, 0).Err()
	if redisErr != nil {
		fmt.Println("Error writing ", key, " with value ", value, "to redis")
	}
}

func complainInt(key string, value int) {
	fmt.Println("Error writing ", key, " with value ", value, "to redis")
}

func complainString(key string, value string) {
	fmt.Println("Error writing ", key, " with value ", value, "to redis")
}

func giveUp(key string) {
	os.Stderr.WriteString(fmt.Sprintf("\n ** redis is enabled but lightwalletd is unable to fetch %s %s", key,
		" from redis - you must run at least one ingestor (lightwalletd without --only-redis) and get the redis cache setup before running lisghtwalletd --no-verusd\n\n"))
	os.Exit(1)
}
