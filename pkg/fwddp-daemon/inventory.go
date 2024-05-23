// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2024 Intel Corporation

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"regexp"

	"github.com/go-logr/logr"
	ethernetv1 "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/apis/ethernet/v1"
	"github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/pkg/utils"
	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/net"
	"github.com/jaypipes/ghw/pkg/pci"
)

const (
	ethtoolPath = "ethtool"
)

var (
	ethtoolRegex = regexp.MustCompile(`^([a-z-]+?)(?:\s*:\s)(.+)$`)
	devlinkRegex = regexp.MustCompile(`^\s+([\w\.]+) (.+)$`)
	cvlKeyRegex  = regexp.MustCompile(`(?i)^e810.*$`)
	fvlKeyRegex  = regexp.MustCompile(`(?i)^[a-z]{0,3}710.*$`)
)

var getPCIDevices = func() ([]*ghw.PCIDevice, error) {
	pci, err := ghw.PCI()
	if err != nil {
		return nil, fmt.Errorf("failed to get PCI info: %v", err)
	}

	devices := pci.ListDevices()
	if len(devices) == 0 {
		return nil, fmt.Errorf("got 0 devices")
	}
	return devices, nil
}

var getNetworkInfo = func() (*net.Info, error) {
	net, err := ghw.Network()
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %v", err)
	}
	return net, nil
}

var execEthtool = func(nicName string) ([]byte, error) {
	return exec.Command(ethtoolPath, "-i", nicName).Output()
}

var execDevlink = func(pciAddr string) ([]byte, error) {
	devName := fmt.Sprintf("pci/%s", pciAddr)

	return exec.Command("devlink", "dev", "info", devName).CombinedOutput()
}

func IsCvlKey(key string) bool {
	return cvlKeyRegex.MatchString(key)
}

func IsFvlKey(key string) bool {
	return fvlKeyRegex.MatchString(key)
}

func getDeviceSupportedOptions(d *pci.Device) (deviceSupported bool, ddpSupported bool) {
	deviceSupported = false
	ddpSupported = false
	if d == nil {
		return
	}

	for key, supported := range *compatibilityMap {
		if supported.VendorID == d.Vendor.ID &&
			supported.Class == d.Class.ID &&
			supported.SubClass == d.Subclass.ID &&
			supported.DeviceID == d.Product.ID {
			deviceSupported = true

			if IsCvlKey(key) {
				ddpSupported = true
			}
			break
		}
	}
	return
}

func GetInventory(log logr.Logger) ([]ethernetv1.Device, error) {
	pciDevices, err := getPCIDevices()
	if err != nil {
		return nil, err
	}

	var devices []ethernetv1.Device

	for _, pciDevice := range pciDevices {
		deviceSupported, ddpSupported := getDeviceSupportedOptions(pciDevice)
		if deviceSupported {
			d := ethernetv1.Device{
				PCIAddress: pciDevice.Address,
				Name:       pciDevice.Product.Name,
				VendorID:   pciDevice.Vendor.ID,
				DeviceID:   pciDevice.Product.ID,
			}
			addNetInfo(log, &d)
			if ddpSupported {
				addDDPInfo(log, &d)
			} else {
				log.Info("Device does not support DDP profiling", "Device Name", d.Name)
			}
			devices = append(devices, d)
		}
	}

	return devices, nil
}

func GetDDPsFromHost(log logr.Logger) []string {
	// see .volumeMounts in assets/200-daemon.yaml as to why this path was used
	rootDirs := []string{"/host/discover-ddp"}
	log.Info("Discovering DDP packages located on host", "dirs", rootDirs)

	discoveredDDPs := discoverDDPs(rootDirs, log)
	log.Info("DDP profiles found on host", "count", len(discoveredDDPs))

	return discoveredDDPs
}

func addNetInfo(log logr.Logger, device *ethernetv1.Device) {
	log.Info("adding netInfo for supported device", "device", device)

	net, err := getNetworkInfo()
	if err != nil {
		log.Error(err, "failed to get network interfaces")
		return
	}

	nicName := ""
	for _, nic := range net.NICs {
		if nic.PCIAddress != nil && *nic.PCIAddress == device.PCIAddress {
			device.Firmware.MAC = nic.MacAddress
			nicName = nic.Name
			break
		}
	}
	if nicName == "" {
		log.Info("failed to find nicName for device", "pciAddress", device.PCIAddress)
		return // NIC not found
	}

	out, err := execEthtool(nicName)
	if err != nil {
		log.Error(err, "failed when executing", "cmd", ethtoolPath)
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		m := ethtoolRegex.FindStringSubmatch(line)
		if len(m) == 3 {
			switch m[1] {
			case "driver":
				device.Driver = m[2]
			case "version":
				device.DriverVersion = m[2]
			case "firmware-version":
				device.Firmware.Version = m[2]
			}
		}
	}

}

func discoverDDPs(directories []string, log logr.Logger) []string {
	regexpPattern := regexp.MustCompile(`^ice.*\.xz$`)

	discoveredDDPs := make([]string, 0, 10)
	for _, dir := range directories {
		err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
			if err != nil {
				// continue traversing directory in case of error
				return err
			}
			// skip if file is symlink
			if info.Type()&os.ModeSymlink != 0 {
				return nil
			}
			if regexpPattern.MatchString(info.Name()) && !info.IsDir() {
				ok, err := isXZArchive(path)
				if err != nil {
					return err
				}
				if ok {
					discoveredDDPs = append(discoveredDDPs, path)
				}
			}

			return nil
		})
		if err != nil {
			log.Error(err, "error while traversing directories during DDP discovering")
		}
	}

	return discoveredDDPs
}

// Check file header to ensure it's XZ archive
func isXZArchive(filename string) (bool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}

	return len(data) >= 6 && string(data[:6]) == "\xfd7zXZ\x00", nil
}

func addDDPInfo(log logr.Logger, device *ethernetv1.Device) {
	out, err := execDevlink(device.PCIAddress)
	if err != nil {
		log.Error(err, "failed when executing devlink", "out", string(out))
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		tokens := devlinkRegex.FindStringSubmatch(line)
		if len(tokens) != 3 {
			continue
		}

		switch tokens[1] {
		case "fw.app.name":
			device.DDP.PackageName = tokens[2]
		case "fw.app":
			device.DDP.Version = tokens[2]
		case "fw.app.bundle_id":
			device.DDP.TrackID = tokens[2]
		}
	}
}

func splitPCIAddr(pciAddr string, log logr.Logger) (string, string, string, string, error) {
	pciAddrList := strings.Split(pciAddr, ":")
	if len(pciAddrList) != 3 {
		return "", "", "", "", fmt.Errorf("pci Address %v format issue, cannot split on colon ':'", pciAddr)
	}
	busDevice := strings.Split(pciAddrList[2], ".")
	if len(busDevice) != 2 {
		return "", "", "", "", fmt.Errorf("bus and device %v format issue, cannot split on dot '.'", pciAddrList[2])
	}

	return pciAddrList[0], pciAddrList[1], busDevice[0], busDevice[1], nil
}

type DeviceIDs utils.SupportedDevice
