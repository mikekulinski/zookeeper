package pkg

import (
	"fmt"
	"slices"
	"strings"
)

type Flag int

const (
	// EPHEMERAL indicates that the ZNode to be created should be automatically destroyed once the session
	// has been terminated (either intentionally or on failure).
	EPHEMERAL Flag = iota
	// SEQUENTIAL indicates that the node to be created should have a monotonically increasing counter appended
	// to the end of the provided name.
	// TODO: Should we enforce ZNodes to only have sequential ZNode children if they have any?
	SEQUENTIAL
	// UNVERSIONED indicates that a created ZNode doesn't care about versioning. This will eventually
	// be saved on the ZNode with version == -1.
	UNVERSIONED
)

type Zookeeper interface {
	// Create creates a ZNode with path name path, stores data[] in it, and returns the name of the new ZNode
	// Flags can also be passed to pick certain attributes you want the ZNode to have.
	Create(path string, data []byte, flags ...Flag) (ZNodeName string, _ error)
	// Delete deletes the ZNode at the given path if that ZNode is at the expected version.
	Delete(path string, version int)
	// Exists returns true if the ZNode with path name path exists, and returns false otherwise. The watch flag
	// enables a client to set a watch on the ZNode.
	Exists(path string, watch bool) bool
	// GetData returns the data and metadata, such as version information, associated with the ZNode.
	// The watch flag works in the same way as it does for exists(), except that ZooKeeper does not set the watch
	// if the ZNode does not exist.
	GetData(path string, watch bool) (data []byte, version int)
	// SetData writes data to the ZNode path if the version number is the current version of the ZNode.
	// TODO: What do we do if the version is invalid? Should we return some sort of error message?
	SetData(path string, data []byte, version int)
	// GetChildren returns the set of names of the children of a ZNode.
	GetChildren(path string, watch bool) []string
	// Sync waits for all updates pending at the start of the operation to propagate to the server
	// that the client is connected to. The path is currently ignored. (Using path is not discussed in the white paper)
	Sync(path string)
}

type Server struct {
	root *ZNode
}

func NewServer() *Server {
	return &Server{
		root: NewZNode("", -1, ZNodeType_STANDARD, nil),
	}
}

func (s *Server) Create(path string, data []byte, flags ...Flag) (string, error) {
	err := validatePath(path)
	if err != nil {
		return "", err
	}
	names := splitPathIntoNodeNames(path)

	// Search down the tree until we hit the parent where we'll be creating this
	// new node.
	znode := s.root
	for _, name := range names[:len(names)-1] {
		z, ok := znode.children[name]
		if !ok {
			return "", fmt.Errorf("parent node [%s] could not be found", name)
		}
		znode = z
	}

	// We are at the parent node of the one we are trying to create. Now let's
	// try to create it.
	// TODO: Implement SEQUENTIAL. We'll need to keep track of the last number we used.
	newName := names[len(names)-1]
	version := 0
	if slices.Contains(flags, UNVERSIONED) {
		version = -1
	}
	nodeType := ZNodeType_STANDARD
	if slices.Contains(flags, EPHEMERAL) {
		nodeType = ZNodeType_EPHEMERAL
	}
	newNode := NewZNode(
		newName,
		version,
		nodeType,
		data,
	)

	if _, ok := znode.children[newName]; ok {
		return "", fmt.Errorf("node [%s] already exists", newName)
	}
	znode.children[newName] = newNode
	return newName, nil
}

func (s *Server) Delete(path string, version int) {

}

func (s *Server) Exists(path string, watch bool) bool {
	return false
}

func (s *Server) GetData(path string, watch bool) ([]byte, int) {
	return nil, 0
}

func (s *Server) SetData(path string, data []byte, version int) {

}

func (s *Server) GetChildren(path string, watch bool) []string {
	return nil
}

func (s *Server) Sync(_ string) {

}

func validatePath(path string) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path does not start at the root")
	}

	if path == "/" {
		return fmt.Errorf("path cannot be the root")
	}

	if strings.HasSuffix(path, "/") {
		return fmt.Errorf("path should end in a node name, a '/'")
	}

	names := strings.Split(path, "/")
	// Since we have a leading /, then we expect the first name to be empty.
	for _, name := range names[1:] {
		if name == "" {
			return fmt.Errorf("path contains an empty node name")
		}
	}
	return nil
}

func splitPathIntoNodeNames(path string) []string {
	// Since we have a leading /, then we expect the first name to be empty.
	return strings.Split(path, "/")[1:]
}
