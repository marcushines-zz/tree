package tree

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openconfig/telemetry/ring"
)

// Tree implements a path tree for storage of openconfig paths. I
type Tree struct {
	mu       sync.Mutex
	edges    map[string]*Node
	index    map[string]*Node
	count    int64
	capacity int
}

// New returns a new initialized Tree.
func New(capacity int) *Tree {
	return &Tree{
		capacity: capacity,
		edges:    map[string]*Node{},
		index:    map[string]*Node{},
	}
}

// Update adds element e to the tree with path.  If key already exists then the value
// is updated. If key == "" or e == nil then no node will be created and nil is
// returned,
func (t *Tree) Update(path []string, v interface{}) *Node {
	t.mu.Lock()
	defer t.mu.Unlock()
	if v == nil {
		return nil
	}
	if len(path) == 0 {
		return nil
	}
	n, ok := t.index[key(path)]
	if !ok {
		n = NewNode([]string{}, path[0], nil)
		t.index[path[0]] = n
		t.edges[path[0]] = n
		for _, e := range path[1:] {
			n.edges[e] = NewNode(n.path, e, nil)
			n = n.edges[e]
			t.index[key(n.path)] = n
		}
	}
	n.Set(v)
	return n
}

// Delete removes path from the if the path is not present then it is is a noop.
// The deleted node will be returned or nil of not found.
func (t *Tree) Delete(path []string) *Node {
	t.mu.Lock()
	defer t.mu.Unlock()
	n, ok := t.index[key(path)]
	if !ok {
		return nil
	}
	delete(t.edges, path[0])
	for i := 0; i <= len(path); i++ {
		delete(t.index, key(path[:i]))
	}
	return n
}

// Get returns the value for path. If path is not found an error is returned.
func (t *Tree) Get(path []string) (*Node, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("'path' must not be empty or nil")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	n, ok := t.index[key(path)]
	if !ok {
		return nil, fmt.Errorf("path %s not found", path)
	}
	return n, nil
}

// Node is a node within the tree.  Values are stored in a ring buffer.
type Node struct {
	mu sync.Mutex
	// path stores the list of edges from the root of the tree to this node.
	path []string
	// edges stores all child edges of the node.
	edges map[string]*Node
	// v holds a ring buffer of the last n values stored in this node.
	v         *ring.Ring
	curr      interface{}
	timestamp time.Time
}

// NewNode returns a new node initialized to value v.
func NewNode(path []string, key string, v interface{}) *Node {
	n := &Node{
		v:         ring.New(100),
		timestamp: time.Now(),
		edges:     map[string]*Node{},
	}
	n.path = make([]string, len(path)+1)
	n.path[copy(n.path, path)] = key
	n.v.Enqueue(v)
	n.curr = v
	return n
}

// Set pushes value v into n.  If n is at capacity the oldest value of n will be
// dequeued.
func (n *Node) Set(v interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.v.Enqueue(v)
}

// Get peeks at the top value of n.
func (n *Node) Get() interface{} {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.v.Peek()
}

func key(path []string) string {
	if len(path) == 0 {
		return ""
	}
	return strings.Join(path, "^")
}
