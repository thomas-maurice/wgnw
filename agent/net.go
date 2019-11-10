package main

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/thomas-maurice/wgnw/proto"
)

// ensureInterface makes sure the interface exists and is of the correct type.
// if not the interface will be destroyed and re-created
func ensureInterface(name string) error {
	link, _ := netlink.LinkByName(name)

	if link != nil {
		if link.Type() != "wireguard" {
			logrus.Infof("Link %s is not of type 'wireguard', recreating", name)
			err := netlink.LinkDel(link)
			if err != nil {
				logrus.WithError(err).Errorf("Could not remove interface %s", name)
				return err
			}
		}
	} else {
		logrus.Warningf("No such device %s", name)
	}

	err := netlink.LinkAdd(&wireguard{LinkAttrs: netlink.LinkAttrs{Name: name}})
	if err != nil && !os.IsExist(err) {
		logrus.WithError(err).Errorf("Could not create interface %s", name)
		return err
	}

	link, _ = netlink.LinkByName(name)
	if link == nil {
		return fmt.Errorf("Could not get a handle on %s", name)
	}
	if err := netlink.LinkSetMTU(link, 1420); err != nil {
		logrus.WithError(err).Errorf("Could not set MTU for %s", name)
		return err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		logrus.WithError(err).Errorf("Could bring interface %s up", name)
		return err
	}

	return nil
}

// ensureBridge makes sure the bridge exists and is of the correct type.
// if not the bridge will be destroyed and re-created
func ensureBridge(name string) error {
	link, _ := netlink.LinkByName(name)

	if link != nil {
		if link.Type() != "bridge" {
			logrus.Infof("Link %s is not of type 'bridge', recreating", name)
			err := netlink.LinkDel(link)
			if err != nil {
				logrus.WithError(err).Errorf("Could not remove bridge %s", name)
				return err
			}
		}
	} else {
		logrus.Warningf("No such device %s", name)
	}

	err := netlink.LinkAdd(&netlink.Bridge{LinkAttrs: netlink.LinkAttrs{Name: name}})
	if err != nil && !os.IsExist(err) {
		logrus.WithError(err).Errorf("Could not create bridge %s", name)
		return err
	}

	link, _ = netlink.LinkByName(name)
	if link == nil {
		return fmt.Errorf("Could not get a handle on %s", name)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		logrus.WithError(err).Errorf("Could bring bridge %s up", name)
		return err
	}

	return nil
}

func ensureIPAddress(name string, address *net.IPNet) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		logrus.Errorf("Could not get a handle on interface %s", name)
		return err
	}

	addrs, err := netlink.AddrList(link, syscall.AF_INET)
	if err != nil {
		logrus.Errorf("Could not get addresses on interface %s", name)
		return err
	}
	for _, addr := range addrs {
		if !addr.IP.Equal(address.IP) || addr.Mask.String() != address.Mask.String() {
			logrus.Infof("Found address %s attached to %s, we only want %s, removing", addr.IPNet.String(), name, address.String())
			err = netlink.AddrDel(link, &addr)
			if err != nil {
				logrus.WithError(err).Errorf("Could not remove address %s from %s", addr.IPNet.String(), name)
			}
		}
	}

	err = netlink.AddrReplace(link, &netlink.Addr{
		IPNet: address,
	})

	if err != nil {
		logrus.WithError(err).Fatalf("Could not set address %s for %s", address.String(), name)
	}

	return nil
}

func configureInterfaceRoute(name string, route *net.IPNet) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		logrus.Errorf("Could not get a handle on interface %s", name)
		return err
	}

	return netlink.RouteReplace(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       route,
		Scope:     netlink.SCOPE_LINK,
	})
}

func configureInterface(name string, lease *proto.Lease, config *proto.ConfigurationResponse) error {
	_, selfNetwork, err := net.ParseCIDR(lease.IpRange)
	if err != nil {
		logrus.WithError(err).Errorf("Could not parse lease address %s", lease.IpRange)
		return err
	}
	_, wgNetwork, err := net.ParseCIDR(config.Network.Address)
	if err != nil {
		logrus.WithError(err).Errorf("Could not parse wireguard network address %s", lease.IpRange)
		return err
	}

	selfNetCopy := selfNetwork
	selfNetCopy.Mask = net.IPv4Mask(255, 255, 255, 255)
	err = ensureIPAddress(name, selfNetCopy)
	if err != nil {
		logrus.WithError(err).Errorf("Could not configure interface %s with its address %s", name, selfNetwork.String())
	}

	err = configureInterfaceRoute(name, wgNetwork)
	if err != nil {
		logrus.WithError(err).Errorf("Could not configure interface %s with the wireguard route %s", name, wgNetwork.String())
	}

	return nil
}

func configureBridge(name string, lease *proto.Lease, config *proto.ConfigurationResponse) error {
	_, selfNetwork, err := net.ParseCIDR(lease.IpRange)
	if err != nil {
		logrus.WithError(err).Errorf("Could not parse lease address %s", lease.IpRange)
		return err
	}
	if ones, _ := selfNetwork.Mask.Size(); ones > 31 {
		logrus.Warningf("Cannot assign an IP address to the bridge for network %s, network is too small", selfNetwork.String())
		return nil
	}
	selfNetwork.IP[3]++

	err = ensureIPAddress(name, selfNetwork)
	if err != nil {
		logrus.WithError(err).Errorf("Could not configure bridge %s with its address %s", name, selfNetwork.String())
		return err
	}

	return nil
}

func configureWireguardInterface(name string, key wgtypes.Key, port int, config *proto.ConfigurationResponse) error {
	client, err := wgctrl.New()
	if err != nil {
		logrus.Fatal(err)
	}

	var peers []wgtypes.PeerConfig
	for _, endpoint := range config.Network.Endpoints {
		// TODO: Make it configurable
		keepaliveDuration := 5 * time.Second

		var peerIPs []net.IPNet
		peerKey, err := wgtypes.ParseKey(endpoint.PublicKey)
		var udpEndpoint *net.UDPAddr
		if endpoint.Peer != nil {
			udpEndpoint = &net.UDPAddr{
				IP:   net.ParseIP(endpoint.Peer.Address),
				Port: int(endpoint.Peer.Port),
			}
		}
		if err != nil {
			logrus.WithError(err).Warningf("Could not parse peer key %s, skipping", endpoint.PublicKey)
			continue
		}
		for _, nw := range endpoint.Networks {
			_, peerNet, err := net.ParseCIDR(nw)
			if peerNet != nil {
				peerIPs = append(peerIPs, *peerNet)
			}
			if err != nil {
				logrus.WithError(err).Warningf("Could not parse peer network %s, skipping", nw)
			}
		}
		peers = append(peers, wgtypes.PeerConfig{
			PublicKey:                   peerKey,
			PersistentKeepaliveInterval: &keepaliveDuration,
			ReplaceAllowedIPs:           true,
			AllowedIPs:                  peerIPs,
			Endpoint:                    udpEndpoint,
		})
	}

	return client.ConfigureDevice(ifaceName, wgtypes.Config{
		PrivateKey:   &key,
		ListenPort:   &port,
		ReplacePeers: true,
		Peers:        peers,
	})
}
