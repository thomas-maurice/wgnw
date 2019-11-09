package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	marshaller         string
	controllerAddress  string
	authToken          string
	useTLS             bool
	certFile           string
	caCert             string
	keyFile            string
	insecureSkipVerify bool
)

var rootCmd = &cobra.Command{
	Use:   "wgnw",
	Short: "wgnw is a tool to define wg networks",
	Long:  ``,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	initNetworkCmd()
	initLeaseCmd()

	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(leaseCmd)
	rootCmd.PersistentFlags().StringVarP(&marshaller, "output", "o", "json", "Output marshaller, json or yaml")
	rootCmd.PersistentFlags().StringVarP(&controllerAddress, "controller", "c", "localhost:10000", "Controller address")
	rootCmd.PersistentFlags().StringVarP(&authToken, "token", "t", "", "Auth token to talk to the API")
	rootCmd.PersistentFlags().StringVar(&caCert, "ca", "", "File containing the CA certificate")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "", "File containing the client certificate")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "", "File containing the client key")
	rootCmd.PersistentFlags().BoolVar(&useTLS, "tls", false, "Wether or not use TLS authentication")
	rootCmd.PersistentFlags().BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "Wether or not verify the CA certificates")
}
