package server

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mikekulinski/zookeeper/pkg/session"
	"github.com/mikekulinski/zookeeper/pkg/znode"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: All these structures are not thread safe. We should add a mutex to control this on the server.
// TODO: Restructure the server to process transactions instead of client requests. That way we can
// simply call these functions on recovery with the transactions from the log.
type Server struct {
	pbzk.UnimplementedZookeeperServer

	db *znode.DB

	// sessions is a map of ClientID to session for all the clients
	// that are currently connected to Zookeeper.
	sessions map[string]*session.Session
	// watches is a mapping of ZNode path to information about the type of watches on that node.
	watches map[string][]*znode.Watch
}

func NewServer() *Server {
	return &Server{
		db:       znode.NewDB(),
		sessions: map[string]*session.Session{},
		watches:  map[string][]*znode.Watch{},
	}
}

// Create creates a ZNode with path name path, stores data in it, and returns the name of the new ZNode
// Flags can also be passed to pick certain attributes you want the ZNode to have.
func (s *Server) Create(ctx context.Context, req *pbzk.CreateRequest) (*pbzk.CreateResponse, error) {
	err := validatePath(req.GetPath())
	if err != nil {
		return nil, err
	}

	clientID, _ := ExtractClientIDHeader(ctx)
	txn := &pbzk.Transaction{
		ClientId:    clientID,
		Zxid:        0, // TODO
		TimestampMs: time.Now().UnixMilli(),
		Txn: &pbzk.Transaction_Create{
			Create: &pbzk.CreateTxn{
				Path:       req.GetPath(),
				Data:       req.GetData(),
				Ephemeral:  slices.Contains(req.GetFlags(), pbzk.CreateRequest_FLAG_EPHEMERAL),
				Sequential: slices.Contains(req.GetFlags(), pbzk.CreateRequest_FLAG_SEQUENTIAL),
			},
		},
	}
	newNode, err := s.db.Create(txn)
	if err != nil {
		return nil, err
	}

	// If this node is ephemeral, then tie it to this session.
	if newNode.NodeType == znode.ZNodeType_EPHEMERAL {
		sess, ok := s.sessions[clientID]
		if !ok {
			return nil, fmt.Errorf("session unexpectedly missing")
		}
		sess.EphemeralNodes[newNode.Name] = newNode
	}

	s.triggerWatches(newNode.Name, pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED)
	resp := &pbzk.CreateResponse{
		ZNodeName: newNode.Name,
	}
	return resp, nil
}

func newFullName(nodeName string, ancestorsNames []string) string {
	nodePath := "/" + nodeName
	if len(ancestorsNames) > 0 {
		return "/" + strings.Join(ancestorsNames, "/") + nodePath
	}
	return nodePath
}

// Delete deletes the ZNode at the given path if that ZNode is at the expected version.
func (s *Server) Delete(ctx context.Context, req *pbzk.DeleteRequest) (*pbzk.DeleteResponse, error) {
	err := validatePath(req.GetPath())
	if err != nil {
		return nil, err
	}

	// First, get the node and validate the request.
	node := s.db.Get(req.GetPath())
	if node == nil {
		return &pbzk.DeleteResponse{}, nil
	}

	// Make sure the node has the right version when deleting.
	if !isValidVersion(req.GetVersion(), node.Version) {
		return nil, fmt.Errorf("invalid version: expected [%d], actual [%d]", req.GetVersion(), node.Version)
	}

	// Nodes with children are not able to be deleted.
	if len(node.Children) > 0 {
		return nil, fmt.Errorf("the node specified has children. Only leaf nodes can be deleted")
	}

	// Actually delete from the DB.
	clientID, _ := ExtractClientIDHeader(ctx)
	txn := &pbzk.Transaction{
		ClientId:    clientID,
		Zxid:        0, // TODO
		TimestampMs: time.Now().UnixMilli(),
		Txn: &pbzk.Transaction_Delete{
			Delete: &pbzk.DeleteTxn{
				Path: req.GetPath(),
			},
		},
	}
	err = s.db.Delete(txn)
	if err != nil {
		return nil, err
	}

	// Clean up any references if this was ephemeral.
	if node.NodeType == znode.ZNodeType_EPHEMERAL {
		// If this session hasn't already been deleted, then clean up the reference to this node here.
		if sess, ok := s.sessions[node.Creator]; ok {
			delete(sess.EphemeralNodes, req.GetPath())
		}
	}
	s.triggerWatches(req.GetPath(), pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED)
	return &pbzk.DeleteResponse{}, nil
}

// Exists returns true if the ZNode with path name path exists, and returns false otherwise. The watch flag
// enables a client to set a watch on the ZNode.
func (s *Server) Exists(ctx context.Context, req *pbzk.ExistsRequest) (*pbzk.ExistsResponse, error) {
	err := validatePath(req.GetPath())
	if err != nil {
		return nil, err
	}

	node := s.db.Get(req.GetPath())

	// If the client wants to watch for changes on this node, then add it to our map of watches.
	if req.GetWatch() {
		clientID, _ := ExtractClientIDHeader(ctx)
		w := &znode.Watch{
			ClientID: clientID,
			Path:     req.GetPath(),
			// Exists calls watch for creates, updates, or deletes to the node specified.
			WatchTypes: []pbzk.WatchEvent_EventType{
				pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED,
				pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
				pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
			},
		}
		s.watches[req.GetPath()] = append(s.watches[req.GetPath()], w)
	}
	return &pbzk.ExistsResponse{
		Exists: node != nil,
	}, nil
}

// GetData returns the data and metadata, such as version information, associated with the ZNode.
// The watch flag works in the same way as it does for exists(), except that ZooKeeper does not set the watch
// if the ZNode does not exist.
func (s *Server) GetData(ctx context.Context, req *pbzk.GetDataRequest) (*pbzk.GetDataResponse, error) {
	err := validatePath(req.GetPath())
	if err != nil {
		return nil, err
	}

	node := s.db.Get(req.GetPath())
	if node == nil {
		return &pbzk.GetDataResponse{}, nil
	}

	// If the client wants to watch for changes on this node, then add it to our map of watches.
	if req.GetWatch() {
		clientID, _ := ExtractClientIDHeader(ctx)
		w := &znode.Watch{
			ClientID: clientID,
			Path:     req.GetPath(),
			// GetData calls watch for updates or deletes to the node specified.
			WatchTypes: []pbzk.WatchEvent_EventType{
				pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
				pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
			},
		}
		s.watches[req.GetPath()] = append(s.watches[req.GetPath()], w)
	}
	return &pbzk.GetDataResponse{
		Data:    node.Data,
		Version: node.Version,
	}, nil
}

// SetData writes data to the ZNode path if the version number is the current version of the ZNode.
func (s *Server) SetData(ctx context.Context, req *pbzk.SetDataRequest) (*pbzk.SetDataResponse, error) {
	err := validatePath(req.GetPath())
	if err != nil {
		return nil, err
	}

	node := s.db.Get(req.GetPath())
	if node == nil {
		return nil, fmt.Errorf("node does not exist")
	}
	if !isValidVersion(req.GetVersion(), node.Version) {
		return nil, fmt.Errorf("invalid version: expected [%d], actual [%d]", req.GetVersion(), node.Version)
	}

	clientID, _ := ExtractClientIDHeader(ctx)
	txn := &pbzk.Transaction{
		ClientId:    clientID,
		Zxid:        0, // TODO
		TimestampMs: time.Now().UnixMilli(),
		Txn: &pbzk.Transaction_SetData{
			SetData: &pbzk.SetDataTxn{
				Path: req.GetPath(),
				Data: req.GetData(),
			},
		},
	}
	err = s.db.SetData(txn)
	if err != nil {
		return nil, err
	}

	s.triggerWatches(req.GetPath(), pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED)
	return &pbzk.SetDataResponse{}, nil
}

// GetChildren returns the set of names of the children of a ZNode.
func (s *Server) GetChildren(ctx context.Context, req *pbzk.GetChildrenRequest) (*pbzk.GetChildrenResponse, error) {
	err := validatePath(req.GetPath())
	if err != nil {
		return nil, err
	}

	node := s.db.Get(req.GetPath())
	if node == nil {
		return &pbzk.GetChildrenResponse{}, nil
	}

	// Just get the names of the children from the map.
	var childrenNames []string
	for name := range node.Children {
		childrenNames = append(childrenNames, name)
	}

	// If the client wants to watch for changes on this node, then add it to our map of watches.
	if req.GetWatch() {
		clientID, _ := ExtractClientIDHeader(ctx)
		w := &znode.Watch{
			ClientID: clientID,
			Path:     req.GetPath(),
			// GetChildren calls only watch for children update events.
			WatchTypes: []pbzk.WatchEvent_EventType{
				pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED,
				// We also watch for delete events to this node since deletes mean we won't have any more children.
				pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
			},
		}
		s.watches[req.GetPath()] = append(s.watches[req.GetPath()], w)
	}
	return &pbzk.GetChildrenResponse{
		Children: childrenNames,
	}, nil
}

// Sync waits for all updates pending at the start of the operation to propagate to the server
// that the client is connected to. The path is currently ignored. (Using path is not discussed in the white paper)
func (s *Server) Sync(_ context.Context, _ *pbzk.SyncRequest) (*pbzk.SyncResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sync not implemented")
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

// triggerWatches will notify all clients that are watching for events for that node.
func (s *Server) triggerWatches(path string, watchType pbzk.WatchEvent_EventType) {
	watchesToTrigger := s.extractWatches(path, watchType)

	// For create/delete events, check if this triggered any child watches in the parent.
	var childWatchesToTrigger []*znode.Watch
	if watchType == pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED ||
		watchType == pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED {
		parentPath := getParent(path)
		childWatchesToTrigger = s.extractWatches(parentPath, pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED)
	}

	// Actually trigger the watches.
	s.triggerEachWatch(watchesToTrigger, watchType)
	s.triggerEachWatch(childWatchesToTrigger, pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED)
}

func (s *Server) extractWatches(path string, watchType pbzk.WatchEvent_EventType) []*znode.Watch {
	var watchesToTrigger []*znode.Watch
	var clientIDsToRemove []string
	for _, watch := range s.watches[path] {
		if slices.Contains(watch.WatchTypes, watchType) {
			watchesToTrigger = append(watchesToTrigger, watch)
			clientIDsToRemove = append(clientIDsToRemove, watch.ClientID)
		}
	}
	for _, id := range clientIDsToRemove {
		s.watches[path] = slices.DeleteFunc(s.watches[path], func(watch *znode.Watch) bool {
			return watch.ClientID == id
		})
	}
	return watchesToTrigger
}

func (s *Server) triggerEachWatch(watches []*znode.Watch, watchType pbzk.WatchEvent_EventType) {
	for _, w := range watches {
		// No need to capture loop var since we're using Go 1.22.
		// Trigger each watch in a separate goroutine since adding to the messages channel is blocking.
		go func() {
			if sess, ok := s.sessions[w.ClientID]; ok {
				event := &session.Event{
					WatchEvent: &pbzk.WatchEvent{
						Type: watchType,
					},
				}
				sess.Messages <- event
			}
		}()
	}
}

func getParent(path string) string {
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return ""
	}
	parentPath := path[:i]
	return parentPath
}
