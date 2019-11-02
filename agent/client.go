package main

import (
	"google.golang.org/grpc"

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
