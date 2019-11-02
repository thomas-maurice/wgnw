package cmd

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thomas-maurice/wgnw/proto"
)

var (
	subnets int32
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manages the networks",
	Long:  ``,
}

var networkCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a network",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			logrus.Fatal("You should pass a network name and a CIDR")
		}

		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}

		data, err := c.CreateNetwork(context.Background(), &proto.CreateNetworkRequest{
			Name:    args[0],
			Address: args[1],
			Subnets: subnets,
		})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List networks",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}

		data, err := c.ListNetworks(context.Background(), &empty.Empty{})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

var networkGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets a network",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("You should only provide a network name")
		}

		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}

		data, err := c.GetNetwork(context.Background(), &proto.GetNetworkRequest{Name: args[0]})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

var networkDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a network",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("You should only provide a network name")
		}

		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}

		data, err := c.DeleteNetwork(context.Background(), &proto.DeleteNetworkRequest{Name: args[0]})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

func initNetworkCmd() {
	networkCreateCmd.PersistentFlags().Int32VarP(&subnets, "subnets", "s", 4, "Number of subnets")
	networkCmd.AddCommand(networkCreateCmd)
	networkCmd.AddCommand(networkListCmd)
	networkCmd.AddCommand(networkGetCmd)
	networkCmd.AddCommand(networkDeleteCmd)
}
