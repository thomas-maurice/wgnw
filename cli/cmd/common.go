package cmd

import (
	"google.golang.org/grpc"

	"github.com/thomas-maurice/wgnw/proto"
)

func getClient() (proto.WireguardServiceClient, error) {
	conn, err := grpc.Dial(controllerAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := proto.NewWireguardServiceClient(conn)
	return client, nil
}
