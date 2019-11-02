package cmd

import (
	"context"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/thomas-maurice/wgnw/proto"
)

var (
	publicKey string
	address   string
	port      int32
)

var leaseCmd = &cobra.Command{
	Use:   "lease",
	Short: "Manages the leases",
	Long:  ``,
}

var leaseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a lease",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("You should pass a network name")
		}

		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}

		if publicKey == "" {
			key, err := wgtypes.GeneratePrivateKey()
			if err != nil {
				logrus.WithError(err).Fatal("Could not generate wireguard private key")
			}
			publicKey = key.PublicKey().String()
		}

		var peer *proto.PublicPeer
		if address != "" && port != 0 {
			peer = &proto.PublicPeer{
				Address: address,
				Port:    port,
			}
		}

		hostname, err := os.Hostname()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get hostname")
		}

		data, err := c.AcquireLease(context.Background(), &proto.AcquireLeaseRequest{
			NetworkName: args[0],
			NodeName:    hostname,
			PublicKey:   publicKey,
			Peer:        peer,
		})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

var leaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List leases",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}
		data, err := c.ListLeases(context.Background(), &empty.Empty{})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

var leaseGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets a lease",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("You should only provide a lease uuid")
		}
		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}
		data, err := c.GetLease(context.Background(), &proto.GetLeaseRequest{Uuid: args[0]})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

var leaseDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a lease",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("You should only provide a lease uuid")
		}
		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}
		data, err := c.DeleteLease(context.Background(), &proto.DeleteLeaseRequest{Uuid: args[0]})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

var leasePurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges old leases",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getClient()
		if err != nil {
			logrus.WithError(err).Fatal("Could not get a client")
		}
		data, err := c.PurgeLeases(context.Background(), &empty.Empty{})
		if err != nil {
			logrus.WithError(err).Fatal("Error")
		}
		output(data)
	},
}

func initLeaseCmd() {
	leaseCreateCmd.PersistentFlags().StringVarP(&address, "address", "a", "", "Address where the peer is reachable")
	leaseCreateCmd.PersistentFlags().StringVarP(&publicKey, "pubkey", "k", "", "Public key for the lease")
	leaseCreateCmd.PersistentFlags().Int32VarP(&port, "port", "p", 0, "Port where the peer is reachable")

	leaseCmd.AddCommand(leaseCreateCmd)
	leaseCmd.AddCommand(leaseListCmd)
	leaseCmd.AddCommand(leaseGetCmd)
	leaseCmd.AddCommand(leaseDeleteCmd)
	leaseCmd.AddCommand(leasePurgeCmd)
}
