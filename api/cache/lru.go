package cache

import "sync"

//type LruLockless struct {
//	ct             int
//	maxSize        int
//	m              map[string]*cacheNode
//	newest, oldest *cacheNode
//}

type LRU struct { // TODO: EXPAND USAGE OF LRU CACHE TO OTHER THINGS, OR CONSIDER USING REDIS (PROBABLY REDIS)
	mu             sync.Mutex
	ct             int
	maxSize        int
	m              sync.Map
	newest, oldest *cacheNode
	moveToFront    chan<- *cacheNode
	addNode        chan<- *cacheNode
	// TODO: channel for adding?
}

func NewLRU(maxSize int) *LRU {
	frontChan := make(chan *cacheNode, maxSize) // TODO: chan size ok?
	addChan := make(chan *cacheNode, maxSize)   // TODO: chan size ok?
	c := &LRU{
		maxSize:     maxSize,
		m:           sync.Map{},
		moveToFront: frontChan, // TODO: ENSURE CHANNEL GETS CLOSED WHEN WE WANT IT TO BE
		addNode:     addChan,   // TODO: ENSURE CHANNEL GETS CLOSED WHEN WE WANT IT TO BE
	}
	go func() {
		for item := range addChan {
			c.mu.Lock() // TODO: should this go before store?
			// TODO: can we make it so that it only needs one channel?
			c.addToFront(item) // TODO: should store be done after addToFront so that there is no race condition?
			c.m.Store(item.key, item)
			c.mu.Unlock()
		}
	}()
	go func() {
		for item := range frontChan {
			c.moveExistingToFront(item)
		}
	}()
	return c
}

func (c *LRU) Get(key string) ([]byte, bool) {
	return c.getAndMoveToFront(key)
}

func (c *LRU) getAndMoveToFront(key string) ([]byte, bool) {
	val, exists := c.m.Load(key)
	if exists {
		node := val.(*cacheNode)
		c.moveToFront <- node // TODO: used to be c.moveExistingToFront(node)
		return node.value, true
	}
	return nil, false
}
func (c *LRU) moveExistingToFront(node *cacheNode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.removeNodeFromLLNoCountChange(node)
	c.addToFrontNoCountChange(node)
}
func (c *LRU) Add(key string, value []byte) (overwritten bool) {
	val, exists := c.m.Load(key)
	if exists {
		existingNode := val.(*cacheNode)
		existingNode.value = value
		c.moveToFront <- existingNode // TODO: used to be c.moveExistingToFront(existingNode)
		return true
	}
	node := &cacheNode{
		key:   key,
		value: value,
	}

	c.addNode <- node
	return false
}
func (c *LRU) Evict(key string) (evicted bool) {
	evicted = false
	val, exists := c.m.Load(key)
	if exists {
		c.mu.Lock()
		evicted = c.removeNodeFromLL(val.(*cacheNode))
		c.mu.Unlock()
	}
	c.m.Delete(key)
	return evicted
}

// addToFront adds a node to the front, and evicts the earliest if needed
func (c *LRU) addToFront(node *cacheNode) {
	if node == nil {
		return
	}
	setOlderNewerPair(c.newest, node)
	c.newest = node
	if c.ct >= c.maxSize {
		currentOldest := c.oldest
		newOldest := currentOldest.newer
		newOldest.setOlder(nil)
		c.oldest = newOldest
		c.m.Delete(currentOldest.key)
	} else {
		if c.ct == 0 {
			c.oldest = node
		}
		c.ct++
	}
}
func (c *LRU) addToFrontNoCountChange(node *cacheNode) {
	setOlderNewerPair(c.newest, node)
	c.newest = node
	if c.ct == 1 {
		c.oldest = node
	}
}

// removeNodeFromLL removes a node from the linked list ONLY
func (c *LRU) removeNodeFromLL(node *cacheNode) (removed bool) {
	removed = c.removeNodeFromLLNoCountChange(node)
	if removed {
		c.ct--
	}
	return removed
}
func (c *LRU) removeNodeFromLLNoCountChange(existingNode *cacheNode) (removed bool) {
	if c == nil {
		panic("lru cache is nil")
	}
	if existingNode == nil {
		return false
	}
	if c.ct == 0 {
		return false
	}
	if c.ct == 1 {
		if existingNode == c.oldest {
			c.oldest, c.newest = nil, nil
			return true
		} else {
			println("node was not in lru cache")
			return false
		}
		return
	}
	older, newer := existingNode.older, existingNode.newer
	setOlderNewerPair(older, newer)
	switch existingNode {
	case c.oldest:
		c.oldest = newer
		break
	case c.newest:
		c.newest = older
		break
	default:
		// Do nothing
		// TODO: ensure we dont need next line. I think it is covered by setOlderNewerPair...
		// older.newer, newer.older = newer, older
		break
	}
	return true
}
func (n *cacheNode) setNewer(node *cacheNode) {
	if n == nil {
		return
	}
	n.newer = node
}
func (n *cacheNode) setOlder(node *cacheNode) {
	if n == nil {
		return
	}
	n.older = node
}
func setOlderNewerPair(older, newer *cacheNode) {
	older.setNewer(newer)
	newer.setOlder(older)
}

type cacheNode struct {
	key          string
	value        []byte
	newer, older *cacheNode
}
