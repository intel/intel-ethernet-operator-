// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2024 Intel Corporation

package labeler

import (
	"context"
	"fmt"
	"os"

	daemon "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/pkg/fwddp-daemon"
	"github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/pkg/utils"
	"github.com/jaypipes/ghw"
	"golang.org/x/exp/maps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	deviceConfig = "./devices.json"
)

var getInclusterConfigFunc = rest.InClusterConfig

var getPCIDevices = func() ([]*ghw.PCIDevice, error) {
	pci, err := ghw.PCI()
	if err != nil {
		return nil, fmt.Errorf("Failed to get PCI info: %v", err)
	}

	devices := pci.ListDevices()
	if len(devices) == 0 {
		return nil, fmt.Errorf("Got 0 devices")
	}

	return devices, nil
}

var findAllSupportedDevices = func(supportedList *utils.SupportedDevices) (map[string]*ghw.PCIDevice, error) {
	if supportedList == nil {
		return nil, fmt.Errorf("config not provided")
	}

	present, err := getPCIDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get PCI devices: %v", err)
	}

	supportedDevices := make(map[string]*ghw.PCIDevice)

	for _, dev := range present {
		supported, key := isDeviceSupported(dev, supportedList)
		if supported {
			supportedDevices[key] = dev
		}
	}
	return supportedDevices, nil
}

func isDeviceSupported(dev *ghw.PCIDevice, supportedList *utils.SupportedDevices) (bool, string) {
	for key, supported := range *supportedList {
		if dev.Vendor.ID == supported.VendorID &&
			dev.Class.ID == supported.Class &&
			dev.Subclass.ID == supported.SubClass &&
			dev.Product.ID == supported.DeviceID {

			fmt.Printf("FOUND %v at %v: Vendor=%v Class=%v:%v Device=%v\n", key,
				dev.Address, dev.Vendor.ID, dev.Class.ID,
				dev.Subclass.ID, dev.Product.ID)
			return true, key
		}
	}

	return false, ""
}

func findSupportedDevice(supportedList *utils.SupportedDevices) (bool, error) {
	supportedDevices, err := findAllSupportedDevices(supportedList)
	fmt.Println("Supported devices: ", supportedDevices)

	if err != nil {
		return false, err
	}

	if len(supportedDevices) == 0 {
		return false, nil
	}

	return true, nil
}

func setNodeLabel(nodeName, label string, isDevicePresent bool) error {
	if label == "" {
		return fmt.Errorf("label is empty (check CVLLABEL or FVLLABEL env vars)")
	}
	if nodeName == "" {
		return fmt.Errorf("nodeName is empty (check the NODENAME env var)")
	}

	cfg, err := getInclusterConfigFunc()
	if err != nil {
		return fmt.Errorf("Failed to get cluster config: %v\n", err.Error())
	}
	cli, err := clientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Failed to initialize clientset: %v\n", err.Error())
	}

	node, err := cli.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get the node object: %v\n", err)
	}
	nodeLabels := node.GetLabels()
	if isDevicePresent {
		nodeLabels[label] = ""
	} else {
		delete(nodeLabels, label)
	}
	node.SetLabels(nodeLabels)
	_, err = cli.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to update the node object: %v\n", err)
	}

	return nil
}

func DeviceDiscovery() error {
	supportedDevices := new(utils.SupportedDevices)
	if err := utils.LoadSupportedDevices(deviceConfig, supportedDevices); err != nil {
		return fmt.Errorf("failed to load devices: %v", err)
	}
	if len(*supportedDevices) == 0 {
		return fmt.Errorf("no devices configured")
	}

	foundSupportedDevices, err := findAllSupportedDevices(supportedDevices)
	if err != nil {
		return fmt.Errorf("failed to find any supported device: %v", err)
	}
	cvlSupported := false
	fvlSupported := false

	for _, key := range maps.Keys(foundSupportedDevices) {
		if daemon.IsCvlKey(key) && !cvlSupported {
			cvlSupported = true
		} else if daemon.IsFvlKey(key) && !fvlSupported {
			fvlSupported = true
		}
	}

	err = setNodeLabel(os.Getenv("NODENAME"), os.Getenv("CVLLABEL"), cvlSupported)

	if err != nil {
		return err
	}

	err = setNodeLabel(os.Getenv("NODENAME"), os.Getenv("FVLLABEL"), fvlSupported)

	return err
}
