package client

import (
	"context"
	"testing"
	"time"

	pbzk "github.com/mikekulinski/zookeeper/proto"
	"github.com/mikekulinski/zookeeper/proto/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestClient_IdleTimeout(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	mockGrpcClient := mock_proto.NewMockZookeeperClient(ctrl)
	mockStream := mock_proto.NewMockZookeeper_MessageClient(ctrl)

	client := &Client{
		ZookeeperClient: mockGrpcClient,
		out:             make(chan *pbzk.ZookeeperRequest),
		in:              make(chan *internalResponse),
		responses:       make(chan *internalResponse),
	}

	// Set up connect to a mock version of the stream.
	mockGrpcClient.EXPECT().Message(ctx).Return(mockStream, nil)
	err := client.Connect(ctx)
	require.NoError(t, err)

	// We expect the client to try sending some heartbeats to the server.
	mockStream.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
	// Have Recv wait for longer than the timeout to verify that we will actually time out.
	mockStream.EXPECT().Recv().DoAndReturn(func() {
		time.Sleep(2 * IdleTimeout)
	})

	// We expect to timeout here.
	resp, err := client.Recv()
	assert.Nil(t, resp)
	assert.Error(t, err)
}
