package znode

import (
	"fmt"
	"testing"

	pbzk "github.com/mikekulinski/zookeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Also add tests with create, then delete, then get.
func TestDB_Get(t *testing.T) {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name          string
		path          string
		node          *ZNode
		errorExpected bool
	}{
		{
			name: "node missing",
			path: "/random",
			node: nil,
		},
		{
			name: "parent node missing",
			path: "/x/y/z",
			node: nil,
		},
		{
			name: "parent exists, child missing",
			path: fmt.Sprintf("/%s/random", rootChildName),
			node: nil,
		},
		{
			name: "node exists, root",
			path: "/" + rootChildName,
			node: &ZNode{
				Name:     "/" + rootChildName,
				NodeType: ZNodeType_STANDARD,
				Data:     []byte(rootChildName),
			},
		},
		{
			name: "node exists, child of another node",
			path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			node: &ZNode{
				Name:     fmt.Sprintf("/%s/%s", rootChildName, childChildName),
				NodeType: ZNodeType_STANDARD,
				Data:     []byte(childChildName),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := NewDB()
			// Add the parent node.
			txn := &pbzk.Transaction{
				Txn: &pbzk.Transaction_Create{
					Create: &pbzk.CreateTxn{
						Path: "/" + rootChildName,
						Data: []byte(rootChildName),
					},
				},
			}
			_, err := db.Create(txn)
			require.NoError(t, err)
			// Add the child node.
			txn = &pbzk.Transaction{
				Txn: &pbzk.Transaction_Create{
					Create: &pbzk.CreateTxn{
						Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
						Data: []byte(childChildName),
					},
				},
			}
			_, err = db.Create(txn)
			require.NoError(t, err)

			node := db.Get(test.path)
			if test.node == nil || node == nil {
				// If at least one is nil, then check that they are both nil.
				assert.Nil(t, test.node)
				assert.Nil(t, node)
			} else {
				// Otherwise verify the data is equal.
				assert.Equal(t, test.node.Name, node.Name)
				assert.Equal(t, test.node.NodeType, node.NodeType)
				assert.Equal(t, test.node.Data, node.Data)
			}
			assert.NoError(t, err)
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
