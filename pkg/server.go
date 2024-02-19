package pkg

type Flag int

const (
	REGULAR int = iota
	EPHEMERAL
	SEQUENTIAL
)

type Zookeeper interface {
	// Create creates a znode with path name path, stores data[] in it, and returns the name of the new znode.
	// flags enables a client to select the type of znode: regular, ephemeral, and set the sequential flag.
	Create(path string, data []byte, flags ...Flag) (znodeName string)
	// Delete deletes the znode at the given path if that znode is at the expected version.
	Delete(path string, version int)
	// Exists returns true if the znode with path name path exists, and returns false otherwise. The watch flag
	// enables a client to set a watch on the znode.
	Exists(path string, watch bool) bool
	// GetData returns the data and metadata, such as version information, associated with the znode.
	// The watch flag works in the same way as it does for exists(), except that ZooKeeper does not set the watch
	// if the znode does not exist.
	GetData(path string, watch bool) (data []byte, version int)
	// SetData writes data to the znode path if the version number is the current version of the znode.
	// TODO: What do we do if the version is invalid? Should we return some sort of error message?
	SetData(path string, data []byte, version int)
	// GetChildren returns the set of names of the children of a znode.
	GetChildren(path string, watch bool) []string
	// Sync waits for all updates pending at the start of the operation to propagate to the server
	// that the client is connected to. The path is currently ignored. (Using path is not discussed in the white paper)
	Sync(path string)
}

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Create(path string, data []byte, flags ...Flag) string {
	return ""
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
