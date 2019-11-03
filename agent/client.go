package main

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/thomas-maurice/wgnw/proto"
)

func getClient(addr string) (proto.WireguardServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := proto.NewWireguardServiceClient(conn)
	return client, nil
}

func getContext() context.Context {
	ctx := context.Background()

	return metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("auth-token", authToken),
	)
}
