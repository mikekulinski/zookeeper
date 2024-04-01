package znode

import (
	"fmt"
	"strings"
	"sync"

	pbzk "github.com/mikekulinski/zookeeper/proto"
)

// DB is the source of truth for all the data stored in the Zookeeper server. It also controls the
// locking mechanism, so it can be abstracted away from the caller.
type DB struct {
	// TODO: Consider using a map to each node.
	root *ZNode
	mu   *sync.RWMutex
}

func NewDB() *DB {
	return &DB{
		root: NewZNode("", ZNodeType_STANDARD, "", nil),
		mu:   &sync.RWMutex{},
	}
}

func (d *DB) Get(path string) *ZNode {
	d.mu.RLock()
	defer d.mu.RUnlock()

	names := splitPathIntoNodeNames(path)

	return findZNode(d.root, names)
}

// findZNode will search down to the tree and return the node specified by the names.
// If the node could not be found, then we will return nil.
func findZNode(start *ZNode, names []string) *ZNode {
	node := start
	for _, name := range names {
		z, ok := node.Children[name]
		if !ok {
			return nil
		}
		node = z
	}
	return node
}

func splitPathIntoNodeNames(path string) []string {
	// Since we have a leading /, then we expect the first name to be empty.
	return strings.Split(path, "/")[1:]
}

func (d *DB) Create(txn *pbzk.Transaction) (*ZNode, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	names := splitPathIntoNodeNames(txn.GetCreate().GetPath())
	// Search down the tree until we hit the parent where we'll be creating this new node.
	parent := findZNode(d.root, names[:len(names)-1])
	if parent == nil {
		return nil, fmt.Errorf("at least one of the anscestors of this node are missing")
	}
	if parent.NodeType == ZNodeType_EPHEMERAL {
		return nil, fmt.Errorf("ephemeral nodes cannot have children")
	}

	// We are at the parent node of the one we are trying to create. Now let's
	// try to create it.
	newName := names[len(names)-1]
	if txn.GetCreate().GetSequential() {
		newName = fmt.Sprintf("%s_%d", newName, parent.NextSequentialNode)
	}
	fullName := newFullName(newName, names[:len(names)-1])

	nodeType := ZNodeType_STANDARD
	if txn.GetCreate().GetEphemeral() {
		nodeType = ZNodeType_EPHEMERAL
	}

	newNode := NewZNode(
		fullName,
		nodeType,
		txn.GetClientId(),
		txn.GetCreate().GetData(),
	)

	if _, ok := parent.Children[newName]; ok {
		return nil, fmt.Errorf("node [%s] already exists at path [%s]", newName, txn.GetCreate().GetPath())
	}
	parent.Children[newName] = newNode
	// Make sure to increment the counter so the next sequential node will have the next number.
	if txn.GetCreate().GetSequential() {
		parent.NextSequentialNode++
	}
	return newNode, nil
}

func newFullName(nodeName string, ancestorsNames []string) string {
	nodePath := "/" + nodeName
	if len(ancestorsNames) > 0 {
		return "/" + strings.Join(ancestorsNames, "/") + nodePath
	}
	return nodePath
}

func (d *DB) Delete(txn *pbzk.Transaction) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	names := splitPathIntoNodeNames(txn.GetDelete().GetPath())

	// Search down the tree until we hit the parent where we'll be creating this new node.
	parent := findZNode(d.root, names[:len(names)-1])
	if parent == nil {
		return fmt.Errorf("at least one of the anscestors of this node are missing")
	}

	nameToDelete := names[len(names)-1]
	// Delete the actual node from the tree.
	delete(parent.Children, nameToDelete)
	return nil
}
