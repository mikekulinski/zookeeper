package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/mikekulinski/zookeeper/pkg/session"
	"github.com/mikekulinski/zookeeper/pkg/utils"
	"github.com/mikekulinski/zookeeper/pkg/znode"
	mock_db "github.com/mikekulinski/zookeeper/pkg/znode/mocks"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type serverTestSuite struct {
	suite.Suite
	MockDB *mock_db.MockZKDB
	ZK     *Server
}

func (s *serverTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.MockDB = mock_db.NewMockZKDB(ctrl)

	s.ZK = NewServer()
	s.ZK.db = s.MockDB
}

// TestServer_Create_Standard verifies that we handle create edge cases properly for normal nodes.
func (s *serverTestSuite) TestServer_Create_Standard() {
	var clientID = uuid.New().String()
	tests := []struct {
		name          string
		path          string
		testFunc      func()
		errorExpected bool
	}{
		{
			name:          "invalid path",
			path:          "invalid",
			testFunc:      func() {},
			errorExpected: true,
		},
		{
			name: "error with create",
			path: "/x/y/z",
			testFunc: func() {
				s.MockDB.EXPECT().Create(gomock.Any()).Return(nil, fmt.Errorf("error with create"))
			},
			errorExpected: true,
		},
		{
			name: "valid create, standard node",
			path: "/xyz",
			testFunc: func() {
				s.MockDB.EXPECT().Create(gomock.Any()).Return(
					znode.NewZNode(
						"/xyz",
						znode.ZNodeType_STANDARD,
						clientID,
						[]byte("data"),
					),
					nil,
				)
			},
			errorExpected: false,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			ctx := context.Background()
			ctx = utils.SetIncomingClientIDHeader(ctx, clientID)

			// Make sure to run the function for each test that sets up the server state.
			test.testFunc()

			req := &pbzk.CreateRequest{
				Path: test.path,
			}
			_, err := s.ZK.Create(ctx, req)
			if test.errorExpected {
				s.Assert().Error(err)
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

// TestServer_Create_Ephemeral verifies that we handle the logic around sessions and ephemeral nodes properly.
func (s *serverTestSuite) TestServer_Create_Ephemeral() {
	var clientID = uuid.New().String()
	newNode := znode.NewZNode(
		"/xyz",
		znode.ZNodeType_EPHEMERAL,
		clientID,
		[]byte("data"),
	)
	tests := []struct {
		name          string
		path          string
		testFunc      func()
		errorExpected bool
	}{
		{
			name: "no session for ephemeral",
			path: "/xyz",
			testFunc: func() {
				// Make sure to remove the session if one exists.
				delete(s.ZK.sessions, clientID)
				s.MockDB.EXPECT().Create(gomock.Any()).Return(newNode, nil)
			},
			errorExpected: true,
		},
		{
			name: "valid create, ephemeral node",
			path: "/xyz",
			testFunc: func() {
				// Make sure to create a session since it's required to be valid.
				s.ZK.sessions[clientID] = session.NewSession()
				s.MockDB.EXPECT().Create(gomock.Any()).Return(newNode, nil)
			},
			errorExpected: false,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			ctx := context.Background()
			ctx = utils.SetIncomingClientIDHeader(ctx, clientID)

			// Make sure to run the function for each test that sets up the server state.
			test.testFunc()

			req := &pbzk.CreateRequest{
				Path: test.path,
			}
			_, err := s.ZK.Create(ctx, req)
			if test.errorExpected {
				s.Assert().Error(err)
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

//	func (s *serverTestSuite) TestServer_Create_Sequential() {
//		const existingNodeName = "existing_5"
//		tests := []struct {
//			name           string
//			path           string
//			parentNodeType znode.ZNodeType
//			errorExpected  bool
//		}{
//			{
//				name:          "node already exists",
//				path:          "/existing",
//				errorExpected: true,
//			},
//			{
//				name:          "valid create, root",
//				path:          "/new",
//				errorExpected: false,
//			},
//			{
//				name:          "valid create, child of existing node",
//				path:          fmt.Sprintf("/%s/new", existingNodeName),
//				errorExpected: false,
//			},
//			{
//				name:           "parent node is ephemeral",
//				path:           fmt.Sprintf("/%s/new", existingNodeName),
//				parentNodeType: znode.ZNodeType_EPHEMERAL,
//				errorExpected:  true,
//			},
//		}
//		for _, test := range tests {
//			s.Run(test.name, func() {
//				ctx := context.Background()
//				zk := NewServer()
//				// Pre-init the server with some nodes so we can also test cases with existing nodes.
//				zk.root.NextSequentialNode = 5
//				existingNode := znode.NewZNode(existingNodeName, test.parentNodeType, "", nil)
//				existingNode.NextSequentialNode = 5
//				zk.root.Children = map[string]*znode.ZNode{
//					existingNodeName: existingNode,
//				}
//
//				req := &pbzk.CreateRequest{
//					Path:  test.path,
//					Flags: []pbzk.CreateRequest_Flag{pbzk.CreateRequest_FLAG_SEQUENTIAL},
//				}
//				resp, err := zk.Create(ctx, req)
//				if test.errorExpected {
//					s.Assert().Empty(resp.GetZNodeName())
//					s.Assert().Error(err)
//				} else {
//					fmt.Println(resp.GetZNodeName())
//					s.Assert().Contains(resp.GetZNodeName(), "new_5")
//					s.Assert().NoError(err)
//				}
//			})
//		}
//	}
//
//	func (s *serverTestSuite) TestServer_Create_RegularThenSequential() {
//		const existingNodeName = "existing"
//		ctx := context.Background()
//		zk := NewServer()
//		// Pre-init the server with some nodes so we can also test cases with existing nodes.
//		zk.root.NextSequentialNode = 5
//		existingNode := znode.NewZNode(existingNodeName, znode.ZNodeType_STANDARD, "", nil)
//		existingNode.NextSequentialNode = 5
//		zk.root.Children = map[string]*znode.ZNode{
//			existingNodeName: existingNode,
//		}
//
//		// Create a regular node. It shouldn't have the sequential suffix.
//		req := &pbzk.CreateRequest{
//			Path: "/" + existingNodeName + "/standard",
//		}
//		resp, err := zk.Create(ctx, req)
//		fmt.Println(resp.GetZNodeName())
//		s.Assert().True(strings.HasSuffix(resp.GetZNodeName(), "standard"))
//		s.Assert().NoError(err)
//		// NextSequentialNode shouldn't be incremented.
//		s.Assert().Equal(5, existingNode.NextSequentialNode)
//
//		// Create a sequential node. It should use 5 as the suffix. NextSequentialNode should only be incremented
//		// after creating another sequential node.
//		req = &pbzk.CreateRequest{
//			Path:  "/" + existingNodeName + "/sequential",
//			Flags: []pbzk.CreateRequest_Flag{pbzk.CreateRequest_FLAG_SEQUENTIAL},
//		}
//		resp, err = zk.Create(ctx, req)
//		fmt.Println(resp.GetZNodeName())
//		s.Assert().True(strings.HasSuffix(resp.GetZNodeName(), "sequential_5"))
//		s.Assert().NoError(err)
//		// NextSequentialNode should be incremented.
//		s.Assert().Equal(6, existingNode.NextSequentialNode)
//	}
//
//	func (s *serverTestSuite) TestServer_Delete() {
//		const rootChildName = "rootChild"
//		const childChildName = "childChild"
//		tests := []struct {
//			name            string
//			path            string
//			version         int64
//			errorExpected   bool
//			deleteSucceeded bool
//		}{
//			{
//				name:          "invalid path",
//				path:          "invalid",
//				errorExpected: true,
//			},
//			{
//				name:          "parent node missing",
//				path:          "/x/y/z",
//				errorExpected: true,
//			},
//			{
//				name:          "node missing, no error",
//				path:          "/random",
//				errorExpected: false,
//			},
//			{
//				name:          "invalid version",
//				path:          "/" + rootChildName,
//				version:       24,
//				errorExpected: true,
//			},
//			{
//				name:          "invalid delete, node has children",
//				path:          "/" + rootChildName,
//				version:       0,
//				errorExpected: true,
//			},
//			{
//				name:            "valid delete, child of existing node",
//				path:            fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//				version:         0,
//				errorExpected:   false,
//				deleteSucceeded: true,
//			},
//			{
//				name:            "valid delete, ignore version check",
//				path:            fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//				version:         -1,
//				errorExpected:   false,
//				deleteSucceeded: true,
//			},
//		}
//		for _, test := range tests {
//			s.Run(test.name, func() {
//				ctx := context.Background()
//				zk := NewServer()
//				cReq := &pbzk.CreateRequest{
//					Path: "/" + rootChildName,
//				}
//				_, err := zk.Create(ctx, cReq)
//				s.Require().NoError(err)
//				cReq = &pbzk.CreateRequest{
//					Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//				}
//				_, err = zk.Create(ctx, cReq)
//				s.Require().NoError(err)
//
//				dReq := &pbzk.DeleteRequest{
//					Path:    test.path,
//					Version: test.version,
//				}
//				_, err = zk.Delete(ctx, dReq)
//				if test.errorExpected {
//					s.Assert().Error(err)
//				} else {
//					s.Assert().NoError(err)
//				}
//
//				fmt.Println(zk.root)
//				// If we successfully deleted the leaf node, then verify we don't see it in the trie.
//				if test.deleteSucceeded {
//					s.Assert().Empty(zk.root.Children[rootChildName].Children)
//				}
//			})
//		}
//	}

func (s *serverTestSuite) TestServer_Exists_NodeCreated() {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name          string
		path          string
		node          *znode.ZNode
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
			node:   nil,
			exists: false,
		},
		{
			name: "node exists, root",
			path: "/" + rootChildName,
			node: &znode.ZNode{
				Name:    "/" + rootChildName,
				Version: 2,
				Data:    []byte("secret stuff"),
			},
			exists: true,
		},
		{
			name: "node exists, child of another node",
			path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			node: &znode.ZNode{
				Name:    fmt.Sprintf("/%s/%s", rootChildName, childChildName),
				Version: 10,
				Data:    []byte("secret stuff"),
			},
			exists: true,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			ctx := context.Background()

			if !test.errorExpected {
				s.MockDB.EXPECT().Get(test.path).Return(test.node)
			}

			eReq := &pbzk.ExistsRequest{
				Path: test.path,
			}
			eResp, err := s.ZK.Exists(ctx, eReq)
			if test.errorExpected {
				s.Assert().False(eResp.GetExists())
				s.Assert().Error(err)
			} else {
				s.Assert().Equal(test.exists, eResp.GetExists())
				s.Assert().NoError(err)
			}
		})
	}
}

func (s *serverTestSuite) TestServer_GetData() {
	const rootChildName = "rootChild"
	const childChildName = "childChild"
	tests := []struct {
		name          string
		path          string
		node          *znode.ZNode
		errorExpected bool
	}{
		{
			name:          "invalid path",
			path:          "invalid",
			errorExpected: true,
		},
		{
			name: "node missing",
			path: "/random",
			node: nil,
		},
		{
			name: "node exists, root",
			path: "/" + rootChildName,
			node: &znode.ZNode{
				Name:    "/" + rootChildName,
				Version: 2,
				Data:    []byte("secret stuff"),
			},
		},
		{
			name: "node exists, child of another node",
			path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
			node: &znode.ZNode{
				Name:    fmt.Sprintf("/%s/%s", rootChildName, childChildName),
				Version: 10,
				Data:    []byte("secret stuff"),
			},
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			ctx := context.Background()

			if !test.errorExpected {
				s.MockDB.EXPECT().Get(test.path).Return(test.node)
			}

			gReq := &pbzk.GetDataRequest{
				Path: test.path,
			}
			gResp, err := s.ZK.GetData(ctx, gReq)
			if test.node == nil {
				s.Assert().Empty(gResp.GetData())
				s.Assert().Zero(gResp.GetVersion())
			} else {
				s.Assert().Equal(test.node.Data, gResp.GetData())
				s.Assert().Equal(test.node.Version, gResp.GetVersion())
			}
			if test.errorExpected {
				s.Assert().Error(err)
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

//
//func (s *serverTestSuite) TestServer_SetData() {
//	const rootChildName = "rootChild"
//	const childChildName = "childChild"
//	tests := []struct {
//		name          string
//		path          string
//		version       int64
//		errorExpected bool
//		writeSucceeds bool
//	}{
//		{
//			name:          "invalid path",
//			path:          "invalid",
//			errorExpected: true,
//		},
//		{
//			name:          "node missing",
//			path:          "/random",
//			errorExpected: true,
//		},
//		{
//			name:          "parent node missing",
//			path:          "/x/y/z",
//			errorExpected: true,
//		},
//		{
//			name:          "parent exists, child missing",
//			path:          fmt.Sprintf("/%s/random", rootChildName),
//			errorExpected: true,
//		},
//		{
//			name:          "node exists, root",
//			path:          "/" + rootChildName,
//			writeSucceeds: true,
//		},
//		{
//			name:          "node exists, child of another node",
//			path:          fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//			writeSucceeds: true,
//		},
//		{
//			name:          "invalid version",
//			path:          "/" + rootChildName,
//			version:       10,
//			errorExpected: true,
//		},
//		{
//			name:          "ignore version check",
//			path:          "/" + rootChildName,
//			version:       -1,
//			writeSucceeds: true,
//		},
//	}
//	for _, test := range tests {
//		s.Run(test.name, func() {
//			ctx := context.Background()
//			dataToSet := []byte("you're a wizard Harry")
//
//			zk := NewServer()
//			req := &pbzk.CreateRequest{
//				Path: "/" + rootChildName,
//			}
//			_, err := zk.Create(ctx, req)
//			s.Require().NoError(err)
//			req = &pbzk.CreateRequest{
//				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//			}
//			_, err = zk.Create(ctx, req)
//			s.Require().NoError(err)
//
//			sReq := &pbzk.SetDataRequest{
//				Path:    test.path,
//				Data:    dataToSet,
//				Version: test.version,
//			}
//			_, err = zk.SetData(ctx, sReq)
//			if test.errorExpected {
//				s.Assert().Error(err)
//			} else {
//				s.Assert().NoError(err)
//			}
//
//			if test.writeSucceeds {
//				gReq := &pbzk.GetDataRequest{
//					Path: test.path,
//				}
//				gResp, err := zk.GetData(ctx, gReq)
//				s.Assert().Equal(dataToSet, gResp.GetData())
//				// We expect the set to increment the version.
//				s.Assert().EqualValues(1, gResp.GetVersion())
//				s.Assert().NoError(err)
//			}
//		})
//	}
//}
//
//func (s *serverTestSuite) TestServer_GetChildren() {
//	const rootChildName = "rootChild"
//	const childChildName = "childChild"
//	tests := []struct {
//		name          string
//		path          string
//		children      []string
//		errorExpected bool
//	}{
//		{
//			name:          "invalid path",
//			path:          "invalid",
//			errorExpected: true,
//		},
//		{
//			name:     "node missing",
//			path:     "/random",
//			children: nil,
//		},
//		{
//			name:     "parent node missing",
//			path:     "/x/y/z",
//			children: nil,
//		},
//		{
//			name:     "parent exists, child missing",
//			path:     fmt.Sprintf("/%s/random", rootChildName),
//			children: nil,
//		},
//		{
//			name:     "node with children",
//			path:     "/" + rootChildName,
//			children: []string{childChildName},
//		},
//		{
//			name:     "leaf node",
//			path:     fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//			children: nil,
//		},
//	}
//	for _, test := range tests {
//		s.Run(test.name, func() {
//			ctx := context.Background()
//			zk := NewServer()
//			req := &pbzk.CreateRequest{
//				Path: "/" + rootChildName,
//			}
//			_, err := zk.Create(ctx, req)
//			s.Require().NoError(err)
//			req = &pbzk.CreateRequest{
//				Path: fmt.Sprintf("/%s/%s", rootChildName, childChildName),
//			}
//			_, err = zk.Create(ctx, req)
//			s.Require().NoError(err)
//
//			// Set the versions on the nodes so we can differentiate between missing nodes and nodes that exists but
//			// have no data.
//			rootChild := zk.root.Children[rootChildName]
//			rootChild.Version = 5
//			rootChild.Children[childChildName].Version = 5
//
//			cReq := &pbzk.GetChildrenRequest{
//				Path: test.path,
//			}
//			cResp, err := zk.GetChildren(ctx, cReq)
//			if test.errorExpected {
//				s.Assert().Empty(cResp.GetChildren())
//				s.Assert().Error(err)
//			} else {
//				s.Assert().Equal(test.children, cResp.GetChildren())
//				s.Assert().NoError(err)
//			}
//		})
//	}
//}
//
//func (s *serverTestSuite) TestServer_NewFullName() {
//	tests := []struct {
//		name           string
//		nodeName       string
//		ancestorsNames []string
//		expectedResult string
//	}{
//		{
//			name:           "no ancestors",
//			nodeName:       "node",
//			ancestorsNames: nil,
//			expectedResult: "/node",
//		},
//		{
//			name:           "1 ancestor",
//			nodeName:       "node",
//			ancestorsNames: []string{"a1"},
//			expectedResult: "/a1/node",
//		},
//		{
//			name:           "multiple ancestors",
//			nodeName:       "node",
//			ancestorsNames: []string{"a1", "a2", "a3"},
//			expectedResult: "/a1/a2/a3/node",
//		},
//	}
//	for _, test := range tests {
//		s.Run(test.name, func() {
//			actualResult := newFullName(test.nodeName, test.ancestorsNames)
//			s.Assert().Equal(test.expectedResult, actualResult)
//		})
//	}
//}
//
//func (s *serverTestSuite) TestServer_ExtractWatches() {
//	tests := []struct {
//		name               string
//		checkingWatchType  pbzk.WatchEvent_EventType
//		clientIDsExtracted []string
//		clientIDsLeft      []string
//	}{
//		{
//			name:               "created",
//			checkingWatchType:  pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED,
//			clientIDsExtracted: []string{"created", "all"},
//			clientIDsLeft:      []string{"deleted", "data changed", "children changed"},
//		},
//		{
//			name:               "deleted",
//			checkingWatchType:  pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
//			clientIDsExtracted: []string{"deleted", "all"},
//			clientIDsLeft:      []string{"created", "data changed", "children changed"},
//		},
//		{
//			name:               "data changed",
//			checkingWatchType:  pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
//			clientIDsExtracted: []string{"data changed", "all"},
//			clientIDsLeft:      []string{"created", "deleted", "children changed"},
//		},
//		{
//			name:               "children changed",
//			checkingWatchType:  pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED,
//			clientIDsExtracted: []string{"children changed", "all"},
//			clientIDsLeft:      []string{"created", "deleted", "data changed"},
//		},
//	}
//	for _, test := range tests {
//		s.Run(test.name, func() {
//			s := NewServer()
//			s.watches = map[string][]*znode.Watch{
//				"/zoo": {
//					{
//						ClientID: "created",
//						Path:     "/zoo",
//						WatchTypes: []pbzk.WatchEvent_EventType{
//							pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED,
//						},
//					},
//					{
//						ClientID: "deleted",
//						Path:     "/zoo",
//						WatchTypes: []pbzk.WatchEvent_EventType{
//							pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
//						},
//					},
//					{
//						ClientID: "data changed",
//						Path:     "/zoo",
//						WatchTypes: []pbzk.WatchEvent_EventType{
//							pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
//						},
//					},
//					{
//						ClientID: "children changed",
//						Path:     "/zoo",
//						WatchTypes: []pbzk.WatchEvent_EventType{
//							pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED,
//						},
//					},
//					{
//						ClientID: "all",
//						Path:     "/zoo",
//						WatchTypes: []pbzk.WatchEvent_EventType{
//							pbzk.WatchEvent_EVENT_TYPE_ZNODE_CREATED,
//							pbzk.WatchEvent_EVENT_TYPE_ZNODE_DELETED,
//							pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
//							pbzk.WatchEvent_EVENT_TYPE_ZNODE_CHILDREN_CHANGED,
//						},
//					},
//				},
//			}
//
//			extractedWatches := s.extractWatches("/zoo", test.checkingWatchType)
//
//			var actualExtractedClientIDs []string
//			for _, w := range extractedWatches {
//				actualExtractedClientIDs = append(actualExtractedClientIDs, w.ClientID)
//			}
//			var actualClientIDsLeft []string
//			for _, w := range s.watches["/zoo"] {
//				fmt.Println(w)
//				actualClientIDsLeft = append(actualClientIDsLeft, w.ClientID)
//			}
//			s.Assert().Equal(test.clientIDsExtracted, actualExtractedClientIDs)
//			s.Assert().Equal(test.clientIDsLeft, actualClientIDsLeft)
//		})
//	}
//}
//
//func (s *serverTestSuite) TestServer_GetParent() {
//	tests := []struct {
//		name       string
//		path       string
//		parentPath string
//	}{
//		{
//			name:       "empty path",
//			path:       "",
//			parentPath: "",
//		},
//		{
//			name:       "no parent",
//			path:       "/zoo",
//			parentPath: "",
//		},
//		{
//			name:       "has parent",
//			path:       "/zoo/giraffe",
//			parentPath: "/zoo",
//		},
//	}
//	for _, test := range tests {
//		s.Run(test.name, func() {
//			actualParent := getParent(test.path)
//			s.Assert().Equal(test.parentPath, actualParent)
//		})
//	}
//}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(serverTestSuite))
}
