package znode

import (
	"fmt"
	"testing"

	pbzk "github.com/mikekulinski/zookeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDB_CreateThenGet verifies that we can fetch newly created nodes.
// TODO: Refactor this to make this a little cleaner.
func TestDB_CreateThenGet(t *testing.T) {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name            string
		path            string
		parentEphemeral bool
		node            *ZNode
		errorExpected   bool
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
		{
			name:            "parent node is ephemeral",
			path:            fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			parentEphemeral: true,
			node:            nil,
			errorExpected:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := NewDB()
			// Add the parent node.
			txn := &pbzk.Transaction{
				Txn: &pbzk.Transaction_Create{
					Create: &pbzk.CreateTxn{
						Path:      "/" + rootChildName,
						Data:      []byte(rootChildName),
						Ephemeral: test.parentEphemeral,
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
			if test.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

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
		})
	}
}

// TestDB_CreateDeleteThenGet verifies that we can't find any nodes that have been deleted.
func TestDB_CreateDeleteThenGet(t *testing.T) {
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

// TODO: Write tests for checking sequential.
//// TestDB_Create_Sequential verifies that we are properly incrementing for sequential nodes.
//func TestDB_Create_Sequential(t *testing.T) {
//	const rootChildName = "rootChild"
//	const childChildName = "childChild"
//	tests := []struct {
//		name          string
//		path          string
//		parentNode    *pbzk.Transaction
//		node          *ZNode
//		errorExpected bool
//	}{
//		{
//			name: "node already exists, root",
//			path: "/" + rootChildName,
//			parentNode: &pbzk.Transaction{
//				Txn: &pbzk.Transaction_Create{
//					Create: &pbzk.CreateTxn{
//						Path: fmt.Sprintf("/%s_0", rootChildName),
//						Data: []byte(rootChildName + "_0"),
//					},
//				},
//			},
//			node: &ZNode{
//				Name:     "/" + rootChildName,
//				NodeType: ZNodeType_STANDARD,
//				Data:     []byte(rootChildName),
//			},
//			errorExpected: true,
//		},
//		// TODO:
//		//{
//		//	name: "node exists, child of another node",
//		//	path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//		//	node: &ZNode{
//		//		Name:     fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//		//		NodeType: ZNodeType_STANDARD,
//		//		Data:     []byte(childChildName),
//		//	},
//		//},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			db := NewDB()
//			// Add the parent node.
//			_, err := db.Create(test.parentNode)
//			require.NoError(t, err)
//			// Add the child node.
//			txn := &pbzk.Transaction{
//				Txn: &pbzk.Transaction_Create{
//					Create: &pbzk.CreateTxn{
//						Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//						Data: []byte(childChildName),
//						// Creating a sequential node.
//						Sequential: true,
//					},
//				},
//			}
//			_, err = db.Create(txn)
//			require.NoError(t, err)
//
//			node := db.Get(test.path)
//			if test.node == nil || node == nil {
//				// If at least one is nil, then check that they are both nil.
//				assert.Nil(t, test.node)
//				assert.Nil(t, node)
//			} else {
//				// Otherwise verify the data is equal.
//				assert.Equal(t, test.node.Name, node.Name)
//				assert.Equal(t, test.node.NodeType, node.NodeType)
//				assert.Equal(t, test.node.Data, node.Data)
//			}
//			assert.NoError(t, err)
//		})
//	}
//}

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
