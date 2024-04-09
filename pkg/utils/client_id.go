package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	ClientIDHeader = "X-Client-ID"
)

// ExtractClientIDHeader extracts the clientID from the context.
func ExtractClientIDHeader(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	values := md.Get(ClientIDHeader)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

func SetClientIDHeader(ctx context.Context, clientID string) context.Context {
	// Add client ID to outgoing metadata
	md := metadata.Pairs(ClientIDHeader, clientID)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}
