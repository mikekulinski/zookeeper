package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/mikekulinski/zookeeper/pkg/znode"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			// Pre-init the server with some nodes so we can also test cases with existing nodes.
			zk.root.Children = map[string]*znode.ZNode{
				existingNodeName: znode.NewZNode(existingNodeName, 0, test.parentNodeType, nil),
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
			existingNode := znode.NewZNode(existingNodeName, 0, test.parentNodeType, nil)
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
				assert.Equal(t, "new_5", resp.GetZNodeName())
				assert.NoError(t, err)
			}
		})
	}
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
