//go:build linux || darwin
// +build linux darwin

package utils

import (
	"fmt"
	"net"
)

// Gets all physical interfaces based on filter results, ignoring all VM, Loopback and Tunnel interfaces.
func GetAllPhysicalInterfaces() []PhysicalInterface {
	ifaces, err := net.Interfaces()

	if err != nil {
		fmt.Println(err)
		return nil
	}

	var outInterfaces []PhysicalInterface

	for _, element := range ifaces {
		if element.Flags&net.FlagLoopback == 0 && element.Flags&net.FlagUp == 1 && isPhysicalInterface(element.HardwareAddr.String()) {
			outInterfaces = append(outInterfaces, PhysicalInterface{MACAddress: element.HardwareAddr.String(), Name: element.Name, FriendlyName: element.Name})
		}
	}

	return outInterfaces
}
