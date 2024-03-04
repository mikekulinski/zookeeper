package server

import (
	"context"

	"github.com/mikekulinski/zookeeper/pkg/client"
	"google.golang.org/grpc/metadata"
)

// ExtractClientIDHeader extracts the clientID from the context.
func ExtractClientIDHeader(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	values := md.Get(client.ClientIDHeader)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}
