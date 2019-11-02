package interfaces

import (
	proto "github.com/thomas-maurice/wgnw/proto"
)

type WireguardService interface {
	CreateNetwork(*proto.Network) error
	ListNetworks() ([]*proto.Network, error)
	GetNetwork(string) (*proto.Network, error)
	DeleteNetwork(string) error

	AcquireLease(*proto.AcquireLeaseRequest) (*proto.Lease, error)
	ListLeases() ([]*proto.Lease, error)
	GetLease(string) (*proto.Lease, error)
	DeleteLease(string) error
	RenewLease(string) (*proto.Lease, error)
	PurgeLeases() error

	FetchConfiguration(string) (*proto.ConfigurationResponse, error)
}
