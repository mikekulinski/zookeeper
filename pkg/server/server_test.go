package server

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mikekulinski/zookeeper/pkg/znode"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Move some of these unit tests to the DB.
func TestServer_Create(t *testing.T) {
	const existingNodeName = "existing"

	tests := []struct {
		name           string
		path           string
		parentNodeType znode.ZNodeType
		errorExpected  bool
	}{
		{
			name:          "invalid path",
			path:          "invalid",
			errorExpected: true,
		},
		{
			name:          "parent node missing",
			path:          "/x/y/z",
			errorExpected: true,
		},
		{
			name:          "node already exists",
			path:          fmt.Sprintf("/%s", existingNodeName),
			errorExpected: true,
		},
		{
			name:          "valid create, root",
			path:          "/xyz",
			errorExpected: false,
		},
		{
			name:          "valid create, child of existing node",
			path:          fmt.Sprintf("/%s/new", existingNodeName),
			errorExpected: false,
		},
		{
			name:           "parent node is ephemeral",
			path:           fmt.Sprintf("/%s/new", existingNodeName),
			parentNodeType: znode.ZNodeType_EPHEMERAL,
			errorExpected:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			zk := NewServer()
			// TODO: Update this so that we use the full name in the ZNode.
			// Pre-init the server with some nodes so we can also test cases with existing nodes.
			zk.root.Children = map[string]*znode.ZNode{
				existingNodeName: znode.NewZNode(existingNodeName, test.parentNodeType, "", nil),
			}

			req := &pbzk.CreateRequest{
				Path: test.path,
			}
			_, err := zk.Create(ctx, req)
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_Create_Sequential(t *testing.T) {
	const existingNodeName = "existing_5"
	tests := []struct {
		name           string
		path           string
		parentNodeType znode.ZNodeType
		errorExpected  bool
	}{
		{
			name:          "node already exists",
			path:          "/existing",
			errorExpected: true,
		},
		{
			name:          "valid create, root",
			path:          "/new",
			errorExpected: false,
		},
		{
			name:          "valid create, child of existing node",
			path:          fmt.Sprintf("/%s/new", existingNodeName),
			errorExpected: false,
		},
		{
			name:           "parent node is ephemeral",
			path:           fmt.Sprintf("/%s/new", existingNodeName),
			parentNodeType: znode.ZNodeType_EPHEMERAL,
			errorExpected:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			zk := NewServer()
			// Pre-init the server with some nodes so we can also test cases with existing nodes.
			zk.root.NextSequentialNode = 5
			existingNode := znode.NewZNode(existingNodeName, test.parentNodeType, "", nil)
			existingNode.NextSequentialNode = 5
			zk.root.Children = map[string]*znode.ZNode{
				existingNodeName: existingNode,
			}

			req := &pbzk.CreateRequest{
				Path:  test.path,
				Flags: []pbzk.CreateRequest_Flag{pbzk.CreateRequest_FLAG_SEQUENTIAL},
			}
			resp, err := zk.Create(ctx, req)
			if test.errorExpected {
				assert.Empty(t, resp.GetZNodeName())
				assert.Error(t, err)
			} else {
				fmt.Println(resp.GetZNodeName())
				assert.Contains(t, resp.GetZNodeName(), "new_5")
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_Create_RegularThenSequential(t *testing.T) {
	const existingNodeName = "existing"
	ctx := context.Background()
	zk := NewServer()
	// Pre-init the server with some nodes so we can also test cases with existing nodes.
	zk.root.NextSequentialNode = 5
	existingNode := znode.NewZNode(existingNodeName, znode.ZNodeType_STANDARD, "", nil)
	existingNode.NextSequentialNode = 5
	zk.root.Children = map[string]*znode.ZNode{
		existingNodeName: existingNode,
	}

	// Create a regular node. It shouldn't have the sequential suffix.
	req := &pbzk.CreateRequest{
		Path: "/" + existingNodeName + "/standard",
	}
	resp, err := zk.Create(ctx, req)
	fmt.Println(resp.GetZNodeName())
	assert.True(t, strings.HasSuffix(resp.GetZNodeName(), "standard"))
	assert.NoError(t, err)
	// NextSequentialNode shouldn't be incremented.
	assert.Equal(t, 5, existingNode.NextSequentialNode)

	// Create a sequential node. It should use 5 as the suffix. NextSequentialNode should only be incremented
	// after creating another sequential node.
	req = &pbzk.CreateRequest{
		Path:  "/" + existingNodeName + "/sequential",
		Flags: []pbzk.CreateRequest_Flag{pbzk.CreateRequest_FLAG_SEQUENTIAL},
	}
	resp, err = zk.Create(ctx, req)
	fmt.Println(resp.GetZNodeName())
	assert.True(t, strings.HasSuffix(resp.GetZNodeName(), "sequential_5"))
	assert.NoError(t, err)
	// NextSequentialNode should be incremented.
	assert.Equal(t, 6, existingNode.NextSequentialNode)
}

func TestServer_Delete(t *testing.T) {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name            string
		path            string
		version         int64
		errorExpected   bool
		deleteSucceeded bool
	}{
		{
			name:          "invalid path",
			path:          "invalid",
			errorExpected: true,
		},
		{
			name:          "parent node missing",
			path:          "/x/y/z",
			errorExpected: true,
		},
		{
			name:          "node missing, no error",
			path:          "/random",
			errorExpected: false,
		},
		{
			name:          "invalid version",
			path:          "/" + rootChildName,
			version:       24,
			errorExpected: true,
		},
		{
			name:          "invalid delete, node has children",
			path:          "/" + rootChildName,
			version:       0,
			errorExpected: true,
		},
		{
			name:            "valid delete, child of existing node",
			path:            fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			version:         0,
			errorExpected:   false,
			deleteSucceeded: true,
		},
		{
			name:            "valid delete, ignore version check",
			path:            fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			version:         -1,
			errorExpected:   false,
			deleteSucceeded: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			zk := NewServer()
			cReq := &pbzk.CreateRequest{
				Path: "/" + rootChildName,
			}
			_, err := zk.Create(ctx, cReq)
			require.NoError(t, err)
			cReq = &pbzk.CreateRequest{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			_, err = zk.Create(ctx, cReq)
			require.NoError(t, err)

			dReq := &pbzk.DeleteRequest{
				Path:    test.path,
				Version: test.version,
			}
			_, err = zk.Delete(ctx, dReq)
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			fmt.Println(zk.root)
			// If we successfully deleted the leaf node, then verify we don't see it in the trie.
			if test.deleteSucceeded {
				assert.Empty(t, zk.root.Children[rootChildName].Children)
			}
		})
	}
}

func TestServer_Exists_NodeCreated(t *testing.T) {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name          string
		path          string
		exists        bool
		errorExpected bool
	}{
		{
			name:          "invalid path",
			path:          "invalid",
			errorExpected: true,
		},
		{
			name:   "node missing",
			path:   "/random",
			exists: false,
		},
		{
			name:   "parent node missing",
			path:   "/x/y/z",
			exists: false,
		},
		{
			name:   "parent exists, child missing",
			path:   fmt.Sprintf("/%s/random", rootChildName),
			exists: false,
		},
		{
			name:   "node exists, root",
			path:   "/" + rootChildName,
			exists: true,
		},
		{
			name:   "node exists, child of another node",
			path:   fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			exists: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			zk := NewServer()
			req := &pbzk.CreateRequest{
				Path: "/" + rootChildName,
			}
			_, err := zk.Create(ctx, req)
			require.NoError(t, err)
			req = &pbzk.CreateRequest{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			_, err = zk.Create(ctx, req)
			require.NoError(t, err)

			eReq := &pbzk.ExistsRequest{
				Path: test.path,
			}
			eResp, err := zk.Exists(ctx, eReq)
			if test.errorExpected {
				assert.False(t, eResp.GetExists())
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.exists, eResp.GetExists())
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_Exists_NodeCreatedThenDeleted(t *testing.T) {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name   string
		path   string
		exists bool
	}{
		{
			name:   "node was deleted, root",
			path:   "/" + rootChildName,
			exists: false,
		},
		{
			name:   "node was deleted, child of another node",
			path:   fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			exists: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			zk := NewServer()
			// Create nodes to check for.
			cReq := &pbzk.CreateRequest{
				Path: "/" + rootChildName,
			}
			_, err := zk.Create(ctx, cReq)
			require.NoError(t, err)
			cReq = &pbzk.CreateRequest{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			_, err = zk.Create(ctx, cReq)
			require.NoError(t, err)

			// Delete all those nodes to verify we deleted them.
			dReq := &pbzk.DeleteRequest{
				Path:    fmt.Sprintf("/%s/%s", rootChildName, childChildName),
				Version: -1,
			}
			_, err = zk.Delete(ctx, dReq)
			require.NoError(t, err)
			dReq = &pbzk.DeleteRequest{
				Path:    "/" + rootChildName,
				Version: -1,
			}
			_, err = zk.Delete(ctx, dReq)
			require.NoError(t, err)

			eReq := &pbzk.ExistsRequest{
				Path: test.path,
			}
			eResp, err := zk.Exists(ctx, eReq)
			assert.Equal(t, test.exists, eResp.GetExists())
			assert.NoError(t, err)
		})
	}
}

func TestServer_GetData(t *testing.T) {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name          string
		path          string
		data          []byte
		version       int64
		errorExpected bool
	}{
		{
			name:          "invalid path",
			path:          "invalid",
			errorExpected: true,
		},
		{
			name:    "node missing",
			path:    "/random",
			data:    nil,
			version: 0,
		},
		{
			name:    "parent node missing",
			path:    "/x/y/z",
			data:    nil,
			version: 0,
		},
		{
			name:    "parent exists, child missing",
			path:    fmt.Sprintf("/%s/random", rootChildName),
			data:    nil,
			version: 0,
		},
		{
			name: "node exists, root",
			path: "/" + rootChildName,
			data: []byte("secret stuff"),
		},
		{
			name: "node exists, child of another node",
			path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			data: []byte("secret stuff"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			zk := NewServer()
			req := &pbzk.CreateRequest{
				Path: "/" + rootChildName,
				Data: test.data,
			}
			_, err := zk.Create(ctx, req)
			require.NoError(t, err)
			req = &pbzk.CreateRequest{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
				Data: test.data,
			}
			_, err = zk.Create(ctx, req)
			require.NoError(t, err)
			require.NoError(t, err)

			// Set the versions on the nodes so we can differentiate between missing nodes and nodes that exists but
			// have no data.
			rootChild := zk.root.Children[rootChildName]
			rootChild.Version = test.version
			rootChild.Children[childChildName].Version = test.version

			gReq := &pbzk.GetDataRequest{
				Path: test.path,
			}
			gResp, err := zk.GetData(ctx, gReq)
			if test.errorExpected {
				assert.Empty(t, gResp.GetData())
				assert.Zero(t, gResp.GetVersion())
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.data, gResp.GetData())
				assert.Empty(t, test.version, gResp.GetVersion())
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_SetData(t *testing.T) {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name          string
		path          string
		version       int64
		errorExpected bool
		writeSucceeds bool
	}{
		{
			name:          "invalid path",
			path:          "invalid",
			errorExpected: true,
		},
		{
			name:          "node missing",
			path:          "/random",
			errorExpected: true,
		},
		{
			name:          "parent node missing",
			path:          "/x/y/z",
			errorExpected: true,
		},
		{
			name:          "parent exists, child missing",
			path:          fmt.Sprintf("/%s/random", rootChildName),
			errorExpected: true,
		},
		{
			name:          "node exists, root",
			path:          "/" + rootChildName,
			writeSucceeds: true,
		},
		{
			name:          "node exists, child of another node",
			path:          fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			writeSucceeds: true,
		},
		{
			name:          "invalid version",
			path:          "/" + rootChildName,
			version:       10,
			errorExpected: true,
		},
		{
			name:          "ignore version check",
			path:          "/" + rootChildName,
			version:       -1,
			writeSucceeds: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			dataToSet := []byte("you're a wizard Harry")

			zk := NewServer()
			req := &pbzk.CreateRequest{
				Path: "/" + rootChildName,
			}
			_, err := zk.Create(ctx, req)
			require.NoError(t, err)
			req = &pbzk.CreateRequest{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			_, err = zk.Create(ctx, req)
			require.NoError(t, err)

			sReq := &pbzk.SetDataRequest{
				Path:    test.path,
				Data:    dataToSet,
				Version: test.version,
			}
			_, err = zk.SetData(ctx, sReq)
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if test.writeSucceeds {
				gReq := &pbzk.GetDataRequest{
					Path: test.path,
				}
				gResp, err := zk.GetData(ctx, gReq)
				assert.Equal(t, dataToSet, gResp.GetData())
				// We expect the set to increment the version.
				assert.EqualValues(t, 1, gResp.GetVersion())
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_GetChildren(t *testing.T) {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name          string
		path          string
		children      []string
		errorExpected bool
	}{
		{
			name:          "invalid path",
			path:          "invalid",
			errorExpected: true,
		},
		{
			name:     "node missing",
			path:     "/random",
			children: nil,
		},
		{
			name:     "parent node missing",
			path:     "/x/y/z",
			children: nil,
		},
		{
			name:     "parent exists, child missing",
			path:     fmt.Sprintf("/%s/random", rootChildName),
			children: nil,
		},
		{
			name:     "node with children",
			path:     "/" + rootChildName,
			children: []string{childChildName},
		},
		{
			name:     "leaf node",
			path:     fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			children: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			zk := NewServer()
			req := &pbzk.CreateRequest{
				Path: "/" + rootChildName,
			}
			_, err := zk.Create(ctx, req)
			require.NoError(t, err)
			req = &pbzk.CreateRequest{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			_, err = zk.Create(ctx, req)
			require.NoError(t, err)

			// Set the versions on the nodes so we can differentiate between missing nodes and nodes that exists but
			// have no data.
			rootChild := zk.root.Children[rootChildName]
			rootChild.Version = 5
			rootChild.Children[childChildName].Version = 5

			cReq := &pbzk.GetChildrenRequest{
				Path: test.path,
			}
			cResp, err := zk.GetChildren(ctx, cReq)
			if test.errorExpected {
				assert.Empty(t, cResp.GetChildren())
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.children, cResp.GetChildren())
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_NewFullName(t *testing.T) {
	tests := []struct {
		name           string
		nodeName       string
		ancestorsNames []string
		expectedResult string
	}{
		{
			name:           "no ancestors",
			nodeName:       "node",
			ancestorsNames: nil,
			expectedResult: "/node",
		},
		{
			name:           "1 ancestor",
			nodeName:       "node",
			ancestorsNames: []string{"a1"},
			expectedResult: "/a1/node",
		},
		{
			name:           "multiple ancestors",
			nodeName:       "node",
			ancestorsNames: []string{"a1", "a2", "a3"},
			expectedResult: "/a1/a2/a3/node",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualResult := newFullName(test.nodeName, test.ancestorsNames)
			assert.Equal(t, test.expectedResult, actualResult)
		})
	}
}

func TestServer_ExtractWatches(t *testing.T) {
	tests := []struct {
		name               string
		checkingWatchType  pbzk.WatchEvent_EventType
		clientIDsExtracted []string
		clientIDsLeft      []string
	}{
		{
			name:               "created",
			checkingWatchType:  pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED,
			clientIDsExtracted: []string{"created", "all"},
			clientIDsLeft:      []string{"deleted", "data changed", "children changed"},
		},
		{
			name:               "deleted",
			checkingWatchType:  pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
			clientIDsExtracted: []string{"deleted", "all"},
			clientIDsLeft:      []string{"created", "data changed", "children changed"},
		},
		{
			name:               "data changed",
			checkingWatchType:  pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
			clientIDsExtracted: []string{"data changed", "all"},
			clientIDsLeft:      []string{"created", "deleted", "children changed"},
		},
		{
			name:               "children changed",
			checkingWatchType:  pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED,
			clientIDsExtracted: []string{"children changed", "all"},
			clientIDsLeft:      []string{"created", "deleted", "data changed"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewServer()
			s.watches = map[string][]*znode.Watch{
				"/zoo": {
					{
						ClientID: "created",
						Path:     "/zoo",
						WatchTypes: []pbzk.WatchEvent_EventType{
							pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED,
						},
					},
					{
						ClientID: "deleted",
						Path:     "/zoo",
						WatchTypes: []pbzk.WatchEvent_EventType{
							pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
						},
					},
					{
						ClientID: "data changed",
						Path:     "/zoo",
						WatchTypes: []pbzk.WatchEvent_EventType{
							pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
						},
					},
					{
						ClientID: "children changed",
						Path:     "/zoo",
						WatchTypes: []pbzk.WatchEvent_EventType{
							pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED,
						},
					},
					{
						ClientID: "all",
						Path:     "/zoo",
						WatchTypes: []pbzk.WatchEvent_EventType{
							pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED,
							pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
							pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
							pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED,
						},
					},
				},
			}

			extractedWatches := s.extractWatches("/zoo", test.checkingWatchType)

			var actualExtractedClientIDs []string
			for _, w := range extractedWatches {
				actualExtractedClientIDs = append(actualExtractedClientIDs, w.ClientID)
			}
			var actualClientIDsLeft []string
			for _, w := range s.watches["/zoo"] {
				fmt.Println(w)
				actualClientIDsLeft = append(actualClientIDsLeft, w.ClientID)
			}
			assert.Equal(t, test.clientIDsExtracted, actualExtractedClientIDs)
			assert.Equal(t, test.clientIDsLeft, actualClientIDsLeft)
		})
	}
}

func TestServer_GetParent(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		parentPath string
	}{
		{
			name:       "empty path",
			path:       "",
			parentPath: "",
		},
		{
			name:       "no parent",
			path:       "/zoo",
			parentPath: "",
		},
		{
			name:       "has parent",
			path:       "/zoo/giraffe",
			parentPath: "/zoo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualParent := getParent(test.path)
			assert.Equal(t, test.parentPath, actualParent)
		})
	}
}
