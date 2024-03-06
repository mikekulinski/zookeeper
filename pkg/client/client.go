package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	ClientIDHeader = "X-Client-ID"
	IdleTimeout    = 3000 * time.Millisecond
)

type Client struct {
	pbzk.ZookeeperClient

	clientID string
	stream   pbzk.Zookeeper_MessageClient

	// Channels that manage the incoming and outgoing requests.
	// Channel of outgoing requests.
	out chan *pbzk.ZookeeperRequest
	// Channel of the messages we get from the server.
	in chan *pbzk.ZookeeperResponse
	// Channel of the responses to return to the client.
	responses chan *pbzk.ZookeeperResponse
}

func NewClient(ctx context.Context, endpoint string) (*Client, error) {
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

	// Initiate the stream with the Zookeeper server.
	stream, err := grpcClient.Message(ctx)
	if err != nil {
		log.Fatal("error initializing the stream with the server")
	}

	c := &Client{
		ZookeeperClient: grpcClient,
		clientID:        clientID,
		stream:          stream,
		out:             make(chan *pbzk.ZookeeperRequest),
		in:              make(chan *pbzk.ZookeeperResponse),
		responses:       make(chan *pbzk.ZookeeperResponse),
	}
	go c.continuouslySendMessages()
	go c.continuouslyReceiveMessages()
	go c.continuouslyReturnMessagesToClient()
	return c, nil
}

func (c *Client) Send(request *pbzk.ZookeeperRequest) error {
	c.out <- request
	return nil
}

func (c *Client) Recv() (*pbzk.ZookeeperResponse, error) {
	resp, ok := <-c.responses
	if !ok {
		// If the channel is closed, then return io.EOF to indicate we're done.
		return nil, io.EOF
	}
	return resp, nil
}

func (c *Client) Close() error {
	err := c.stream.CloseSend()
	if err != nil {
		return fmt.Errorf("error closing send: %w", err)
	}

	// Close the outgoing channel, so we will stop our long running goroutines.
	close(c.out)
	return nil
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

func (c *Client) continuouslySendMessages() {
	for {
		select {
		case m, ok := <-c.out:
			if !ok {
				return
			}
			// The client elected to send a message. Send that to the server.
			err := c.stream.Send(m)
			if err != nil {
				log.Printf("Error sending message to the client stream: %+v\n", err)
				return
			}
		case <-time.After(IdleTimeout / 3):
			// Send a heartbeat to keep the connection alive since we haven't sent a message in a bit.
			heartbeat := &pbzk.ZookeeperRequest{
				Message: &pbzk.ZookeeperRequest_Heartbeat{
					Heartbeat: &pbzk.HeartbeatRequest{
						SentTsMs: time.Now().UnixMilli(),
					},
				},
			}
			err := c.stream.Send(heartbeat)
			if err != nil {
				log.Printf("Error sending heartbeat to the client stream: %+v\n", err)
				return
			}
			log.Println("Sent heartbeat")
		}
	}
}

func (c *Client) continuouslyReturnMessagesToClient() {
	defer close(c.responses)
	for {
		select {
		case resp, ok := <-c.in:
			// If our other goroutine closed the channel, then stop trying to process messages.
			if !ok {
				return
			}

			switch resp.GetMessage().(type) {
			case *pbzk.ZookeeperResponse_Heartbeat:
				// Do nothing for heartbeat responses.
				continue
			default:
				// Enqueue the response to be sent back to the client.
				c.responses <- resp
			}
		case <-time.After(IdleTimeout):
			// We timed out waiting for the server to respond.
			// TODO: Find a new server once the server is distributed.
			log.Println("Timed out waiting for server to respond")
			return
		}
	}
}

func (c *Client) continuouslyReceiveMessages() {
	defer close(c.in)
	for {
		resp, err := c.stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			log.Printf("Error receiving message from client stream: %+v\n", err)
			return
		}
		c.in <- resp
	}
}
