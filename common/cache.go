package common

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/asherda/lightwalletd/walletrpc"
	"github.com/golang/protobuf/proto"
)

type blockCacheEntry struct {
	data []byte
}

// BlockCache contains a consecutive set of recent compact blocks in marshalled form.
type BlockCache struct {
	lengthsName, blocksName string // pathnames
	lengthsFile, blocksFile *os.File
	starts                  []int64 // Starting offset of each block within blocksFile
	FirstBlock              int     // height of the first block in the cache (usually Sapling activation)
	NextBlock               int     // height of the first block not in the cache
	LatestHash              []byte  // hash of the most recent (highest height) block, for detecting reorgs.
	mutex                   sync.RWMutex
}

func (c *BlockCache) blockLength(height int) int {
	index := height - c.FirstBlock
	return int(c.starts[index+1] - c.starts[index])
}

// NewBlockCache returns an instance of a block cache object.
// Note if you call this with a different height than was used
// to create the db files, first delete those files (fixme)
func NewBlockCache(chainName string, startHeight int) *BlockCache {
	c := &BlockCache{}
	c.FirstBlock = startHeight
	c.NextBlock = startHeight
	c.lengthsName, c.blocksName = fileNames(chainName)
	var err error
	c.blocksFile, err = os.OpenFile(c.blocksName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		Log.Fatal("open ", c.blocksName, " failed: ", err)
	}
	c.lengthsFile, err = os.OpenFile(c.lengthsName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		Log.Fatal("open ", c.lengthsName, " failed: ", err)
	}
	lengths, err := ioutil.ReadFile(c.lengthsName)
	if err != nil {
		Log.Fatal("read ", c.lengthsName, " failed: ", err)
	}
	// the last entry in starts[] is where to write the next block
	var offset int64
	c.starts = append(c.starts, 0)
	for i := 0; i < len(lengths)/4; i++ {
		length := binary.LittleEndian.Uint32(lengths[i*4 : i*4+4])
		offset += int64(length)
		c.starts = append(c.starts, offset)
		c.NextBlock++
	}
	if c.NextBlock > c.FirstBlock {
		// There is at least one block; get the last block's hash
		b := make([]byte, c.blockLength(c.NextBlock-1))
		_, err := c.blocksFile.ReadAt(b, c.starts[c.NextBlock-c.FirstBlock-1])
		if err != nil {
			Log.Fatal("blocks read failed: ", err)
		}
		block := &walletrpc.CompactBlock{}
		err = proto.Unmarshal(b, block)
		if err != nil {
			println("Error unmarshalling compact block")
			return nil
		}
		c.LatestHash = make([]byte, len(block.Hash))
		copy(c.LatestHash, block.Hash)
	}
	return c
}

func fileNames(chainName string) (string, string) {
	return fmt.Sprintf("db-%s-lengths", chainName),
		fmt.Sprintf("db-%s-blocks", chainName)
}

// only used for testing to ensure we're not using files from a previous run
func CacheTestClean(chainName string) {
	lengthsName, blockName := fileNames(chainName)
	os.Remove(lengthsName)
	os.Remove(blockName)
}

// Add adds the given block to the cache at the given height, returning true
// if a reorg was detected.
func (c *BlockCache) Add(block *walletrpc.CompactBlock) error {
	// Invariant: m[FirstBlock..NextBlock) are valid.
	c.mutex.Lock()
	defer c.mutex.Unlock()

	height := int(block.Height)
	if height != c.NextBlock {
		Log.Fatalf("bug, cache.Add non-consecutive blocks: height: %v, ccache: %+v", height, c)
	}

	// Add the new block and its length to the db files
	data, err := proto.Marshal(block)
	if err != nil {
		return err
	}
	_, err = c.blocksFile.Write(data)
	if err != nil {
		Log.Fatal("blocks write failed: ", err)
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(len(data)))
	_, err = c.lengthsFile.Write(b)
	if err != nil {
		Log.Fatal("lengths write failed: ", err)
	}

	// update the in-memory variables
	offset := c.starts[len(c.starts)-1]
	c.starts = append(c.starts, offset+int64(len(data)))

	if c.LatestHash == nil {
		c.LatestHash = make([]byte, len(block.Hash))
	}
	copy(c.LatestHash, block.Hash)
	c.NextBlock++
	// Invariant: m[FirstBlock..NextBlock) are valid.
	return nil
}

// Reorg resets NextBlock (the block that should be Add()ed next) to the given
// height, taking care of all the details. It returns the height that should be
// added next (different if less than FirstBlock).
func (c *BlockCache) Reorg(height int) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if height < c.FirstBlock {
		height = c.FirstBlock
	}
	if height > c.NextBlock {
		Log.Fatalf("bug, cache.Reorg moves ahead: height: %v, ccache: %+v", height, c)
	}
	// Remove the end of the cache.
	c.NextBlock = height
	newCacheLen := height - c.FirstBlock
	c.starts = c.starts[:newCacheLen+1]

	if err := c.lengthsFile.Truncate(int64(4 * newCacheLen)); err != nil {
		Log.Fatal("truncate failed: ", err)
	}
	if err := c.blocksFile.Truncate(c.starts[newCacheLen]); err != nil {
		Log.Fatal("truncate failed: ", err)
	}
	c.Sync()
	c.LatestHash = nil
	if c.NextBlock > c.FirstBlock {
		// There is at least one block; get the last block's hash
		b := make([]byte, c.blockLength(c.NextBlock-1))
		_, err := c.blocksFile.ReadAt(b, c.starts[c.NextBlock-c.FirstBlock-1])
		if err != nil {
			Log.Fatal("blocks read failed: ", err)
		}
		block := &walletrpc.CompactBlock{}
		err = proto.Unmarshal(b, block)
		if err != nil {
			Log.Warn("Error unmarshalling compact block: ", err)
			return height
		}
		c.LatestHash = make([]byte, 32)
		copy(c.LatestHash, block.Hash)
	}
	return height
}

// Get returns the compact block at the requested height if it is
// in the cache, else nil.
func (c *BlockCache) Get(height int) *walletrpc.CompactBlock {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if height < c.FirstBlock || height >= c.NextBlock {
		return nil
	}

	b := make([]byte, c.blockLength(height))
	_, err := c.blocksFile.ReadAt(b, c.starts[height-c.FirstBlock])
	if err != nil {
		Log.Fatal("blocks read failed: ", err)
	}
	block := &walletrpc.CompactBlock{}
	err = proto.Unmarshal(b, block)
	if err != nil {
		println("Error unmarshalling compact block")
		return nil
	}
	return block
}

// GetLatestHeight returns the block with the greatest height, or nil
// if the cache is empty.
func (c *BlockCache) GetLatestHeight() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.FirstBlock == c.NextBlock {
		return -1
	}
	return c.NextBlock - 1
}

func (c *BlockCache) Sync() {
	c.lengthsFile.Sync()
	c.blocksFile.Sync()
}

// Currently used only for testing.
func (c *BlockCache) Close() {
	// Some operating system require you to close files before you can remove them.
	if c.lengthsFile != nil {
		c.lengthsFile.Close()
		c.lengthsFile = nil
	}
	if c.blocksFile != nil {
		c.blocksFile.Close()
		c.blocksFile = nil
	}
}
