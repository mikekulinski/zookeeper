package server

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
	Path  string
	Data  []byte
	Flags []Flag
}

type CreateResp struct {
	ZNodeName string
}

type DeleteReq struct {
	Path    string
	Version int
}

type DeleteResp struct{}

type ExistsReq struct {
	Path  string
	Watch bool
}

type ExistsResp struct {
	Exists bool
}

type GetDataReq struct {
	Path  string
	Watch bool
}

type GetDataResp struct {
	Data    []byte
	Version int
}

type SetDataReq struct {
	Path    string
	Data    []byte
	Version int
}

type SetDataResp struct{}

type GetChildrenReq struct {
	Path  string
	Watch bool
}

type GetChildrenResp struct {
	ChildrenNames []string
}

type SyncReq struct {
	Path string
}

type SyncResp struct{}
