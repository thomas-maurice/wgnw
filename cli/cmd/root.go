package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	marshaller        string
	controllerAddress string
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
}
