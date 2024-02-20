package server

import (
	"fmt"
	"testing"

	"github.com/mikekulinski/zookeeper/pkg/znode"
	"github.com/stretchr/testify/assert"
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
