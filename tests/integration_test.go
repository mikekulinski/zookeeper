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
	zookeeper "github.com/mikekulinski/zookeeper/pkg/server"
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
	zk := zookeeper.NewServer()
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

func (i *integrationTestSuite) TestCreateThenGetData() {
	ctx := context.Background()

	client, err := zkc.NewClient(serverAddress)
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

	responses, err := sendAllRequests(ctx, client, requests)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}
}

func sendAllRequests(ctx context.Context, client *zkc.Client, requests []*pbzk.ZookeeperRequest) ([]*pbzk.ZookeeperResponse, error) {
	stream, err := client.Message(ctx)
	if err != nil {
		return nil, err
	}

	waitc := make(chan struct{})
	var responses []*pbzk.ZookeeperResponse
	go func() {
		for {
			resp, err := stream.Recv()
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
		if err := stream.Send(request); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
	err = stream.CloseSend()
	if err != nil {
		log.Fatal("failed to close the stream")
	}
	<-waitc
	return responses, nil
}

func (i *integrationTestSuite) TestWatchEvents() {
	ctx := context.Background()

	client, err := zkc.NewClient(serverAddress)
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

	responses, err := sendAllRequests(ctx, client, requests)
	fmt.Println(responses)
	i.Require().NoError(err)
	for j := range expectedResponses {
		expected := expectedResponses[j]
		actual := responses[j]
		i.True(proto.Equal(expected, actual))
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	suite.Run(t, new(integrationTestSuite))
}
