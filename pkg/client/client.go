package client

import "github.com/mikekulinski/zookeeper/pkg/zookeeper"

type Zookeeper interface {
	// Create creates a ZNode with path name path, stores data in it, and returns the name of the new ZNode
	// Flags can also be passed to pick certain attributes you want the ZNode to have.
	Create(req *zookeeper.CreateReq, resp *zookeeper.CreateResp) error
	// Delete deletes the ZNode at the given path if that ZNode is at the expected version.
	Delete(req *zookeeper.DeleteReq, resp *zookeeper.DeleteResp) error
	// Exists returns true if the ZNode with path name path exists, and returns false otherwise. The watch flag
	// enables a client to set a watch on the ZNode.
	Exists(req *zookeeper.ExistsReq, resp *zookeeper.ExistsResp) error
	// GetData returns the data and metadata, such as version information, associated with the ZNode.
	// The watch flag works in the same way as it does for exists(), except that ZooKeeper does not set the watch
	// if the ZNode does not exist.
	GetData(req *zookeeper.GetDataReq, resp *zookeeper.GetDataResp) error
	// SetData writes data to the ZNode path if the version number is the current version of the ZNode.
	SetData(req *zookeeper.SetDataReq, resp *zookeeper.SetDataResp) error
	// GetChildren returns the set of names of the children of a ZNode.
	GetChildren(req *zookeeper.GetChildrenReq, resp *zookeeper.GetChildrenResp) error
	// Sync waits for all updates pending at the start of the operation to propagate to the server
	// that the client is connected to. The path is currently ignored. (Using path is not discussed in the white paper)
	Sync(req *zookeeper.SyncReq, resp *zookeeper.SyncResp) error
}
