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
	m              *sync.Map
	newest, oldest *cacheNode
}

func NewLRU(maxSize int) *LRU {
	return &LRU{
		maxSize: maxSize,
		m:       &sync.Map{},
	}
}

func (c *LRU) Get(key string) ([]byte, bool) {
	return c.getAndMoveToFront(key)
}

func (c *LRU) getAndMoveToFront(key string) ([]byte, bool) {
	val, exists := c.m.Load(key)
	if exists {
		node := val.(*cacheNode)
		c.moveExistingToFront(node)
		return node.value, true
	}
	return nil, false
}
func (c *LRU) moveExistingToFront(node *cacheNode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.removeNodeFromLLNoCountChange(node)
	c.addToFrontNoCountChange(node)
}
func (c *LRU) Add(key string, value []byte) (overwritten bool) {
	val, exists := c.m.Load(key)
	//existingNode, exists := c.m[key]
	if exists {
		existingNode := val.(*cacheNode)
		existingNode.value = value
		c.moveExistingToFront(existingNode)
		return true
	}
	node := &cacheNode{
		key:   key,
		value: value,
	}

	c.m.Store(key, node)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.addToFront(node)
	return false
}
func (c *LRU) Evict(key string) (evicted bool) {
	evicted = false
	val, exists := c.m.Load(key)
	if exists {
		existingNode := val.(*cacheNode)
		c.mu.Lock()
		evicted = c.removeNodeFromLL(existingNode)
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
}

// removeNodeFromLL removes a node from the linked list ONLY
func (c *LRU) removeNodeFromLL(node *cacheNode) (removed bool) {
	if c == nil {
		panic("lru cache is nil")
	}
	if node == nil {
		return false
	}
	if c.ct == 1 {
		if node == c.oldest {
			c.oldest, c.newest, c.ct = nil, nil, 0
			return true
		} else {
			println("node was not in lru cache")
			return false
		}
	}
	older, newer := node.older, node.newer
	setOlderNewerPair(older, newer)
	switch node {
	case c.oldest:
		c.oldest = newer
		break
	case c.newest:
		c.newest = older
		break
	default:
		older.newer, newer.older = newer, older
		break
	}
	c.ct--
	return true
}
func (c *LRU) removeNodeFromLLNoCountChange(existingNode *cacheNode) {
	if c == nil {
		panic("lru cache is nil")
	}
	if existingNode == nil {
		return
	}
	if c.ct == 1 {
		if existingNode == c.oldest {
			c.oldest, c.newest = nil, nil
		} else {
			println("node was not in lru cache")
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
		older.newer, newer.older = newer, older
		break
	}
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
