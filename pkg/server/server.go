package server

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mikekulinski/zookeeper/pkg/znode"
)

type Zookeeper interface {
	// Create creates a ZNode with path name path, stores data in it, and returns the name of the new ZNode
	// Flags can also be passed to pick certain attributes you want the ZNode to have.
	Create(req *CreateReq, resp *CreateResp) error
	// Delete deletes the ZNode at the given path if that ZNode is at the expected version.
	Delete(req *DeleteReq, resp *DeleteResp) error
	// Exists returns true if the ZNode with path name path exists, and returns false otherwise. The watch flag
	// enables a client to set a watch on the ZNode.
	Exists(req *ExistsReq, resp *ExistsResp) error
	// GetData returns the data and metadata, such as version information, associated with the ZNode.
	// The watch flag works in the same way as it does for exists(), except that ZooKeeper does not set the watch
	// if the ZNode does not exist.
	GetData(req *GetDataReq, resp *GetDataResp) error
	// SetData writes data to the ZNode path if the version number is the current version of the ZNode.
	// TODO: What do we do if the version is invalid? Should we return some sort of error message?
	SetData(req *SetDataReq, resp *SetDataResp) error
	// GetChildren returns the set of names of the children of a ZNode.
	GetChildren(path string, watch bool) (childrenNames []string, err error)
	// Sync waits for all updates pending at the start of the operation to propagate to the server
	// that the client is connected to. The path is currently ignored. (Using path is not discussed in the white paper)
	Sync(path string)
}

type Server struct {
	root *znode.ZNode
}

func NewServer() *Server {
	return &Server{
		root: znode.NewZNode("", -1, znode.ZNodeType_STANDARD, nil),
	}
}

func (s *Server) Create(req *CreateReq, resp *CreateResp) error {
	err := validatePath(req.Path)
	if err != nil {
		return err
	}
	names := splitPathIntoNodeNames(req.Path)

	// Search down the tree until we hit the parent where we'll be creating this new node.
	parent := findZNode(s.root, names[:len(names)-1])
	if parent == nil {
		return fmt.Errorf("at least one of the anscestors of this node are missing")
	}
	if parent.NodeType == znode.ZNodeType_EPHEMERAL {
		return fmt.Errorf("ephemeral nodes cannot have children")
	}

	// We are at the parent node of the one we are trying to create. Now let's
	// try to create it.
	newName := names[len(names)-1]
	if slices.Contains(req.Flags, SEQUENTIAL) {
		newName = fmt.Sprintf("%s_%d", newName, parent.NextSequentialNode)
	}
	nodeType := znode.ZNodeType_STANDARD
	if slices.Contains(req.Flags, EPHEMERAL) {
		nodeType = znode.ZNodeType_EPHEMERAL
	}
	newNode := znode.NewZNode(
		newName,
		0,
		nodeType,
		req.Data,
	)

	if _, ok := parent.Children[newName]; ok {
		return fmt.Errorf("node [%s] already exists at path [%s]", newName, req.Path)
	}
	parent.Children[newName] = newNode
	// Make sure to increment the counter so the next sequential node will have the next number.
	parent.NextSequentialNode++
	// Set the response and return.
	resp.ZNodeName = newName
	return nil
}

func (s *Server) Delete(req *DeleteReq, _ *DeleteResp) error {
	err := validatePath(req.Path)
	if err != nil {
		return err
	}
	names := splitPathIntoNodeNames(req.Path)

	// Search down the tree until we hit the parent where we'll be creating this new node.
	parent := findZNode(s.root, names[:len(names)-1])
	if parent == nil {
		return fmt.Errorf("at least one of the anscestors of this node are missing")
	}

	nameToDelete := names[len(names)-1]
	node, ok := parent.Children[nameToDelete]
	if !ok {
		// If the node doesn't exist, then act like the operation succeeded.
		return nil
	}
	if !isValidVersion(req.Version, node.Version) {
		return fmt.Errorf("invalid version: expected [%d], actual [%d]", req.Version, node.Version)
	}
	if len(node.Children) > 0 {
		return fmt.Errorf("the node specified has children. Only leaf nodes can be deleted")
	}
	delete(parent.Children, nameToDelete)

	return nil
}

func (s *Server) Exists(req *ExistsReq, resp *ExistsResp) error {
	err := validatePath(req.Path)
	if err != nil {
		return err
	}
	names := splitPathIntoNodeNames(req.Path)

	node := findZNode(s.root, names)
	// TODO: Implement watching mechanism.
	// Set the response and return.
	resp.Exists = node != nil
	return nil
}

func (s *Server) GetData(req *GetDataReq, resp *GetDataResp) error {
	err := validatePath(req.Path)
	if err != nil {
		return err
	}
	names := splitPathIntoNodeNames(req.Path)

	node := findZNode(s.root, names)
	if node == nil {
		// TODO: Should we return an error if the node doesn't exist?
		return nil
	}
	// TODO: Implement watching mechanism.
	// Set the fields in the response and return.
	resp.Data = node.Data
	resp.Version = node.Version
	return nil
}

func (s *Server) SetData(req *SetDataReq, _ *SetDataResp) error {
	err := validatePath(req.Path)
	if err != nil {
		return err
	}
	names := splitPathIntoNodeNames(req.Path)

	node := findZNode(s.root, names)
	if node == nil {
		return fmt.Errorf("node does not exist")
	}
	if !isValidVersion(req.Version, node.Version) {
		return fmt.Errorf("invalid version: expected [%d], actual [%d]", req.Version, node.Version)
	}
	node.Data = req.Data
	node.Version++
	return nil
}

func (s *Server) GetChildren(path string, watch bool) ([]string, error) {
	err := validatePath(path)
	if err != nil {
		return nil, err
	}
	names := splitPathIntoNodeNames(path)

	node := findZNode(s.root, names)
	if node == nil {
		return nil, nil
	}

	// Just get the names of the children from the map.
	var childrenNames []string
	for name := range node.Children {
		childrenNames = append(childrenNames, name)
	}
	// TODO: Implement watching mechanism.
	return childrenNames, nil
}

func (s *Server) Sync(_ string) {

}

func splitPathIntoNodeNames(path string) []string {
	// Since we have a leading /, then we expect the first name to be empty.
	return strings.Split(path, "/")[1:]
}

// findZNode will search down to the tree and return the node specified by the names.
// If the node could not be found, then we will return nil.
func findZNode(start *znode.ZNode, names []string) *znode.ZNode {
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
