package tests

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"testing"
	"time"

	zkc "github.com/mikekulinski/zookeeper/pkg/client"
	zks "github.com/mikekulinski/zookeeper/pkg/server"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	serverAddress = "localhost"
)

type integrationTestSuite struct {
	suite.Suite
	Server *grpc.Server
}

func (i *integrationTestSuite) SetupTest() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	zk := zks.NewServer()
	pbzk.RegisterZookeeperServer(s, zk)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	i.Server = s
}

func (i *integrationTestSuite) TearDownTest() {
	i.Server.GracefulStop()
}

// TODO: Consider adding the uber goroutine leak checker here.
func (i *integrationTestSuite) TestCreateThenGetData() {
	ctx := context.Background()

	client := zkc.NewClient(serverAddress)
	err := client.Connect(ctx)
	i.Require().NoError(err)

	requests := []*pbzk.ZookeeperRequest{
		{
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo",
					Data: []byte("Secrets hahahahaha!!"),
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo/giraffe",
					Data: []byte("More secrets"),
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path: "/zoo",
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path: "/zoo/giraffe",
				},
			},
		},
	}
	expectedResponses := []*pbzk.ZookeeperResponse{
		{
			Message: &pbzk.ZookeeperResponse_Create{
				Create: &pbzk.CreateResponse{
					ZNodeName: "/zoo",
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_Create{
				Create: &pbzk.CreateResponse{
					ZNodeName: "/zoo/giraffe",
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{
					Data:    []byte("Secrets hahahahaha!!"),
					Version: 0,
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{
					Data:    []byte("More secrets"),
					Version: 0,
				},
			},
		},
	}

	responses, err := sendAllRequests(client, requests, 0)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}
}

func (i *integrationTestSuite) TestWatchEvents() {
	ctx := context.Background()

	client := zkc.NewClient(serverAddress)
	err := client.Connect(ctx)
	i.Require().NoError(err)

	requests := []*pbzk.ZookeeperRequest{
		{
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo",
					Data: []byte("Secrets hahahahaha!!"),
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path:  "/zoo",
					Watch: true,
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_SetData{
				SetData: &pbzk.SetDataRequest{
					Path: "/zoo",
					Data: []byte("This one is better"),
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path: "/zoo",
				},
			},
		},
	}
	expectedResponses := []*pbzk.ZookeeperResponse{
		{
			Message: &pbzk.ZookeeperResponse_Create{
				Create: &pbzk.CreateResponse{
					ZNodeName: "/zoo",
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{
					Data:    []byte("Secrets hahahahaha!!"),
					Version: 0,
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_SetData{
				SetData: &pbzk.SetDataResponse{},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_WatchEvent{
				WatchEvent: &pbzk.WatchEvent{
					Type: pbzk.WatchEvent_EVENT_TYPE_ZNODE_DATA_CHANGED,
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{
					Data:    []byte("This one is better"),
					Version: 1,
				},
			},
		},
	}

	// Wait a bit between requests to make sure we have enough time to send the watch event before closing the connection.
	responses, err := sendAllRequests(client, requests, 10*time.Millisecond)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}
}

func (i *integrationTestSuite) TestHeartbeat_KeepsConnectionAlive() {
	ctx := context.Background()

	client := zkc.NewClient(serverAddress)
	err := client.Connect(ctx)
	i.Require().NoError(err)

	requests := []*pbzk.ZookeeperRequest{
		{
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo",
					Data: []byte("Secrets hahahahaha!!"),
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path:  "/zoo",
					Watch: false,
				},
			},
		},
	}
	expectedResponses := []*pbzk.ZookeeperResponse{
		{
			Message: &pbzk.ZookeeperResponse_Create{
				Create: &pbzk.CreateResponse{
					ZNodeName: "/zoo",
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{
					Data:    []byte("Secrets hahahahaha!!"),
					Version: 0,
				},
			},
		},
	}

	// Wait longer than the idle timeout between requests to make sure we are properly
	// using heartbeats to keep the connection alive.
	responses, err := sendAllRequests(client, requests, 2*zkc.IdleTimeout)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}
}

// TestEphemeral_SessionDeletesNode verifies that we properly delete the ephemeral node once the session closes.
func (i *integrationTestSuite) TestEphemeral_SessionDeletesNode() {
	ctx := context.Background()

	// First client creates an ephemeral node, and then closes the session.
	client := zkc.NewClient(serverAddress)
	err := client.Connect(ctx)
	i.Require().NoError(err)

	requests := []*pbzk.ZookeeperRequest{
		{
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo",
					Data: []byte("Secrets hahahahaha!!"),
					Flags: []pbzk.CreateRequest_Flag{
						pbzk.CreateRequest_FLAG_EPHEMERAL,
					},
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path:  "/zoo",
					Watch: false,
				},
			},
		},
	}
	expectedResponses := []*pbzk.ZookeeperResponse{
		{
			Message: &pbzk.ZookeeperResponse_Create{
				Create: &pbzk.CreateResponse{
					ZNodeName: "/zoo",
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{
					Data:    []byte("Secrets hahahahaha!!"),
					Version: 0,
				},
			},
		},
	}

	// Send all the requests and verify the responses.
	responses, err := sendAllRequests(client, requests, 0)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}

	// Second client tries to fetch that node after the session was closed. The node should be gone.
	client = zkc.NewClient(serverAddress)
	err = client.Connect(ctx)
	i.Require().NoError(err)

	requests = []*pbzk.ZookeeperRequest{
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path:  "/zoo",
					Watch: false,
				},
			},
		},
	}
	expectedResponses = []*pbzk.ZookeeperResponse{
		{
			// We expect an empty response since the node shouldn't exist.
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{},
			},
		},
	}

	// Send all the requests and get all the responses.
	responses, err = sendAllRequests(client, requests, 0)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}
}

// TestEphemeral_NodeManuallyDeleted verifies that closing a session for an already deleted ephemeral node doesn't cause
// any problems.
func (i *integrationTestSuite) TestEphemeral_NodeManuallyDeleted() {
	ctx := context.Background()

	// First client creates an ephemeral node, and then closes the session.
	client := zkc.NewClient(serverAddress)
	err := client.Connect(ctx)
	i.Require().NoError(err)

	requests := []*pbzk.ZookeeperRequest{
		// Create a node that has a parent to verify we are handling full paths correctly.
		{
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo",
					Data: []byte("Secrets hahahahaha!!"),
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo/giraffe",
					Data: []byte("It's a tall animal"),
					Flags: []pbzk.CreateRequest_Flag{
						pbzk.CreateRequest_FLAG_EPHEMERAL,
					},
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path:  "/zoo/giraffe",
					Watch: false,
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_Delete{
				Delete: &pbzk.DeleteRequest{
					Path:    "/zoo/giraffe",
					Version: 0,
				},
			},
		},
	}
	expectedResponses := []*pbzk.ZookeeperResponse{
		{
			Message: &pbzk.ZookeeperResponse_Create{
				Create: &pbzk.CreateResponse{
					ZNodeName: "/zoo",
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_Create{
				Create: &pbzk.CreateResponse{
					ZNodeName: "/zoo/giraffe",
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{
					Data:    []byte("It's a tall animal"),
					Version: 0,
				},
			},
		},
		{
			Message: &pbzk.ZookeeperResponse_Delete{
				Delete: &pbzk.DeleteResponse{},
			},
		},
	}

	// Send all the requests and verify the responses.
	responses, err := sendAllRequests(client, requests, 0)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}

	// Second client tries to fetch that node after the session was closed. The node should be gone.
	client = zkc.NewClient(serverAddress)
	err = client.Connect(ctx)
	i.Require().NoError(err)

	requests = []*pbzk.ZookeeperRequest{
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path:  "/zoo/giraffe",
					Watch: false,
				},
			},
		},
	}
	expectedResponses = []*pbzk.ZookeeperResponse{
		{
			// We expect an empty response since the node shouldn't exist.
			Message: &pbzk.ZookeeperResponse_GetData{
				GetData: &pbzk.GetDataResponse{},
			},
		},
	}

	// Send all the requests and get all the responses.
	responses, err = sendAllRequests(client, requests, 0)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}
}

func sendAllRequests(client *zkc.Client, requests []*pbzk.ZookeeperRequest, interval time.Duration) ([]*pbzk.ZookeeperResponse, error) {
	waitc := make(chan struct{})
	var responses []*pbzk.ZookeeperResponse
	go func() {
		for {
			resp, err := client.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a message : %v", err)
				return
			}
			responses = append(responses, resp)
		}
	}()
	for _, request := range requests {
		if err := client.Send(request); err != nil {
			log.Fatalf("Failed to send a request: %v", err)
		}
		time.Sleep(interval)
	}
	err := client.Close()
	if err != nil {
		log.Fatal("failed to close the stream")
	}
	<-waitc
	return responses, nil
}

func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	suite.Run(t, new(integrationTestSuite))
}
