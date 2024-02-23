package server

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mikekulinski/zookeeper/pkg/session"
	"github.com/mikekulinski/zookeeper/pkg/znode"
	"github.com/mikekulinski/zookeeper/pkg/zookeeper"
)

type Server struct {
	root *znode.ZNode
	// Sessions is a map of ClientID to session for all the clients
	// that are currently connected to Zookeeper.
	sessions map[string]*session.Session
}

func NewServer() *Server {
	return &Server{
		root: znode.NewZNode("", -1, znode.ZNodeType_STANDARD, nil),
	}
}

// Create creates a ZNode with path name path, stores data in it, and returns the name of the new ZNode
// Flags can also be passed to pick certain attributes you want the ZNode to have.
func (s *Server) Create(req *zookeeper.CreateReq, resp *zookeeper.CreateResp) error {
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
	if slices.Contains(req.Flags, zookeeper.SEQUENTIAL) {
		newName = fmt.Sprintf("%s_%d", newName, parent.NextSequentialNode)
	}
	nodeType := znode.ZNodeType_STANDARD
	if slices.Contains(req.Flags, zookeeper.EPHEMERAL) {
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

// Delete deletes the ZNode at the given path if that ZNode is at the expected version.
func (s *Server) Delete(req *zookeeper.DeleteReq, _ *zookeeper.DeleteResp) error {
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

// Exists returns true if the ZNode with path name path exists, and returns false otherwise. The watch flag
// enables a client to set a watch on the ZNode.
func (s *Server) Exists(req *zookeeper.ExistsReq, resp *zookeeper.ExistsResp) error {
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

// GetData returns the data and metadata, such as version information, associated with the ZNode.
// The watch flag works in the same way as it does for exists(), except that ZooKeeper does not set the watch
// if the ZNode does not exist.
func (s *Server) GetData(req *zookeeper.GetDataReq, resp *zookeeper.GetDataResp) error {
	err := validatePath(req.Path)
	if err != nil {
		return err
	}
	names := splitPathIntoNodeNames(req.Path)

	node := findZNode(s.root, names)
	if node == nil {
		return nil
	}
	// TODO: Implement watching mechanism.
	// Set the fields in the response and return.
	resp.Data = node.Data
	resp.Version = node.Version
	return nil
}

// SetData writes data to the ZNode path if the version number is the current version of the ZNode.
func (s *Server) SetData(req *zookeeper.SetDataReq, _ *zookeeper.SetDataResp) error {
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

// GetChildren returns the set of names of the children of a ZNode.
func (s *Server) GetChildren(req *zookeeper.GetChildrenReq, resp *zookeeper.GetChildrenResp) error {
	err := validatePath(req.Path)
	if err != nil {
		return err
	}
	names := splitPathIntoNodeNames(req.Path)

	node := findZNode(s.root, names)
	if node == nil {
		return nil
	}

	// Just get the names of the children from the map.
	var childrenNames []string
	for name := range node.Children {
		childrenNames = append(childrenNames, name)
	}
	// TODO: Implement watching mechanism.
	// Set the fields in the response and return.
	resp.Children = childrenNames
	return nil
}

// Sync waits for all updates pending at the start of the operation to propagate to the server
// that the client is connected to. The path is currently ignored. (Using path is not discussed in the white paper)
func (s *Server) Sync(_ *zookeeper.SyncReq, _ *zookeeper.SyncResp) error {
	return fmt.Errorf("not implemented")
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
