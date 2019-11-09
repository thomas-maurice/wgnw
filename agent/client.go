package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/thomas-maurice/wgnw/common"
	"github.com/thomas-maurice/wgnw/proto"
)

func getClient() (proto.WireguardServiceClient, error) {
	if useTLS {
		tlsConfig, err := common.GetTLSConfig(caCert, certFile, certKeyFile, insecureSkipVerify)

		if err != nil {
			logrus.WithError(err).Fatal("Could not setup TLS listener")
		}
		return common.GetClient(svcAddr, !useTLS, tlsConfig)
	} else {
		return common.GetClient(svcAddr, useTLS, nil)
	}
}

func getContext() context.Context {
	ctx := context.Background()

	return metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("auth-token", authToken),
	)
}
