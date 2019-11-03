package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	proto "github.com/thomas-maurice/wgnw/proto"
	"github.com/thomas-maurice/wgnw/server/auth"
	"github.com/thomas-maurice/wgnw/server/sql"
)

var (
	sqlDriver         string
	sqlConnString     string
	listenAddress     string
	promListenAddress string
	hashedAccessToken string
	debug             bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.StringVar(&sqlDriver, "sql-driver", "sqlite3", "SQL driver name, can be 'sqlite3' 'mysql' or 'postgres'")
	flag.StringVar(&listenAddress, "listen", "0.0.0.0:10000", "Address to listen on")
	flag.StringVar(&promListenAddress, "listen-prometheus", "0.0.0.0:10001", "Address to listen on for prometheus")
	flag.StringVar(&sqlConnString, "sql-string", "db.sqlite3", "SQL driver connstring")
	flag.StringVar(&hashedAccessToken, "hashed-token", "", "Auth token used to identify")
}

func main() {
	flag.Parse()

	if hashedAccessToken == "" {
		logrus.Warning("Running without an auth token, anyone can access the API")
	}

	entry := logrus.NewEntry(logrus.New())
	grpc_logrus.ReplaceGrpcLogger(entry)
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_logrus.StreamServerInterceptor(entry,
				grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
			),
			grpc_auth.StreamServerInterceptor(auth.NewAuthFunction(hashedAccessToken)),
			grpc_prometheus.StreamServerInterceptor,
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(entry,
				grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
			),
			grpc_auth.UnaryServerInterceptor(auth.NewAuthFunction(hashedAccessToken)),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	grpc_prometheus.EnableHandlingTimeHistogram()

	wgService, err := sql.NewSQLWireguardService(sqlDriver, sqlConnString, debug)
	if err != nil {
		logrus.WithError(err).Fatal("Could not create wireguard service")
	}

	wgServer, err := NewWireguardServer(wgService)
	if err != nil {
		logrus.WithError(err).Fatal("Could not create wireguard server")
	}
	proto.RegisterWireguardServiceServer(s, wgServer)
	grpc_prometheus.Register(s)

	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logrus.WithError(err).Fatal("Could not create listener")
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logrus.Fatal(http.ListenAndServe(promListenAddress, nil))
	}()

	log.Fatal(s.Serve(lis))
}
