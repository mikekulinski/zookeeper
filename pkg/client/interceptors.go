package client

import (
	"context"

	"github.com/mikekulinski/zookeeper/pkg/utils"
	"google.golang.org/grpc"
)

// clientIDStreamInterceptor returns a gRPC stream interceptor that adds a client ID to outgoing streams.
func clientIDStreamInterceptor(clientID string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = utils.SetClientIDHeader(ctx, clientID)

		// Call the streamer to establish a client stream
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		return clientStream, nil
	}
}
