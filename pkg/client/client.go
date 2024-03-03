package client

import (
	"context"
	"log"

	"github.com/google/uuid"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	ClientIDHeader = "X-Client-ID"
)

type Client struct {
	pbzk.ZookeeperClient

	clientID string
}

func NewClient(endpoint string) (*Client, error) {
	clientID := uuid.New().String()

	// Set up a connection to the server.
	conn, err := grpc.Dial(
		endpoint+":8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainStreamInterceptor(clientIDStreamInterceptor(clientID)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	grpcClient := pbzk.NewZookeeperClient(conn)

	// Initiate the connection with the Zookeeper server.
	return &Client{
		ZookeeperClient: grpcClient,
		clientID:        clientID,
	}, nil
}

// clientIDStreamInterceptor returns a gRPC stream interceptor that adds a client ID to outgoing streams.
func clientIDStreamInterceptor(clientID string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Add client ID to outgoing metadata
		md := metadata.Pairs(ClientIDHeader, clientID)
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Call the streamer to establish a client stream
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		return clientStream, nil
	}
}
