package main

import (
	"context"
	"fmt"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/golang/protobuf/ptypes/empty"

	proto "github.com/thomas-maurice/wgnw/proto"
	"github.com/thomas-maurice/wgnw/server/interfaces"
)

type WireguardServer struct {
	wgService interfaces.WireguardService
}

func NewWireguardServer(wgService interfaces.WireguardService) (*WireguardServer, error) {
	return &WireguardServer{
		wgService: wgService,
	}, nil
}

func (s *WireguardServer) AcquireLease(ctx context.Context, leaseRequest *proto.AcquireLeaseRequest) (*proto.AcquireLeaseResponse, error) {
	lease, err := s.wgService.AcquireLease(leaseRequest)
	return &proto.AcquireLeaseResponse{
		Lease: lease,
	}, err
}

func (s *WireguardServer) GetLease(ctx context.Context, l *proto.GetLeaseRequest) (*proto.GetLeaseResponse, error) {
	lease, err := s.wgService.GetLease(l.Uuid)
	return &proto.GetLeaseResponse{
		Lease: lease,
	}, err
}

func (s *WireguardServer) ListLeases(ctx context.Context, nothing *empty.Empty) (*proto.ListLeasesResponse, error) {
	leases, err := s.wgService.ListLeases()
	return &proto.ListLeasesResponse{
		Leases: leases,
	}, err
}

func (s *WireguardServer) DeleteLease(ctx context.Context, l *proto.DeleteLeaseRequest) (*proto.DeleteLeaseResponse, error) {
	err := s.wgService.DeleteLease(l.Uuid)
	return &proto.DeleteLeaseResponse{
		Uuid: l.Uuid,
	}, err
}

func (s *WireguardServer) RenewLease(ctx context.Context, l *proto.RenewLeaseRequest) (*proto.RenewLeaseResponse, error) {
	lease, err := s.wgService.RenewLease(l.Uuid)
	return &proto.RenewLeaseResponse{
		Lease: lease,
	}, err
}

func (s *WireguardServer) FetchConfiguration(ctx context.Context, cfg *proto.ConfigurationRequest) (*proto.ConfigurationResponse, error) {
	c, err := s.wgService.FetchConfiguration(cfg.NetworkName)
	return c, err
}

func (s *WireguardServer) CreateNetwork(ctx context.Context, spec *proto.CreateNetworkRequest) (*proto.CreateNetworkResponse, error) {
	_, network, err := net.ParseCIDR(spec.Address)
	extraBits := len(fmt.Sprintf("%b", spec.Subnets-1))

	// If we want only one network then whatever mate
	if spec.Subnets == 1 {
		extraBits = 0
	}

	var subnets []string
	for i := int32(0); i < spec.Subnets; i++ {
		subnet, err := cidr.Subnet(network, extraBits, int(i))
		if err != nil {
			return &proto.CreateNetworkResponse{}, err
		}
		subnets = append(subnets, subnet.String())
	}
	if err != nil {
		return &proto.CreateNetworkResponse{}, err
	}

	err = s.wgService.CreateNetwork(&proto.Network{
		Name:       spec.Name,
		Address:    network.String(),
		Subnets:    subnets,
		NumSubnets: spec.Subnets,
	})

	if err != nil {
		return &proto.CreateNetworkResponse{}, err
	}

	return &proto.CreateNetworkResponse{
		Network: &proto.Network{
			Name:       spec.Name,
			Address:    network.String(),
			Subnets:    subnets,
			NumSubnets: spec.Subnets,
		},
	}, nil
}

func (s *WireguardServer) ListNetworks(ctx context.Context, nothing *empty.Empty) (*proto.ListNetworksResponse, error) {
	networks, err := s.wgService.ListNetworks()
	if err != nil {
		return &proto.ListNetworksResponse{}, err
	}

	return &proto.ListNetworksResponse{
		Networks: networks,
	}, nil
}

func (s *WireguardServer) GetNetwork(ctx context.Context, spec *proto.GetNetworkRequest) (*proto.GetNetworkResponse, error) {
	network, err := s.wgService.GetNetwork(spec.Name)
	if err != nil {
		return &proto.GetNetworkResponse{}, err
	}

	return &proto.GetNetworkResponse{
		Network: network,
	}, nil
}

func (s *WireguardServer) DeleteNetwork(ctx context.Context, spec *proto.DeleteNetworkRequest) (*proto.DeleteNetworkResponse, error) {
	err := s.wgService.DeleteNetwork(spec.Name)
	if err != nil {
		return &proto.DeleteNetworkResponse{}, err
	}

	return &proto.DeleteNetworkResponse{
		Name: spec.Name,
	}, nil
}

func (s *WireguardServer) PurgeLeases(ctx context.Context, nothing *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.wgService.PurgeLeases()
}
