package znode

type ZNodeType int

const (
	ZNodeType_STANDARD ZNodeType = iota
	ZNodeType_EPHEMERAL
)

type ZNode struct {
	// ZNode metadata.
	Name               string
	Version            int
	Children           map[string]*ZNode
	NodeType           ZNodeType
	NextSequentialNode int

	// Data is the data stored here by the client.
	Data []byte
}

func NewZNode(name string, version int, nodeType ZNodeType, data []byte) *ZNode {
	return &ZNode{
		Name:    name,
		Version: version,
		// Init the children to an empty map instead of nil to avoid panics when writing to
		// a nil map.
		Children: map[string]*ZNode{},
		NodeType: nodeType,
		Data:     data,
	}
}