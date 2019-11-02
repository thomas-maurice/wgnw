package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func output(obj interface{}) {
	switch marshaller {
	case "json":
		b, err := json.MarshalIndent(&obj, "", "  ")
		if err != nil {

			logrus.WithError(err).Fatal("Could not marshall struct")
		}
		fmt.Println(string(b))
	case "yaml":
		b, err := yaml.Marshal(&obj)
		if err != nil {

			logrus.WithError(err).Fatal("Could not marshall struct")
		}
		fmt.Println(string(b))
	default:
		fmt.Println(obj)
	}
}
