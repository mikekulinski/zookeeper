package znode

import (
	pbzk "github.com/mikekulinski/zookeeper/proto"
)

type ZNodeType int

const (
	ZNodeType_STANDARD ZNodeType = iota
	ZNodeType_EPHEMERAL
)

type ZNode struct {
	// ZNode metadata.
	// Name is the full name of the ZNode from the root of the tree.
	Name               string
	Version            int64
	Children           map[string]*ZNode
	NodeType           ZNodeType
	NextSequentialNode int
	// Creator is the ClientID of who created this node. This is helpful when working with ephemeral nodes.
	Creator string

	// Data is the data stored here by the client.
	Data []byte
}

func NewZNode(name string, nodeType ZNodeType, creator string, data []byte) *ZNode {
	return &ZNode{
		Name:    name,
		Version: 0,
		// Init the children to an empty map instead of nil to avoid panics when writing to
		// a nil map.
		Children: map[string]*ZNode{},
		NodeType: nodeType,
		Creator:  creator,
		Data:     data,
	}
}

type Watch struct {
	ClientID   string
	Path       string
	WatchTypes []pbzk.WatchEvent_EventType
}
