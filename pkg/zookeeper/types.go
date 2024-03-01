package zookeeper

// ClientID is the ID used to maintain a client/server session. This is expected
// to be included in every request to the server. The caller isn't expected to set
// this value. The Zookeeper client library will do this automatically and will overwrite
// any value directly set by the client anyway.
type ClientID struct {
	ID string
}

// TODO: Add comments for this whole file.
type Flag int

const (
	// EPHEMERAL indicates that the ZNode to be created should be automatically destroyed once the session
	// has been terminated (either intentionally or on failure).
	EPHEMERAL Flag = iota
	// SEQUENTIAL indicates that the node to be created should have a monotonically increasing counter appended
	// to the end of the provided name.
	SEQUENTIAL
)

type CreateReq struct {
	ClientID

	Path  string
	Data  []byte
	Flags []Flag
}

type CreateResp struct {
	ZNodeName string
}

type DeleteReq struct {
	ClientID

	Path    string
	Version int
}

type DeleteResp struct{}

type ExistsReq struct {
	ClientID

	Path string
	// TODO: Should this be implemented as a function callback instead? How are we going to send another response
	// to the same request? Maybe switch to using gRPC with streaming.
	Watch bool
}

type ExistsResp struct {
	Exists bool
}

type GetDataReq struct {
	ClientID

	Path  string
	Watch bool
}

type GetDataResp struct {
	Data    []byte
	Version int
}

type SetDataReq struct {
	ClientID

	Path    string
	Data    []byte
	Version int
}

type SetDataResp struct{}

type GetChildrenReq struct {
	ClientID

	Path  string
	Watch bool
}

type GetChildrenResp struct {
	Children []string
}

type SyncReq struct {
	ClientID

	Path string
}

type SyncResp struct{}

/*
Server/Client connections
*/
type ConnectReq struct {
	ClientID
}

type ConnectResp struct{}

type CloseReq struct {
	ClientID
}

type CloseResp struct{}
