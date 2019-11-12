package common

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/thomas-maurice/wgnw/proto"
)

// GetTLSConfig returns a TLS configuration
func GetTLSConfig(rootCA, clientCert, clientKey string, skipVerify bool) (*tls.Config, error) {
	rootPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	if rootCA != "" {
		b, err := ioutil.ReadFile(rootCA)
		if err != nil {
			return nil, err
		}
		if !rootPool.AppendCertsFromPEM(b) {
			return nil, errors.New("credentials: failed to append certificates to the pool")
		}
	}

	config := tls.Config{
		RootCAs:            rootPool,
		InsecureSkipVerify: skipVerify,
	}

	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {

	} else {
		config.Certificates = []tls.Certificate{cert}
	}

	config.BuildNameToCertificate()

	return &config, nil
}

// GetClient generates a client
func GetClient(addr string, useTLS bool, config *tls.Config) (proto.WireguardServiceClient, error) {
	var conn *grpc.ClientConn
	var err error
	if !useTLS {
		conn, err = grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = grpc.Dial(addr,
			grpc.WithTransportCredentials(credentials.NewTLS(config)),
			grpc.FailOnNonTempDialError(true),
			grpc.WithBackoffConfig(grpc.DefaultBackoffConfig),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, err
		}
	}
	client := proto.NewWireguardServiceClient(conn)
	return client, nil
}

// GetContext generates a context that has a client token
func GetContext(authToken string) context.Context {
	ctx := context.Background()

	return metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("auth-token", authToken),
	)
}
