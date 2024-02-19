package pkg

type ZNodeType int

const (
	ZNodeType_STANDARD ZNodeType = iota
	ZNodeType_EPHEMERAL
)

type ZNode struct {
	// ZNode metadata.
	name     string
	version  int
	children map[string]*ZNode
	nodeType ZNodeType

	// data is the data stored here by the client.
	data []byte
}

func NewZNode(name string, version int, nodeType ZNodeType, data []byte) *ZNode {
	return &ZNode{
		name:    name,
		version: version,
		// Init the children to an empty map instead of nil to avoid panics when writing to
		// a nil map.
		children: map[string]*ZNode{},
		nodeType: nodeType,
		data:     data,
	}
}
