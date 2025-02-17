// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2024 Intel Corporation

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Pattern=`^[a-fA-F0-9]{4}:[a-fA-F0-9]{2}:[01][a-fA-F0-9]\.[0-7]$`
type PciAddress string

type DeviceSelector struct {
	// +kubebuilder:validation:MinItems=1
	// VendorIds of devices to be selected. If value is not set, then cards with any VendorId are selected
	VendorIDs []string `json:"vendorIds,omitempty"`
	// DeviceIds of devices to be selected. If value is not set, then cards with any DeviceId are selected
	// +kubebuilder:validation:MinItems=1
	DeviceIDs []string `json:"deviceIds,omitempty"`
	// +kubebuilder:validation:MinItems=1
	// PciAddresses of devices to be selected. If value is not set, then cards with any PciAddress are selected
	PCIAddresses []PciAddress `json:"pciAddresses,omitempty"`
}

type DeviceConfig struct {
	// Path to .zip DDP package to be applied
	// +kubebuilder:validation:Pattern=[a-zA-Z0-9\.\-\/]+
	DDPURL string `json:"ddpURL,omitempty"`
	// SHA-1 checksum of .zip DDP package
	// +kubebuilder:validation:Pattern=`^[a-fA-F0-9]{40}$`
	DDPChecksum string `json:"ddpChecksum,omitempty"`
	// Path to .xz DDP profile package discovered on host
	// +kubebuilder:validation:Pattern=`ice.*\.xz$`
	DiscoveredDDPPath string `json:"discoveredDDPPath,omitempty"`

	// Path to .tar.gz Firmware (NVMUpdate package) to be applied
	// +kubebuilder:validation:Pattern=[a-zA-Z0-9\.\-\/]+
	FWURL string `json:"fwURL,omitempty"`
	// +kubebuilder:validation:Pattern=`^[a-fA-F0-9]{40}$`
	// SHA-1 checksum of .tar.gz Firmware
	FWChecksum string `json:"fwChecksum,omitempty"`
	// Additional arguments for NVMUpdate utility
	// e.g. "./nvmupdate64e -u -m 40a6b79ee660 -c ./nvmupdate.cfg -o update.xml -l <fwUpdateParam>"
	FWUpdateParam string `json:"fwUpdateParam,omitempty"`
}

// EthernetClusterConfigSpec defines the desired state of EthernetClusterConfig
type EthernetClusterConfigSpec struct {
	// Selector for nodes. If value is not set, then configuration is applied to all nodes with CVL cards in cluster
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	NodeSelector map[string]string `json:"nodeSelectors,omitempty"`
	// Selector for devices on nodes. If value is not set, then configuration is applied to all CVL cards on selected nodes
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	DeviceSelector DeviceSelector `json:"deviceSelector,omitempty"`
	// Contains configuration which will be applied to selected devices
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	DeviceConfig DeviceConfig `json:"deviceConfig"`

	// Higher priority policies can override lower ones.
	// If several ClusterConfigs have same Priority, then operator will apply ClusterConfig with highest CreationTimestamp (newest one)
	Priority int `json:"priority,omitempty"`

	// Set to true to retry update every 5 minutes
	// Default is set to false - no retries will occur
	RetryOnFail bool `json:"retryOnFail,omitempty"`
}

// EthernetClusterConfigStatus defines the observed state of EthernetClusterConfig
type EthernetClusterConfigStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=ecc

// EthernetClusterConfig is the Schema for the ethernetclusterconfigs API
// +operator-sdk:csv:customresourcedefinitions:resources={{DaemonSet,v1,fwddp-daemon}}
type EthernetClusterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EthernetClusterConfigSpec   `json:"spec,omitempty"`
	Status EthernetClusterConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EthernetClusterConfigList contains a list of EthernetClusterConfig
type EthernetClusterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EthernetClusterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EthernetClusterConfig{}, &EthernetClusterConfigList{})
}
