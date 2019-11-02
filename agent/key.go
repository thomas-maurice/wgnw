package main

import (
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func getWireguardKey(filename string) (*wgtypes.Key, error) {
	var key wgtypes.Key
	if _, err := os.Stat(filename); err != nil && !os.IsNotExist(err) {
		logrus.WithError(err).Errorf("Could not stat key file %s", filename)
		return nil, err
	} else if os.IsNotExist(err) {
		logrus.Info("Generating a new private key...")
		key, err = wgtypes.GeneratePrivateKey()
		if err != nil {
			logrus.WithError(err).Error("Cannot generate a private key")
			return nil, err
		}
		err = ioutil.WriteFile(filename, []byte(key.String()), 0600)
		if err != nil {
			logrus.WithError(err).Errorf("Could not write private key to %s", filename)
			return nil, err
		}

		return &key, nil
	} else {
		logrus.Info("Reading private key from disk")
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			logrus.WithError(err).Errorf("Could not read private key from %s", filename)
			return nil, err
		}
		key, err = wgtypes.ParseKey(string(b))
		if err != nil {
			logrus.WithError(err).Fatalf("Could not parse private key from %s", filename)
			return nil, err
		}
		return &key, nil
	}
}
