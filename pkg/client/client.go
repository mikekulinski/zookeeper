package client

import (
	"log"

	"github.com/google/uuid"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	pbzk.ZookeeperClient

	clientID string
}

func NewClient(endpoint string) (*Client, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(endpoint+":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	grpcClient := pbzk.NewZookeeperClient(conn)

	// Initiate the connection with the Zookeeper server.
	clientID := uuid.New().String()
	// TODO: Include the clientID in the headers of every request.
	return &Client{
		ZookeeperClient: grpcClient,
		clientID:        clientID,
	}, nil
}
