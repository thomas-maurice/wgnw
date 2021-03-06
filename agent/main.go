package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/thomas-maurice/wgnw/proto"
)

var (
	ifaceName          string
	createBridge       bool
	networkName        string
	publicIP           string
	port               int
	svcAddr            string
	stateFile          string
	keyFile            string
	authToken          string
	useTLS             bool
	insecureSkipVerify bool
	caCert             string
	certFile           string
	certKeyFile        string
)

func init() {
	flag.StringVar(&ifaceName, "iface", "wg-0", "Name of the interface")
	flag.StringVar(&networkName, "net", "", "Name of the network")
	flag.StringVar(&publicIP, "public", "", "Public IP")
	flag.IntVar(&port, "port", 6666, "Port to use")
	flag.StringVar(&svcAddr, "controller", "localhost:10000", "Address of the controller")
	flag.StringVar(&stateFile, "state", "/tmp/wgagent.state", "Statefile location")
	flag.StringVar(&keyFile, "key-file", "/tmp/wgagent.key", "Private key file location")
	flag.StringVar(&authToken, "auth-token", "", "Auth token to talk to the API")
	flag.BoolVar(&useTLS, "tls", false, "Use TLS or not")
	flag.BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "Skip CA verification")
	flag.StringVar(&caCert, "ca", "", "CA cert file")
	flag.StringVar(&certFile, "cert", "", "Cert file to use")
	flag.StringVar(&certKeyFile, "key", "", "Key file to use")
	flag.BoolVar(&createBridge, "bridge", false, "Create also a bridge")
}

func main() {
	flag.Parse()

	if networkName == "" {
		logrus.Fatal("'-net' flag is mandatory")
	}

	logrus.Infof("Network name: %s", networkName)
	logrus.Infof("Interface name: %s", ifaceName)
	logrus.Infof("State file: %s", stateFile)
	logrus.Infof("Key file: %s", keyFile)

	state, err := loadState(stateFile)
	if err != nil {
		logrus.WithError(err).Warningf("Could not load the statefile %s", stateFile)
	}

	key, err := getWireguardKey(keyFile)

	if err != nil {
		logrus.WithError(err).Fatal("Could not get private key")
	}

	logrus.Infof("Using public key: %s", key.PublicKey().String())

	c, err := getClient()
	if err != nil {
		logrus.WithError(err).Fatal("Could not get a client")
	}

	var lease *proto.Lease
	var publicInfo *proto.PublicPeer

	if publicIP != "" {
		publicInfo = &proto.PublicPeer{
			Address: publicIP,
			Port:    int32(port),
		}
	}

	err = ensureSysctl()
	if err != nil {
		logrus.WithError(err).Fatal("Could not setup sysctls")
	}

	for {
		lease, err = getOrRenewLease(c, networkName, key.PublicKey().String(), publicInfo, &state)
		if err != nil || lease == nil {
			logrus.WithError(err).Error("Could not renew lease, sleeping 10s")
			time.Sleep(time.Second * 10)
			continue
		}
		err = saveState(stateFile, lease)
		if err != nil {
			logrus.WithError(err).Fatal("Could not save the state")
		}

		err = ensureInterface(ifaceName)
		if err != nil {
			logrus.WithError(err).Fatalf("Could not ensure the interface %s", ifaceName)
		}

		if createBridge {
			err = ensureBridge(fmt.Sprintf("br-%s", ifaceName))
			if err != nil {
				logrus.WithError(err).Fatalf("Could not ensure the bridge %s", fmt.Sprintf("br-%s", ifaceName))
			}
		}

		config, err := c.FetchConfiguration(getContext(), &proto.ConfigurationRequest{NetworkName: lease.Network})
		if err != nil {
			logrus.WithError(err).Error("Could not fetch configuration, will retry in 10s")
			time.Sleep(10 * time.Second)
			continue
		}

		err = configureInterface(ifaceName, lease, config)
		if err != nil {
			logrus.WithError(err).Error("Could not configure interface, will retry in 10s")
			time.Sleep(10 * time.Second)
			continue
		}

		if createBridge {
			err = configureBridge(fmt.Sprintf("br-%s", ifaceName), lease, config)
			if err != nil {
				logrus.WithError(err).Warning("Could not configure bridge, will retry in 10s")
			}
		}

		err = configureWireguardInterface(ifaceName, *key, port, config)
		if err != nil {
			logrus.WithError(err).Error("Could not apply wireguard configuration, will retry in 10s")
			time.Sleep(10 * time.Second)
			continue
		}

		time.Sleep(time.Second * 10)
	}
}
