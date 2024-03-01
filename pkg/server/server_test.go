package server

import (
	"fmt"
	"testing"

	"github.com/mikekulinski/zookeeper/pkg/znode"
	"github.com/mikekulinski/zookeeper/pkg/zookeeper"
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
			zk := NewServer()
			// Pre-init the server with some nodes so we can also test cases with existing nodes.
			zk.root.Children = map[string]*znode.ZNode{
				existingNodeName: znode.NewZNode(existingNodeName, 0, test.parentNodeType, nil),
			}

			req := &zookeeper.CreateReq{
				Path: test.path,
			}
			resp := &zookeeper.CreateResp{}
			err := zk.Create(req, resp)
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
			zk := NewServer()
			// Pre-init the server with some nodes so we can also test cases with existing nodes.
			zk.root.NextSequentialNode = 5
			existingNode := znode.NewZNode(existingNodeName, 0, test.parentNodeType, nil)
			existingNode.NextSequentialNode = 5
			zk.root.Children = map[string]*znode.ZNode{
				existingNodeName: existingNode,
			}

			req := &zookeeper.CreateReq{
				Path:  test.path,
				Flags: []zookeeper.Flag{zookeeper.SEQUENTIAL},
			}
			resp := &zookeeper.CreateResp{}
			err := zk.Create(req, resp)
			if test.errorExpected {
				assert.Empty(t, resp.ZNodeName)
				assert.Error(t, err)
			} else {
				assert.Equal(t, "new_5", resp.ZNodeName)
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
		version         int
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
			zk := NewServer()
			cReq := &zookeeper.CreateReq{
				Path: "/" + rootChildName,
			}
			err := zk.Create(cReq, &zookeeper.CreateResp{})
			require.NoError(t, err)
			cReq = &zookeeper.CreateReq{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			err = zk.Create(cReq, &zookeeper.CreateResp{})
			require.NoError(t, err)

			dReq := &zookeeper.DeleteReq{
				Path:    test.path,
				Version: test.version,
			}
			err = zk.Delete(dReq, &zookeeper.DeleteResp{})
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
			zk := NewServer()
			req := &zookeeper.CreateReq{
				Path: "/" + rootChildName,
			}
			err := zk.Create(req, &zookeeper.CreateResp{})
			require.NoError(t, err)
			req = &zookeeper.CreateReq{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			err = zk.Create(req, &zookeeper.CreateResp{})
			require.NoError(t, err)

			eReq := &zookeeper.ExistsReq{
				Path: test.path,
			}
			eResp := &zookeeper.ExistsResp{}
			err = zk.Exists(eReq, eResp)
			if test.errorExpected {
				assert.False(t, eResp.Exists)
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.exists, eResp.Exists)
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
			zk := NewServer()
			// Create nodes to check for.
			cReq := &zookeeper.CreateReq{
				Path: "/" + rootChildName,
			}
			err := zk.Create(cReq, &zookeeper.CreateResp{})
			require.NoError(t, err)
			cReq = &zookeeper.CreateReq{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			err = zk.Create(cReq, &zookeeper.CreateResp{})
			require.NoError(t, err)

			// Delete all those nodes to verify we deleted them.
			dReq := &zookeeper.DeleteReq{
				Path:    fmt.Sprintf("/%s/%s", rootChildName, childChildName),
				Version: -1,
			}
			err = zk.Delete(dReq, &zookeeper.DeleteResp{})
			require.NoError(t, err)
			dReq = &zookeeper.DeleteReq{
				Path:    "/" + rootChildName,
				Version: -1,
			}
			err = zk.Delete(dReq, &zookeeper.DeleteResp{})
			require.NoError(t, err)

			eReq := &zookeeper.ExistsReq{
				Path: test.path,
			}
			eResp := &zookeeper.ExistsResp{}
			err = zk.Exists(eReq, eResp)
			assert.Equal(t, test.exists, eResp.Exists)
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
		version       int
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
			zk := NewServer()
			req := &zookeeper.CreateReq{
				Path: "/" + rootChildName,
				Data: test.data,
			}
			err := zk.Create(req, &zookeeper.CreateResp{})
			require.NoError(t, err)
			req = &zookeeper.CreateReq{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
				Data: test.data,
			}
			err = zk.Create(req, &zookeeper.CreateResp{})
			require.NoError(t, err)
			require.NoError(t, err)

			// Set the versions on the nodes so we can differentiate between missing nodes and nodes that exists but
			// have no data.
			rootChild := zk.root.Children[rootChildName]
			rootChild.Version = test.version
			rootChild.Children[childChildName].Version = test.version

			gReq := &zookeeper.GetDataReq{
				Path: test.path,
			}
			gResp := &zookeeper.GetDataResp{}
			err = zk.GetData(gReq, gResp)
			if test.errorExpected {
				assert.Empty(t, gResp.Data)
				assert.Zero(t, gResp.Version)
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.data, gResp.Data)
				assert.Empty(t, test.version, gResp.Version)
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
		version       int
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
			dataToSet := []byte("you're a wizard Harry")

			zk := NewServer()
			req := &zookeeper.CreateReq{
				Path: "/" + rootChildName,
			}
			err := zk.Create(req, &zookeeper.CreateResp{})
			require.NoError(t, err)
			req = &zookeeper.CreateReq{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			err = zk.Create(req, &zookeeper.CreateResp{})
			require.NoError(t, err)

			sReq := &zookeeper.SetDataReq{
				Path:    test.path,
				Data:    dataToSet,
				Version: test.version,
			}
			sResp := &zookeeper.SetDataResp{}
			err = zk.SetData(sReq, sResp)
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if test.writeSucceeds {
				gReq := &zookeeper.GetDataReq{
					Path: test.path,
				}
				gResp := &zookeeper.GetDataResp{}
				err := zk.GetData(gReq, gResp)
				assert.Equal(t, dataToSet, gResp.Data)
				// We expect the set to increment the version.
				assert.Equal(t, 1, gResp.Version)
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
			zk := NewServer()
			req := &zookeeper.CreateReq{
				Path: "/" + rootChildName,
			}
			err := zk.Create(req, &zookeeper.CreateResp{})
			require.NoError(t, err)
			req = &zookeeper.CreateReq{
				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			}
			err = zk.Create(req, &zookeeper.CreateResp{})
			require.NoError(t, err)

			// Set the versions on the nodes so we can differentiate between missing nodes and nodes that exists but
			// have no data.
			rootChild := zk.root.Children[rootChildName]
			rootChild.Version = 5
			rootChild.Children[childChildName].Version = 5

			cReq := &zookeeper.GetChildrenReq{
				Path: test.path,
			}
			cResp := &zookeeper.GetChildrenResp{}
			err = zk.GetChildren(cReq, cResp)
			if test.errorExpected {
				assert.Empty(t, cResp.Children)
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.children, cResp.Children)
				assert.NoError(t, err)
			}
		})
	}
}
