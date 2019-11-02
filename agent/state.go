package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/thomas-maurice/wgnw/proto"
)

type State struct {
	LeaseUUID string `json:"lease_uuid"`
}

func saveState(filename string, lease *proto.Lease) error {
	b, err := json.Marshal(&State{LeaseUUID: lease.Uuid})
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, b, 0600)
}

func loadState(filename string) (State, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return State{}, err
	}

	var state State
	err = json.Unmarshal(b, &state)
	return state, err
}
