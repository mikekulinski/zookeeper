package client

import (
	"fmt"
	"net/rpc"

	"github.com/google/uuid"
	"github.com/mikekulinski/zookeeper/pkg/zookeeper"
)

type Client struct {
	rpcClient *rpc.Client
	clientID  string
}

func NewClient(endpoint string) (*Client, error) {
	// Dial the initial RPC connection.
	rpcClient, err := rpc.DialHTTP("tcp", endpoint+":8080")
	if err != nil {
		return nil, fmt.Errorf("dialing: %w", err)
	}

	clientID := uuid.New().String()
	// Initiate the connection with the Zookeeper server.
	req := &zookeeper.ConnectReq{
		ClientID: zookeeper.ClientID{ID: clientID},
	}
	resp := &zookeeper.ConnectResp{}
	err = rpcClient.Call("Zookeeper.Connect", req, resp)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Zookeeper: %w", err)
	}
	return &Client{
		rpcClient: rpcClient,
		clientID:  clientID,
	}, nil
}

func (c *Client) Close() error {
	req := &zookeeper.CloseReq{
		ClientID: zookeeper.ClientID{ID: c.clientID},
	}
	resp := &zookeeper.CloseResp{}
	err := c.rpcClient.Call("Zookeeper.Close", req, resp)
	if err != nil {
		return fmt.Errorf("error closing the Zookeeper connection: %w", err)
	}
	return nil
}

// Create creates a ZNode with path name path, stores data in it, and returns the name of the new ZNode
// Flags can also be passed to pick certain attributes you want the ZNode to have.
func (c *Client) Create(req *zookeeper.CreateReq) (*zookeeper.CreateResp, error) {
	req.ClientID.ID = c.clientID
	resp := &zookeeper.CreateResp{}
	err := c.rpcClient.Call("Zookeeper.Create", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Delete deletes the ZNode at the given path if that ZNode is at the expected version.
func (c *Client) Delete(req *zookeeper.DeleteReq) (*zookeeper.DeleteResp, error) {
	req.ClientID.ID = c.clientID
	resp := &zookeeper.DeleteResp{}
	err := c.rpcClient.Call("Zookeeper.Delete", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Exists returns true if the ZNode with path name path exists, and returns false otherwise. The watch flag
// enables a client to set a watch on the ZNode.
func (c *Client) Exists(req *zookeeper.ExistsReq) (*zookeeper.ExistsResp, error) {
	req.ClientID.ID = c.clientID
	resp := &zookeeper.ExistsResp{}
	err := c.rpcClient.Call("Zookeeper.Exists", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetData returns the data and metadata, such as version information, associated with the ZNode.
// The watch flag works in the same way as it does for exists(), except that ZooKeeper does not set the watch
// if the ZNode does not exist.
func (c *Client) GetData(req *zookeeper.GetDataReq) (*zookeeper.GetDataResp, error) {
	req.ClientID.ID = c.clientID
	resp := &zookeeper.GetDataResp{}
	err := c.rpcClient.Call("Zookeeper.GetData", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SetData writes data to the ZNode path if the version number is the current version of the ZNode.
func (c *Client) SetData(req *zookeeper.SetDataReq) (*zookeeper.SetDataResp, error) {
	req.ClientID.ID = c.clientID
	resp := &zookeeper.SetDataResp{}
	err := c.rpcClient.Call("Zookeeper.SetData", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetChildren returns the set of names of the children of a ZNode.
func (c *Client) GetChildren(req *zookeeper.GetChildrenReq) (*zookeeper.GetChildrenResp, error) {
	req.ClientID.ID = c.clientID
	resp := &zookeeper.GetChildrenResp{}
	err := c.rpcClient.Call("Zookeeper.GetChildren", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Sync waits for all updates pending at the start of the operation to propagate to the server
// that the client is connected to. The path is currently ignored. (Using path is not discussed in the white paper)
func (c *Client) Sync(req *zookeeper.SyncReq) (*zookeeper.SyncResp, error) {
	req.ClientID.ID = c.clientID
	resp := &zookeeper.SyncResp{}
	err := c.rpcClient.Call("Zookeeper.Sync", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
