```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2024 Intel Corporation
```

<!-- omit in toc -->
# Release Notes

This document provides high-level system features, issues, and limitations information for Intel® Ethernet Operator.

- [Release history](#release-history)
- [Features for Release](#features-for-release)
- [Changes to Existing Features](#changes-to-existing-features)
- [Fixed Issues](#fixed-issues)
- [Known Issues and Limitations](#known-issues-and-limitations)
- [Release Content](#release-content)
- [Hardware Compatibility](#hardware-compatibility)
- [Supported Operating Systems](#supported-operating-systems)
- [Package Versions](#package-versions)

# Release history

| Version            | Release Date  | Cluster Compatibility                          | Verified on                  |
|--------------------|---------------|------------------------------------------------|------------------------------|
| 0.0.1              | January  2022 | OCP 4.9                                        | OCP 4.9.7                    |
| 0.0.2              | April    2022 | BMRA 22.01(K8S v1.22.3)<br>OCP 4.10            | OCP 4.10.3                   |
| 0.1.0-redhat-cert  | December 2022 | OCP 4.9, 4.10, 4.11                            | OCP 4.9.51, 4.10.34, 4.11.13 |
| v22.11             | December 2022 | OCP 4.9, 4.10, 4.11<br>BMRA 22.11(K8S v1.25.3) | OCP 4.9.51, 4.10.34, 4.11.13 |
| v23.07-redhat-cert | July     2023 | OCP 4.9, 4.10, 4.11, 4.12, 4.13                | OCP 4.12.21, 4.13.3          |
| v23.08             | August   2023 | K8S v1.25.3                                    | K8S v1.25.3                  |
| v23.08-operatorhub | November 2023 | K8S >v1.21                                     | K8S v1.25.3                  |
| v24.04             | April    2024 | K8S >v1.21, OCP >4.11                          | K8S v1.29.0, OCP 4.14.7      |

# Features for Release

***v24.04***

- Add firmware updates support for 700 series (FVL) NICs.
- Add operator Helm Chart.
- Load RedHat provided DDP profiles from host on OCP -> on OCP there are different DDP versions preconfigured in
  `/lib/firmware/intel/ice/` directory on each of cluster hosts, intel-ethernet-operator can now discover and load these
  profiles. This way of updating DDP is requested by the users in the same way as regular ones, via EthernetClusterConfig.
- Change EthernetClusterConfig API -> `deviceSelector` fields are now arrays of strings instead of strings. This change
  enables creation of ECCs that could target more than single device. Before, there was an option to either use single
  pciAddress to update one device or do not provide deviceSelector at all and update all the cards on the node.
- Add clarifications and updates to documentation.
- Allow retries after a FW/DDP update failure to be switched on or off through the `retryOnFail` field in the
  EthernetClusterConfig CR (this feature was available in 23.07-redhat-cert release but it wasn't in 23.08 (main branch)).

***v23.08-operatorhub***

- Add Intel Ethernet Operator to [OperatorHub.io](https://operatorhub.io/) catalog.

***v23.08***

- Allow additional parameters to be added to the nvmupdate tool through the `fwUpdateParam` field in the
  EthernetClusterConfig CR
- Add support for alternative firmware search path which is detected through query performed
  on node by manager pod
- Add mTLS for webhook server

***v23.07***

- Add support for DDP profile with in-tree
- Certification on OCP 4.12 and 4.13
- Allow retries after a FW/DDP update failure to be switched on or off through the `retryOnFail` field in the
  EthernetClusterConfig CR
- Allow additional parameters to be added to the nvmupdate tool through the `fwUpdateParam` field in the
  EthernetClusterConfig CR

***v22.11***

- Introduced new API `ClusterFlowConfig` for cluster level Flow configuration
- Improved stability for `NodeFlowConfig` daemon
- Add user provided certificate validation during FW/DDP package download
- Operator is updated with Operator SDK v1.25.0
- Set Min TLS v1.3 for validation webhook for improved security

***v0.1.0 (certified on OCP)***

- FW update supported on in-tree driver
- DDP profile update and traffic flow configuration are not supported when using in-tree
- Operator is updated with Operator SDK v1.25.0

***v0.0.2***

- Operator has been ported to Vanilla Kubernetes

***v0.0.1***

- Intel Ethernet Operator
  - The operator handles the Firmware update of Intel® Ethernet Network Adapter E810 Series.
  - The operator handles the DDP (Dynamic Device Personalization) profile update of Intel® Ethernet Network Adapter
  E810 Series.
  - The operator handles the traffic flow configuration of Intel® Ethernet Network Adapter E810 Series.

# Changes to Existing Features

***v24.04***

- Remove sriov-network-operator as dependency to intel-ethernet-operator.
- Change EthernetClusterConfig API -> `deviceSelector` fields are now arrays of strings instead of strings. This change
  enables creation of ECCs that could target more than single device. Before, there was an option to either use single
  pciAddress to update one device or do not provide deviceSelector at all and update all the cards on the node.

***v23.08-operatorhub***

None

***v23.08***

- Nvmupdate tool error codes 50 & 51 are no longer treated as update failures
- Alternative firmware search path is no longer enabled on OCP and disabled on K8S by default, query by manager
  pod is performed on node to decide variant for cluster

***v23.07***

- DDP update is now possible on in-tree driver
- Nvmupdate tool error codes 50 & 51 are no longer treated as update failures
- Reboots after a fw update no longer take place by default

***v22.11***

- Use SHA-1 instead of MD5 checksum for FW/DDP update
- Default UFT image version has been updated v22.03 -> v22.07

***v22.11***

- Use SHA-1 instead of MD5 checksum for FW/DDP update
- Default UFT image version has been updated v22.03 -> v22.07

***v0.1.0***

- FW update is now possible on in-tree driver

***v0.0.2***

- Any update of DDP packages causes node reboot
- DCF Tool has been updated v21.08 -> v22.03
- Proxy configuration for FWDDP Daemon app has been added
- Updated documentation for CRDs (EthernetClusterConfig, EthernetNodeConfig)
- Replicas of Controller Manager are now distributed accross a cluster
- EthernetClusterConfig.DrainSkip flag has been removed, IEO detects cluster type automatically and decides if drain
  is needed.

***v0.0.1***

- There are no unsupported or discontinued features relevant to this release.

# Fixed Issues

***v24.04***

- Change DeepDerivative to DeepEqual in EthernetClusterConfig creation -> DeepDerivative ignores fields that are not
  present in new ECC so when user applies empty ECC or ECC with smaller amount of fields it's treated as equal to
  older ECC. DeepEqual checks all fields of ECC spec.
- Fix logic that checks if reboot is needed -> There might be case when reboot is needed after updating first device
  from the queue, but when second device from queue is updated and reboot is not requested, reboot requested by first
  device will be overwritten. This is fixed in v24.04.

***v22.11***

- Fixed checksum verification for FW and DDP update
- FlowConfig daemon pod cleanup correctly
- Fixed an incorrect flow rules deletion issue

***v0.0.2***

- fixed DCF tool image registry URL reference issue. The DCF tool registry URL will be read from `IMAGE_REGISTRY`
  env variable during operator image build

***v0.0.1***

- n/a - this is the first release.

# Known Issues and Limitations

- On rare occasions, firmware update operation might end up in failure with error reporting `exit status 6`. This issue
  is caused by known ICE driver instability. It can occur on various versions (both in-tree and out-of-tree) of the
  driver. Reloading ice driver (or restarting node) after encountering the issue, and then requesting firmware update
  again fixes the issue.
- Operator support updates to firmware versions 3.0 or newer.
- To perform fw update to 4.2 `fwUpdateParam: -if ioctl` needs to be added to `EthernetClusterConfig` CR:

  ```yaml
  apiVersion: ethernet.intel.com/v1
  kind: EthernetClusterConfig
  metadata:
    name: <name>
    namespace: <namespace>
  spec:
    nodeSelectors:
      kubernetes.io/hostname: <hostname>
    deviceSelector:
      pciAddresses:
        - "<pci-address>"
    deviceConfig:
      fwURL: "<URL_to_firmware>"
      fwChecksum: "<file_checksum_SHA-1_hash>"
      fwUpdateParam: "-if ioctl"

  ```

- The creation of trusted VFs to be used by the Flow Configuration controller of the operator and the creation of VFs
  to be used by the applications is out of scope for this operator. The user is required to create necessary VFs.
- The installation of the out-of-tree [ICE driver](https://www.intel.com/content/www/us/en/download/19630/29746/) is
  necessary to leverage certain features of the operator (traffic flow configuration). The provisioning/installation of
  this driver is out of scope for this operator, the user is required to provide/install the
  [OOT ICE driver](https://www.intel.com/content/www/us/en/download/19630/29746/intel-network-adapter-driver-for-e810-series-devices-under-linux.html)
  on the desired platforms.
  **BMRA distribution comes with required version of ICE driver and no additional steps are required.**
  
***v23.08-operatorhub***

- OperatorHub release do not contain Flow Configuration. Flow Configuration will not be available when installing
  Intel Ethernet Operator from OperatorHub catalog.

***v.0.1***

- The certified version 0.1.0 only functionality is fw update, DDP and traffic flow configuration is not possible on
  in-tree driver

# Release Content

- Intel Ethernet Operator
- Documentation

# Hardware Compatibility

- [Intel® Ethernet Network Adapter X710-DA2/DA4](https://cdrdv2-public.intel.com/641693/Intel%20Ethernet%20Converged%20Network%20Adapter%20X710-DA2-DA4.pdf)
- [Intel® Ethernet Network Adapter E810-CQDA1/CQDA2](https://cdrdv2.intel.com/v1/dl/getContent/641676?explicitVersion=true)
- [Intel® Ethernet Network Adapter E810-XXVDA4](https://cdrdv2.intel.com/v1/dl/getContent/641676?explicitVersion=true)
- [Intel® Ethernet Network Adapter E810-XXVDA2](https://cdrdv2.intel.com/v1/dl/getContent/641674?explicitVersion=true)

# Supported Operating Systems

***v24.04***

- K8S >v1.21 (tested on K8S v1.29, Ubuntu 22.04 - 6.5.0-25-generic in-tree ICE driver and 1.9.11 OOT ICE driver)
- OCP >4.11 (tested on 4.14.7, Red Hat Enterprise Linux CoreOS 414.92.202312132152-0 - 5.14.0-284.45.1.el9_2.x86_64
  in-tree ICE driver)

***v23.08-operatorhub***

- K8S >v1.21 (tested on Ubuntu 22.04 - 5.15.0-88-generic intree ICE driver)

***v23.08*** was tested using the following

K8S >v1.21 (tested on Ubuntu 22.04 - 1.9.11 OOT ICE driver)

***v23.07*** was tested using the following

- Openshift
  - 4.12.21
  - 4.13.3

***v22.11***

- Intel BMRA v22.11 (Ubuntu 20.04 & 22.04)

***v0.1.0*** was tested using the following:

- OpenShift
  - 4.9.51
  - 4.10.34
  - 4.11.13

***v0.0.2*** was tested using the following:

- BMRA 22.01
- Kubernetes v1.22.3
- OS: Ubuntu 20.04.3 LTS (Focal Fossa)
- NVM Package:  v1.37.13.5

- OpenShift: 4.10.3
- OS: Red Hat Enterprise Linux CoreOS 410.84.202202251620-0 (Ootpa)
- Kubernetes:  v1.23.3+e419edf
- NVM Package:  v1.37.13.5

***v0.0.1*** was tested using the following:

- OpenShift: 4.9.7
- OS: Red Hat Enterprise Linux CoreOS 49.84.202111022104-0
- Kubernetes:  v1.22.2+5e38c72
- NVM Package:  v1.37.13.5

# Package Versions

***v23.08***

- Kubernetes: v1.29.0
- Golang: v1.21
- DCF Tool: v22.07

***v23.08***

- Kubernetes: v1.25.3  
- Golang: v1.20
- DCF Tool: v22.07

***v23.07***

- Kubernetes: v1.22.2+5e38c72
- Golang: v1.19
- DCF Tool: v22.07

***v22.11***

- Kubernetes: v1.22.2+5e38c72
- Golang: v1.17.3
- DCF Tool: v22.07

***v0.0.2 Packages***

- Kubernetes: v1.22.2+5e38c72
- Golang: v1.17.3
- DCF Tool: v21.11

***v0.0.1 Packages***

- Kubernetes:  v1.22.2+5e38c72|v1.22.3
- Golang: v1.17.3
- DCF Tool: v21.08
