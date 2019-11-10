package main

import (
	"github.com/lorenzosaino/go-sysctl"
)

func ensureSysctl() error {
	return sysctl.Set("net.ipv4.ip_forward", "1")
}
