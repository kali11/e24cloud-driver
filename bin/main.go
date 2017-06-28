package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/kali11/e24cloud-driver"
)

func main() {
	plugin.RegisterDriver(new(e24cloud.Driver))
}