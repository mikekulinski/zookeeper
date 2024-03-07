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
)

const (
	ClientIDHeader = "X-Client-ID"
	IdleTimeout    = 3000 * time.Millisecond
)

var (
	ErrIdleTimeout = fmt.Errorf("timed out waiting for server to respond")
)

type internalResponse struct {
	zkResponse *pbzk.ZookeeperResponse
	err        error
}

func (i *internalResponse) GetZkResponse() *pbzk.ZookeeperResponse {
	if i == nil {
		return nil
	}
	return i.zkResponse
}

type Client struct {
	pbzk.ZookeeperClient

	clientID string
	stream   pbzk.Zookeeper_MessageClient

	// Channels that manage the incoming and outgoing requests.
	// Channel of outgoing requests.
	out chan *pbzk.ZookeeperRequest
	// Channel of the messages we get from the server.
	in chan *internalResponse
	// Channel of the responses to return to the client.
	responses chan *internalResponse
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
		in:              make(chan *internalResponse),
		responses:       make(chan *internalResponse),
	}
	go c.continuouslySendMessages()
	go c.continuouslyReceiveMessages()
	go c.continuouslyReturnMessagesToClient()
	return c, nil
}

// Send will enqueue a new message to be sent to the server.
func (c *Client) Send(request *pbzk.ZookeeperRequest) error {
	c.out <- request
	return nil
}

// Recv tries to receive the latest message from the channel of response we have
// gotten from the server.
func (c *Client) Recv() (*pbzk.ZookeeperResponse, error) {
	resp, ok := <-c.responses
	if !ok {
		// If the channel is closed, then return io.EOF to indicate we're done.
		return nil, io.EOF
	}
	if resp.err != nil {
		return nil, resp.err
	}
	return resp.zkResponse, nil
}

// Close will close the stream to tell the server that we're no longer going to be sending
// more messages. It will also close the channel we use for sending outgoing messages
// so we can properly clean up the goroutine that reads from it.
func (c *Client) Close() error {
	err := c.stream.CloseSend()
	if err != nil {
		return fmt.Errorf("error closing send: %w", err)
	}

	// Close the outgoing channel, so we will stop our long running goroutines.
	close(c.out)
	return nil
}

// continuouslySendMessages will continuously try to send messages from our client to the server.
// If we haven't received anything from the client to send, then we will send heartbeat messages
// to keep the connection alive.
func (c *Client) continuouslySendMessages() {
	for {
		select {
		case m, ok := <-c.out:
			// If the channel is closed, then we have no more message we'll need to send.
			if !ok {
				return
			}
			// The client elected to send a message. Send that to the server.
			err := c.stream.Send(m)
			if err != nil {
				// TODO: Find a way to get this back to the client.
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

// continuouslyReturnMessagesToClient will try reading from our channel of incoming
// responses from the server. We use this channel so that we can time out if we haven't
// received a response in a long time.
func (c *Client) continuouslyReturnMessagesToClient() {
	defer close(c.responses)
	for {
		select {
		case resp, ok := <-c.in:
			// If our other goroutine closed the channel, then stop trying to process messages.
			if !ok {
				return
			}

			switch resp.GetZkResponse().GetMessage().(type) {
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
			c.responses <- &internalResponse{err: ErrIdleTimeout}
			return
		}
	}
}

// continuouslyReceiveMessages will try to receive the responses we get from the server
// and will enqueue them to be sent back to the client.
func (c *Client) continuouslyReceiveMessages() {
	defer close(c.in)
	for {
		resp, err := c.stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			c.in <- &internalResponse{
				err: fmt.Errorf("error receiving message from client stream: %w", err),
			}
			return
		}
		c.in <- &internalResponse{zkResponse: resp}
	}
}
