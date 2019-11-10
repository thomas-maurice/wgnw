package cmd

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/thomas-maurice/wgnw/common"
	"github.com/thomas-maurice/wgnw/proto"
)

func getClient() (proto.WireguardServiceClient, error) {
	if useTLS {
		tlsConfig, err := common.GetTLSConfig(caCert, certFile, keyFile, insecureSkipVerify)

		if err != nil {
			logrus.WithError(err).Fatal("Could not setup TLS client")
		}
		return common.GetClient(controllerAddress, useTLS, tlsConfig)
	} else {
		return common.GetClient(controllerAddress, useTLS, nil)
	}
}

func getContext() context.Context {
	ctx := context.Background()

	return metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("auth-token", authToken),
	)
}
