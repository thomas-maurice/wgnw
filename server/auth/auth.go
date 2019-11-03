package auth

import (
	"context"
	"crypto/sha512"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func extractTokenFromMetadata(md metadata.MD) (string, bool) {
	r := md.Get("auth-token")
	if len(r) == 0 {
		return "", false
	}

	return r[0], true
}

func NewAuthFunction(accessTokenHash string) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		if accessTokenHash == "" {
			return ctx, nil
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ctx, nil
		}
		token, found := extractTokenFromMetadata(md)
		if !found {
			return nil, grpc.Errorf(codes.Unauthenticated, "unauthenticated")
		}

		h := sha512.New()
		io.WriteString(h, token)
		hashedKey := fmt.Sprintf("%x", h.Sum(nil))

		if hashedKey == accessTokenHash {
			return ctx, nil
		}

		return nil, grpc.Errorf(codes.Unauthenticated, "invalid auth token")
	}
}
