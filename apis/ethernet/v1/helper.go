// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2024 Intel Corporation

package v1

import "reflect"

func (ds DeviceSelector) Matches(d Device) bool {
	// if deviceSelector is empty, update all devices on node
	if reflect.DeepEqual(ds, DeviceSelector{}) {
		return true
	}

	match := false
	for _, vendorId := range ds.VendorIDs {
		if vendorId != "" && vendorId == d.VendorID {
			match = true
			break
		}
	}
	for _, deviceId := range ds.DeviceIDs {
		if deviceId != "" && deviceId == d.DeviceID {
			match = true
			break
		}
	}
	for _, pciAddress := range ds.PCIAddresses {
		if pciAddress != "" && pciAddress == PciAddress(d.PCIAddress) {
			match = true
			break
		}
	}

	return match
}
