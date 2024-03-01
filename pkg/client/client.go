package client

import (
	"context"
	"fmt"
	"log"
	"time"

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

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	// Initiate the connection with the Zookeeper server.
	clientID := uuid.New().String()
	req := &pbzk.ConnectRequest{
		ClientId: clientID,
	}
	_, err = grpcClient.Connect(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Zookeeper: %w", err)
	}
	return &Client{
		ZookeeperClient: grpcClient,
		clientID:        clientID,
	}, nil
}

func (c *Client) Close() error {
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	req := &pbzk.CloseRequest{
		ClientId: c.clientID,
	}
	_, err := c.ZookeeperClient.Close(ctx, req)
	if err != nil {
		return fmt.Errorf("error closing the Zookeeper connection: %w", err)
	}
	return nil
}
