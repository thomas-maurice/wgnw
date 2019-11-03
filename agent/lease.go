package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/thomas-maurice/wgnw/proto"
)

func newLease(client proto.WireguardServiceClient,
	network string,
	pubkey string,
	publicPeer *proto.PublicPeer,
) (*proto.Lease, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	leaseRequest, err := client.AcquireLease(getContext(), &proto.AcquireLeaseRequest{
		PublicKey:   pubkey,
		NetworkName: network,
		NodeName:    hostname,
		Peer:        publicPeer,
	})
	if err != nil {
		logrus.WithError(err).Error("Could not acquire lease")
		return nil, err
	}
	return leaseRequest.Lease, nil
}

func getOrRenewLease(client proto.WireguardServiceClient,
	network string,
	pubkey string,
	publicPeer *proto.PublicPeer,
	state *State, // State will be modified
) (*proto.Lease, error) {
	if state.LeaseUUID == "" {
		// Create a new lease if we don't have any
		lease, err := newLease(client, network, pubkey, publicPeer)
		if err != nil {
			logrus.WithError(err).Error("Could not acquire lease")
			return nil, err
		}
		state.LeaseUUID = lease.Uuid
		return lease, nil
	}

	renewedLease, err := client.RenewLease(getContext(), &proto.RenewLeaseRequest{
		Uuid: state.LeaseUUID,
	})

	if err != nil {
		logrus.WithError(err).Errorf("Could not renew lease %s", state.LeaseUUID)
		return nil, err
	}

	if err != nil || renewedLease.Lease.Expired {
		// We have to get a new lease
		lease, err := newLease(client, network, pubkey, publicPeer)
		if err != nil {
			logrus.WithError(err).Error("Could not acquire lease")
			return nil, err
		}
		state.LeaseUUID = lease.Uuid
		return lease, nil
	}

	state.LeaseUUID = renewedLease.Lease.Uuid
	return renewedLease.Lease, nil
}
