package lfucache

import (
	"container/list"
	"sync"
)

// TODO: implement a minimum lfu lifetime,
// to give the item a chance to increase its frequency.
// The goal is to not remove recently added items.
// When the cache is full and the minimum lfu lifetime
// hasn't passed, do not add new items.

// TODO: implement cache stats

// TODO: use a custom linked list implementation
// to avoid casting of value to string

type LFUCache struct {
	lowerFreq int
	freqs     map[int]*list.List // list of keys
	cache     map[string]*cacheItem
	maxCount  int
	// TODO: remove the mutex. use a go routine,
	// to process the frequency increase/lfu removal
	mtx sync.Mutex
}

type cacheItem struct {
	value  any
	freq   int
	freqEl *list.Element
}

func New(maxCount int) *LFUCache {
	return &LFUCache{
		maxCount:  maxCount,
		lowerFreq: 0,
		freqs:     make(map[int]*list.List),
		cache:     make(map[string]*cacheItem),
	}
}

func (c *LFUCache) Add(key string, value any) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	cachedItem, ok := c.cache[key]

	if ok {
		c.increaseFreq(cachedItem, key)
		c.cache[key] = cachedItem
		return
	}

	if c.maxCount == len(c.cache) {
		c.removeLfu()
	}

	c.lowerFreq = 0

	zeroFreqKeys, ok := c.freqs[0]

	if !ok {
		c.freqs[0] = list.New()
		zeroFreqKeys = c.freqs[0]
	}

	cachedItem = &cacheItem{
		value: value,
		freq:  0,
	}
	c.cache[key] = cachedItem
	cachedItem.freqEl = zeroFreqKeys.PushBack(key)
}

func (c *LFUCache) increaseFreq(cachedItem *cacheItem, key string) {
	prevFreq := cachedItem.freq

	// increase the frequency on cached item
	cachedItem.freq++

	// remove the item from the previous frequency list
	prevFreqList := c.freqs[prevFreq]
	prevFreqList.Remove(cachedItem.freqEl)

	// if previous frequency list is empty, delete it from freq map
	if prevFreqList.Len() == 0 {
		delete(c.freqs, prevFreq)
	}

	// move the item to next frequency list
	nextFreqList, ok := c.freqs[cachedItem.freq]

	// if the next frequency doesn't exist, create it
	if !ok {
		c.freqs[cachedItem.freq] = list.New()
		nextFreqList = c.freqs[cachedItem.freq]
	}

	cachedItem.freqEl = nextFreqList.PushFront(key)

	// if the previous item frequency is equal to the lower frequency,
	// and the previous frequency list is empty
	// increase lower frequency
	if prevFreq == c.lowerFreq && prevFreqList.Len() == 0 {
		c.lowerFreq++
		return
	}

	// if the current item frequency is lower than the lower frequency,
	// item frequency is the new lower
	if cachedItem.freq < c.lowerFreq {
		c.lowerFreq = cachedItem.freq
	}
}

func (c *LFUCache) removeLfu() {
	// get the lfu freq list
	lfuList := c.freqs[c.lowerFreq]

	// get the lfu list element
	lfuEl := lfuList.Back()

	// remove it from the frequency list
	lfuList.Remove(lfuEl)

	// if the frequency list is empty
	// remove it from map
	if lfuList.Len() == 0 {
		delete(c.freqs, c.lowerFreq)
	}

	// remove the lfu item from cache
	delete(c.cache, lfuEl.Value.(string))
}

func (c *LFUCache) Get(key string) (any, bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	item, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	c.increaseFreq(item, key)
	return item.value, true
}

func (c *LFUCache) Count() int {
	return len(c.cache)
}

// GetAllKeys returns a channel with all keys stored
// on cache.
// The key may be deleted by concurrent access after
// being retrieved, so check if the cache item exists
// when using the key.
func (c *LFUCache) GetAllKeys() <-chan string {
	iterator := make(chan string, 1)
	go func() {
		for key := range c.cache {
			iterator <- key
		}
		close(iterator)
	}()
	return iterator
}
