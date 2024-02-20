package server

import (
	"fmt"
	"testing"

	"github.com/mikekulinski/zookeeper/pkg/znode"
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

			_, err := zk.Create(test.path, nil)
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

			actualName, err := zk.Create(test.path, nil, SEQUENTIAL)
			if test.errorExpected {
				assert.Empty(t, actualName)
				assert.Error(t, err)
			} else {
				assert.Equal(t, "new_5", actualName)
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
			_, err := zk.Create("/"+rootChildName, nil)
			require.NoError(t, err)
			_, err = zk.Create(fmt.Sprintf("/%s/%s", rootChildName, childChildName), nil)
			require.NoError(t, err)

			err = zk.Delete(test.path, test.version)
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
			_, err := zk.Create("/"+rootChildName, nil)
			require.NoError(t, err)
			_, err = zk.Create(fmt.Sprintf("/%s/%s", rootChildName, childChildName), nil)
			require.NoError(t, err)

			exists, err := zk.Exists(test.path, false)
			if test.errorExpected {
				assert.False(t, exists)
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.exists, exists)
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
			_, err := zk.Create("/"+rootChildName, nil)
			require.NoError(t, err)
			_, err = zk.Create(fmt.Sprintf("/%s/%s", rootChildName, childChildName), nil)
			require.NoError(t, err)
			// Delete all those nodes to verify we deleted them.
			err = zk.Delete(fmt.Sprintf("/%s/%s", rootChildName, childChildName), -1)
			require.NoError(t, err)
			err = zk.Delete("/"+rootChildName, -1)
			require.NoError(t, err)

			exists, err := zk.Exists(test.path, false)
			assert.Equal(t, test.exists, exists)
			assert.NoError(t, err)
		})
	}
}
