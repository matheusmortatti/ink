package render

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"sync"

	"github.com/matheusmortatti/ink/internal/block"
)

// RenderCache stores rendered block output keyed by content hash and width.
type RenderCache struct {
	entries map[string]string
	mu      sync.RWMutex
}

// NewRenderCache creates an empty render cache.
func NewRenderCache() *RenderCache {
	return &RenderCache{entries: make(map[string]string)}
}

// Get retrieves cached rendering if available.
func (c *RenderCache) Get(b block.Block, width int) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := cacheKey(hashContent(b.Raw), width)
	val, ok := c.entries[key]
	return val, ok
}

// Put stores a rendered block in the cache.
func (c *RenderCache) Put(b block.Block, width int, rendered string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := cacheKey(hashContent(b.Raw), width)
	c.entries[key] = rendered
}

// InvalidateAll clears the entire cache (for terminal resize).
func (c *RenderCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]string)
}

// InvalidateBlock clears all cache entries for a specific block.
// Removes entries at all widths for the given block content.
func (c *RenderCache) InvalidateBlock(b block.Block) {
	c.mu.Lock()
	defer c.mu.Unlock()
	h := hashContent(b.Raw)
	prefix := h + "_"
	for key := range c.entries {
		if strings.HasPrefix(key, prefix) {
			delete(c.entries, key)
		}
	}
}

// hashContent computes an FNV-1a hash of the raw content string.
// FNV-1a is chosen for speed over collision resistance, which is
// appropriate for cache keys where collisions are harmless.
func hashContent(raw string) string {
	h := fnv.New64a()
	h.Write([]byte(raw))
	return strconv.FormatUint(h.Sum64(), 16)
}

// cacheKey builds the cache key from content hash and width.
func cacheKey(contentHash string, width int) string {
	return fmt.Sprintf("%s_%d", contentHash, width)
}
