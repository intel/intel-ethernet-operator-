```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2024 Intel Corporation
```
<!-- omit in toc -->
# Intel Ethernet Operator Documentation

- [Overview](#overview)
- [Intel Ethernet Operator](#intel-ethernet-operator)
  - [Intel Ethernet Operator - Controller-Manager](#intel-ethernet-operator---controller-manager-x710e810-nics)
  - [Intel Ethernet Operator - Device Discovery](#intel-ethernet-operator---device-discovery-x710e810-nics)
  - [Intel Ethernet Operator - FW/DDP Daemon](#dynamic-device-personalization-ddp-functionality-e810-nics-only)
    - [Firmware Update (FW) Functionality](#firmware-update-fw-functionality-x710e810-nics)
    - [Dynamic Device Personalization (DDP) Functionality](#dynamic-device-personalization-ddp-functionality-e810-nics-only)
  - [Intel Ethernet Operator - Flow Configuration](#intel-ethernet-operator---flow-configuration-e810-nics-only)
    - [Node Flow Configuration Controller](#node-flow-configuration-controller-e810-nics-only)
    - [Unified Flow Tool](#unified-flow-tool-e810-nics-only)
  - [Prerequisites](#prerequisites-e810-nics-only)
    - [ICE driver reload after reboot](#warning-ice-driver-reload-after-reboot)
    - [Alternative firmware search path on nodes with */lib/firmware* read-only](#warning-alternative-firmware-search-path-on-nodes-with-libfirmware-read-only)
    - [ICE driver variant](#ice-driver-variant-e810-nics-only)
    - [SRIOV Network Operator](#sriov-network-operator-needed-only-for-flow-configuration-functionality-e810-nics-only)
    - [Hugepages](#hugepages-needed-only-for-flow-configuration-functionality-e810-nics-only)
- [Deploying the Operator](#deploying-the-operator-x710e810-nics)
- [Using the Operator](#using-the-operator-x710e810-nics)
  - [Webserver for disconnected environment](#webserver-for-disconnected-environment-x710e810-nics)
  - [Certificate validation](#certificate-validation-x710e810-nics)
  - [Detecting NIC devices](#detecting-nic-devices-x710e810-nics)
  - [Updating Firmware](#updating-firmware-x710e810-nics)
  - [Updating DDP](#updating-ddp-e810-nics-only)
  - [Deploying Flow Configuration Agent](#deploying-flow-configuration-agent-e810-nics-only)
    - [Creating Trusted VF using SRIOV Network Operator](#creating-trusted-vf-using-sriov-network-operator-e810-nics-only)
    - [Checking node status](#checking-node-status-e810-nics-only)
    - [Creating DCF capable SRIOV Network](#creating-dcf-capable-sriov-network-e810-nics-only)
    - [Building UFT image](#building-uft-image-e810-nics-only)
    - [Creating FlowConfig Node Agent Deployment CR](#creating-flowconfig-node-agent-deployment-cr-e810-nics-only)
    - [Verifying that FlowConfig Daemon is running on available nodes:](#verifying-that-flowconfig-daemon-is-running-on-available-nodes-e810-nics-only)
  - [Creating Flow Configuration rules with ClusterFlowConfig](#creating-flow-configuration-rules-with-clusterflowconfig-e810-nics-only)
  - [Creating Flow Configuration rules with NodeFlowConfig](#creating-flow-configuration-rules-with-nodeflowconfig-e810-nics-only)
    - [Updating a sample Node Flow Configuration rule](#updating-a-sample-node-flow-configuration-rule-e810-nics-only)
- [Uninstalling operator](#uninstalling-operator-x710e810-nics)
  - [Deployed by OLM](#deployed-by-olm-operator-lifecycle-manager-x710e810-nics)
  - [Deployed by Helm](#deployed-by-helm-x710e810-nics)
- [Hardware Validation Environment](#hardware-validation-environment-x710e810-nics)
- [Summary](#summary)

## Overview

This document provides the instructions for using the Intel Ethernet Operator on supported Kubernetes clusters (Vanilla
K8s or Red Hat's OpenShift Container Platform). This operator was developed with aid of the Operator SDK project.

## Intel Ethernet Operator

The role of the Intel Ethernet Operator is to orchestrate and manage the configuration of the capabilities exposed by
the Intel X710/E810 Series network interface cards (NICs). The operator is a state machine which will configure certain
functions of the card and then monitor the status and act autonomously based on the user interaction. The operator
design of the Intel Ethernet Operator supports the following X710/E810 series cards:

- [Intel® Ethernet Network Adapter X710-DA2/DA4](https://cdrdv2-public.intel.com/641693/Intel%20Ethernet%20Converged%20Network%20Adapter%20X710-DA2-DA4.pdf)
- [Intel® Ethernet Network Adapter E810-CQDA1/CQDA2](https://cdrdv2.intel.com/v1/dl/getContent/641671?explicitVersion=true)
- [Intel® Ethernet Network Adapter E810-XXVDA4](https://cdrdv2.intel.com/v1/dl/getContent/641676?explicitVersion=true)
- [Intel® Ethernet Network Adapter E810-XXVDA2](https://cdrdv2.intel.com/v1/dl/getContent/641674?explicitVersion=true)

The Intel Ethernet Operator provides functionality for:

- Update of the devices' FW (Firmware) via [NVM Update tool](https://www.intel.com.au/content/www/au/en/support/articles/000088453/ethernet-products.html) (X710/E810 NICs).
- Update of the devices' DDP ([Dynamic Device Personalization](https://www.intel.com/content/www/us/en/architecture-and-technology/ethernet/dynamic-device-personalization-brief.html)) profile (E810 NICs only).
- Flow configuration of traffic handling for the devices, based on supported DDP profile (E810 NICs only).

The user interacts with the operator by providing CRs (CustomResources). The operator constantly monitors the state of
the CRs to detect any changes and acts based on the changes detected. There is a separate CR to be provided for the
FW/DDP update functionality and the Flow Configuration functionality. Once the CR is applied or updated, the
operator/daemon checks if the configuration is already applied and if it is not, it applies the configuration.

> :warning: **If you want to use Flow Configuration**, dependencies (out-of-tree ICE driver, SRIOV Network Operator,
Hugepages) must be fullfilled before the deployment of this operator - these dependencies are listed in the
[prerequisites section](#prerequisites-e810-nics-only). **If you want to use Operator only for modyfing firmware and DDP**,
configuration of some of the mentioned dependencies might be skipped. All the required steps are outlined in
[deployment section](#deploying-the-operator).

![Intel Ethernet Operator Design](images/Diagram1.png)

### Intel Ethernet Operator - Controller-Manager (X710/E810 NICs)

The controller manager pod is the first pod of the operator, it is responsible for deployment of other assets, exposing
the APIs, handling of the CRs and executing the validation webhook. It contains the logic for accepting and splitting
the FW/DDP CRs into node CRs and reconciling the status of each CR.

The validation webhook of the controller manager is responsible for checking each CR for invalid arguments.

### Intel Ethernet Operator - Device Discovery (X710/E810 NICs)

The CVL-discovery pod is a DaemonSet deployed on each node in the cluster. It's responsibility is to check if a
supported hardware is discovered on the platform and label the node accordingly.

To get all the nodes containing the supported devices run:

```shell
$ kubectl get EthernetNodeConfig -A

NAMESPACE                 NAME       UPDATE
intel-ethernet-operator   worker-1   InProgress
intel-ethernet-operator   worker-2   InProgress
```

To get the list of supported devices to be found by the discovery pod run:

```shell
$ kubectl describe configmap supported-cvl-devices -n intel-ethernet-operator
```

### Intel Ethernet Operator - FW/DDP Daemon (FW X710/E810 NICs, DDP E810 NICs only)

The FW/DDP daemon pod is a DaemonSet deployed as part of the operator. It is deployed on each node labeled with
appropriate label indicating that a supported X710/E810 Series NIC is detected on the platform. It is a reconcile loop
which monitors the changes in each node's `EthernetNodeConfig` and acts on the changes. The logic implemented into
this Daemon takes care of updating the NICs firmware and DDP profile. It is also responsible for draining the nodes,
taking them out of commission and rebooting when required by the update.

#### Firmware Update (FW) Functionality (X710/E810 NICs)

Once the operator/daemon detects a change to a CR related to the update of the Intel® X710/E810 NIC firmware, it tries to
perform an update. The firmware for the Intel® X710/E810 NICs is expected to be provided by the user in form of a `tar.gz`
file. The user is also responsible to verify that the firmware version is compatible with the device. The user is
required to place the firmware on an accessible HTTP server and provide an URL for it in the CR. If the file is provided
correctly and the firmware is to be updated, the Ethernet Configuration Daemon will update the Intel® X710/E810 NICs with the
NVM utility provided.

To update the NVM firmware of the Intel® X710/E810 NICs user must create a CR containing the information about which card
should be programmed. The Physical Functions of the NICs will be updated in logical pairs. The user needs to provide
the FW URL and checksum (SHA-1) in the CR.

For a sample CR go to [Updating Firmware](#updating-firmware-x710e810-nics).

#### Dynamic Device Personalization (DDP) Functionality (E810 NICs only)

Once the operator/daemon detects a change to a CR related to the update of the Intel® E810 DDP profile, it tries to
perform an update. The DDP profile for the Intel® E810 NICs is expected to be provided by the user. The user is also
responsible to verify that the DDP version is compatible with the device. The user is required to place the DDP package
on an accessible HTTP server and provide an URL for it in the CR. If the file is provided correctly and the DDP is to
be updated, the Ethernet Configuration Daemon will update the DDP profile of Intel® E810 NICs by placing it in correct
filesystem on the host.

To update the DDP profile of the Intel® E810 NIC user must create a CR containing the information about which card
should be programmed. All the Physical Functions of the NICs will be updated for each NIC.

For a sample CR go to [Updating DDP](#updating-ddp).

### Intel Ethernet Operator - Flow Configuration (E810 NICs only)

The Flow Configuration pod is a DaemonSet deployed with a CRD `FlowConfigNodeAgentDeployment` provided by Ethernet
operator once it is up and running and the required DCF VF pools and their network attachment definitions are
created with SRIOV Network Operator APIs. It is deployed on each node that exposes DCF VF pool as extended node
resource. It is a reconcile loop which monitors the changes in each node's CR and acts on the changes. The logic
implemented into this Daemon takes care of updating the NIC's traffic flow configuration. It consists of two
components Flow Config controller container and UFT container.

#### Node Flow Configuration Controller (E810 NICs only)

The Node Flow Configuration Controller watches for flow rules changes via a node specific CRD - `NodeFlowConfig` named
same as the node name. Once the operator/daemon detects a change to this CR related to the Intel® E810 Flow
Configuration, it tries to create/delete rules via UFT over an internal gPRC API call.

#### Unified Flow Tool (E810 NICs only)

Once the Flow Config change is required the Flow Config Controller will communicate with the UFT container running a
DPDK DCF application. The UFT application accepts an input with the configuration and programmes the device using a
trusted VF created for this device (it is responsibility of the user to provide the trusted VF as an allocatable K8s
resource - see [prerequisites](#prerequisites-e810-nics-only) section).

### Prerequisites (E810 NICs only)

#### :warning: ICE driver reload after reboot

Take note that for DDP profile update to take effect, ICE driver needs to be reloaded after reboot. Reboot is performed
by operator after updating DDP profile to one requested in `EthernetClusterConfig`, but reloading of ICE driver is
responsibility of user.

Such reload can be achieved by creating systemd service that executes reload
[script](../extras/ice-driver-reload.sh) on boot.

```shell
[Unit]
Description=ice driver reload
# Start after the network is up
Wants=network-online.target
After=network-online.target
# Also after docker.service (no effect on systems without docker)
After=docker.service
# Before kubelet.service (no effect on systems without kubernetes)
Before=kubelet.service
[Service]
Type=oneshot
TimeoutStartSec=25m
RemainAfterExit=true
ExecStart=/usr/bin/sh <path to reload script>
StandardOutput=journal+console
[Install]
WantedBy=default.target
```

If working on OCP, a sample [MachineConfig](../extras/ice-driver-reload-machine-config.yaml) can be used.

#### :warning: Alternative firmware search path on nodes with */lib/firmware* read-only

Some orchestration platforms based on kubernetes have */lib/firmware* directory immutable. Intel Ethernet Operator needs
read and write permissions in this directory to perform FW and DDP updates. Solution to this problem is
[alternative firmware search path](https://docs.kernel.org/driver-api/firmware/fw_search_path.html#:~:text=There%20is%20an%20alternative%20to,module%2Ffirmware_class%2Fparameters%2Fpath).
Custom firmware path needs be set up on all nodes in the cluster. Controller manager pod checks content of
`/sys/module/firmware_class/parameters/path` on node, on which it was deployed and takes that path into consideration
while managing rest of the operator resources.

One of the orchestration platform with */lib/firmware* directory immutable is RedHat OpenShift Container Platform. To
enable FW and DDP configuration on OCP, `MachineConfig` that sets `firmware_class.path` kernel argument to a directory
with `read` and `write` permisisons must be used. User might make use of example
[MachineConfig](../extras/fw-search-path-machine-config.yaml) provided in this repository.

#### ICE driver variant (E810 NICs only)

**In order for the Flow Configuration to be possible**, the platform needs to provide an OOT ICE driver v1.9.11. This is
required since current implementations of in-tree drivers do not support all required features. It is a responsibility
of the cluster admin to provide and install this driver and it is out of scope of this Operator at this time. This
particular version of the driver is required, as Intel Ethernet Operator currently only supports DPDK v22.07. Kernel
driver, DDP and firmware matching list can be found [here](https://doc.dpdk.org/guides/nics/ice.html).

**Firmware and DDP configuration** is possible with most of the in-tree drivers included in popular Linux distributions.
List of supported distributions is described in
[Feature Support Matrix](https://cdrdv2-public.intel.com/630155/630155_E810%20Feature%20Summary_Rev4_1.pdf)
(Chapter 2.0 - Operating Systems Supported).

#### SRIOV Network Operator (needed only for Flow Configuration functionality, E810 NICs only)

In order for the Flow Configuration feature to be able to configure the flow configuration of the NICs traffic the
configuration must happen using a trusted Virtual Function (VF) from each Physical Function (PF) in the NIC. Usually
it is the VF0 of a PF that has the trust mode set to `on` and bound to `vfio-pci` driver. This VF pool needs to be
created by the user and be allocatable as a K8s resource. This VF pool will be used exclusively by the UFT container
and no application container.

For user applications additional VF pools should be created separately as needed.

One way of creating and providing this trusted VF and application VFs is to configure it through
**SRIOV Network Operator**. The configuration and creation of the trusted VF and application VFs is out of scope of
Intel Ethernet Operator and is user responsibility.

#### IOMMU (needed only for Flow Configuration functionality, E810 NICs only)

In order to use SRIOV functionality of E810 series cards be sure that IOMMU is enabled on nodes. Enabling it usually
consists of making changes in UEFI/BIOS.

#### Hugepages (needed only for Flow Configuration functionality, E810 NICs only)

In order for the Flow Configuration to work, Hugepages needs to be configured on node on which flow rules will be
applied. This is standard requirement for any application that uses DPDK. Responsibilty for provisioning Hugepages lies
on the user. For more information, see [DPDK system requirements](https://doc.dpdk.org/guides/linux_gsg/sys_reqs.html)
(2.3.2. Use of Hugepages in the Linux Environment) and
[Hugepages in Kubernetes](https://kubernetes.io/docs/tasks/manage-hugepages/scheduling-hugepages/).

## Deploying the Operator (X710/E810 NICs)

There are multiple ways to deploy Intel Ethernet Operator.

- Build Operator images from source code and then use [Operator Lifecycle Manager](https://olm.operatorframework.io/)
  (OLM) or [Helm](https://helm.sh/) to deploy Operator into cluster.
- If working on OCP,
  [RedHat Certified release](https://catalog.redhat.com/software/container-stacks/detail/62f23e4cca08fe3e0ca92a9c#deploy-instructions)
  might be used. Be informed, that certified variant follows different release cycle and
  **Flow Configuration is not supported**. Refer to [release notes](./intel-ethernet-operator-release-notes.md) for
  more information.
- If working on standard Kubernetes cluster, [OperatorHub release](https://operatorhub.io/operator/intel-ethernet-operator)
  might be used. Be informed that OperatorHub variant also follows different release cycle and
  **Flow Configuration is not supported**. Refer to [release notes](./intel-ethernet-operator-release-notes.md) for
  more information.
  
> :warning: No matter which deployment variant is selected by the user, [prerequisites](#prerequisites-e810-nics-only) adequate to
  expected functionality must be fullfilled. Based on selected variant, please follow one of the deployment steps from
  list below.
  
- [Deploy from source code on OCP](deployment/source-code-ocp.md)
- [Deploy from source code on K8s](deployment/source-code-k8s.md)
- [Deploy from Certified Operators Catalog on OCP](deployment/certified-ocp.md)
- [Deploy from OperatorHub Catalog on K8s](deployment/operatorhub-k8s.md)

## Using the Operator (X710/E810 NICs)

Once the operator is successfully deployed, the user interacts with it by creating CRs which will be interpreted by the
operator.

Note: Example code below uses `kubectl` and the client binary. You can substitute `kubectl` with `oc` if you are
operating in a OCP cluster.

### Webserver for disconnected environment (X710/E810 NICs)

If cluster is running in disconnected environment, then user has to create local cache (e.g webserver) which will serve
required files. Cache should be created on machine with access to Internet. Start by creating dedicated folder for
webserver.

```shell
$ mkdir webserver
$ cd webserver
```

Create nginx Dockerfile.

```shell
$ echo "
FROM nginx
COPY files /usr/share/nginx/html
" >> Dockerfile
```

Create `files` folder.

```shell
$ mkdir files
$ cd files
```

Download required packages into `files` directory.

```shell
$ curl -OjL https://downloadmirror.intel.com/769278/E810_NVMUpdatePackage_v4_20_Linux.tar.gz
```

Build image with packages.

```shell
$ cd ..
$ podman build -t webserver:1.0.0 .
```

Push image to registry that is available in disconnected environment (or copy binary image to machine via USB flash
driver by using `podman save` and `podman load` commands).

```shell
$ podman push localhost/webserver:1.0.0 $IMAGE_REGISTRY/webserver:1.0.0
```

Create Deployment on cluster that will expose packages.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ice-cache
  namespace: default
spec:
  selector:
    matchLabels:
      run: ice-cache
  replicas: 1
  template:
    metadata:
      labels:
        run: ice-cache
    spec:
      containers:
        - name: ice-cache
          image: $IMAGE_REGISTRY/webserver:1.0.0
          ports:
            - containerPort: 80
```

And Service to make it accessible within cluster.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ice-cache
  namespace: default
  labels:
    run: ice-cache
spec:
  ports:
  - port: 80
    protocol: TCP
  selector:
    run: ice-cache
```

After that package will be available in cluster under following path:
(<http://ice-cache.default.svc.cluster.local/E810_NVMUpdatePackage_v4_20_Linux.tar.gz>)

### Certificate validation (X710/E810 NICs)

To update FW on FVL/CVL(X710/E810) card or DDP on CVL(E810) card, you have to download corresponding packages for them. For security reasons, you might
want to validate certificate that is exposed by server before downloading and this optional step describes how to add
trusted certificate.

If operator is already installed, you can apply new configuration by manually restarting `fwddp-daemon` pods.

Prepare trusted X509 certificate `certificate.der` that will be added to trusted store. It could be either root CA
certificate or intermediate certificate. It must contain `Subject Alternative Name` or `IPAdresses` identical to path
from which packages will be downloaded.

```shell
$ kubectl create -n intel-ethernet-operator secret generic tls-cert --from-file=tls.crt=certificate.der 
```

Restart `fwddp-daemon` pods or install/reinstall operator.

Check that certificate is correctly loaded in `fwddp-daemon` pods.

```shell
$ kubectl logs -n intel-ethernet-operator pod/fwddp-daemon|grep -i "found certificate - using HTTPS client"
{"level":"info","logger":"daemon","msg": "found certificate - using HTTPS client"}
```

### Detecting NIC devices (X710/E810 NICs)

To find the NIC devices belonging to the Intel® X710/E810 NIC, run following command. You can detect devices information of
the NICs from the output:

```shell
$ kubectl get enc -n intel-ethernet-operator <nodename> -o jsonpath={.status}  | jq
```

### Updating Firmware (X710/E810 NICs)

To update the Firmware of the supported device run following steps:

#### Note that

- If `deviceSelector` is left empty, the `EthernetClusterConfig` will target all compatible devices on node specified in
  `nodeSelector`.
- If `nodeSelector` is left empty, the `EthernetClusterConfig` will target devices with PCI address specified in
  `deviceSelector` on all available nodes.
- If `nodeSelector` and `deviceSelector` are both left empty, the `EthernetClusterConfig` will target all compatible
  devices on all available nodes.
- `retryOnFail` field defaults to false. If you want update to retry 5 minutes after it encounters a failure, please
  set it to true.
- `fwUpdateParam` field is optional and can be omitted in CR if not used.

> :warning: If X710/E810 NIC that you are targetting for an update has more than one device, note that Firmware for X710/E810 cards and DDP
  version for E810 cards is NIC wide. This means that if you set `deviceSelector` to PCI address of one of the NIC's device,
  Firmware and/or DDP version will be applied to all devices on that NIC.

Create a CR `yaml` file:

```yaml
apiVersion: ethernet.intel.com/v1
kind: EthernetClusterConfig
metadata:
  name: config
  namespace: <namespace>
spec:
  retryOnFail: false
  nodeSelectors:
    kubernetes.io/hostname: <hostname>
  deviceSelector:                 # Use cases:  
    deviceIds:                    #   If you want to target specific cards on your node, use pciAddresses list.
      - "<deviceID>"              #   If you want to target all E810 devices on your node, you might want to use deviceIDs list.
      - "<deviceID>"              # Also check list of notes above to create ECC that is matching your needs.
    pciAddresses:           
      - "<pci-address>"
      - "<pci-address>"
  deviceConfig:
    fwURL: "<URL_to_firmware>"
    fwChecksum: "<file_checksum_SHA-1_hash>"
    fwUpdateParam: "<optional_param>"
```

If `fwUrl` points to external location, then depending on the environment, proxy configuration might be required. Steps
describing how to configure proxy in Operator are present in [deployment guides](#deploying-the-operator).

The CR can be applied by running:

```shell
$ kubectl apply -f <filename>
```

The firmware update status can be checked by running:

```shell
$ kubectl get -n intel-ethernet-operator enc <nodename> -o jsonpath={.status.conditions}  | jq
[
  {
    "lastTransitionTime": "2021-12-17T15:25:32Z",
    "message": "Updated successfully",
    "observedGeneration": 3,
    "reason": "Succeeded",
    "status": "True",
    "type": "Updated"
  }
]
```

The user can observe the change of the NICs firmware:

```shell
$ kubectl get -n intel-ethernet-operator enc <nodename> -o jsonpath={.status.devices[<index _of_device>].firmware}  | jq
{
  "MAC": "40:a6:b7:67:1f:c0",
  "version": "3.00 0x80008271 1.2992.0"
}
```

### Updating DDP (E810 NICs only)

To update the DDP profile of the supported device run following steps:

#### Note that

- If `deviceSelector` is left empty, the `EthernetClusterConfig` will target all compatible devices on node specified in
  `nodeSelector`.
- If `nodeSelector` is left empty, the `EthernetClusterConfig` will target devices with PCI address specified in
  `deviceSelector` on all available nodes.
- If `nodeSelector` and `deviceSelector` are both left empty, the `EthernetClusterConfig` will target all compatible
  devices on all available nodes.
- `retryOnFail` field defaults to false. If you want update to retry 5 minutes after it encounters a failure, please
  set it to true.

> :warning: If X710/E810 NIC that you are targetting for an update has more than one device, note that Firmware for X710/E810 cards and DDP
  version for E810 cards is NIC wide. This means that if you set `deviceSelector` to PCI address of one of the NIC's device,
  Firmware and/or DDP version will be applied to all devices on that NIC.


Create a CR `yaml` file:

```yaml
apiVersion: ethernet.intel.com/v1
kind: EthernetClusterConfig
metadata:
  name: <name>
  namespace: <namespace>
spec:
  retryOnFail: false
  nodeSelectors:
    kubernetes.io/hostname: <hostname>
  deviceSelector:                 # Use cases:  
    deviceIds:                    #   If you want to target specific cards on your node, use pciAddresses list.
      - "<deviceID>"              #   If you want to target all E810 devices on your node, you might want to use deviceIDs list.
      - "<deviceID>"              # Also check list of notes above to create ECC that is matching your needs.
    pciAddresses:           
      - "<pci-address>"
      - "<pci-address>"
  deviceConfig:
    ddpURL: "<URL_to_DDP>"
    ddpChecksum: "<file_checksum_SHA-1_hash>"
```

If `ddpUrl` points to external location, then depending on the environment, proxy configuration might be required. Steps
describing how to configure proxy in Operator are present in [deployment guides](#deploying-the-operator).

The CR can be applied by running:

```shell
$ kubectl apply -f <filename>
```

Once the DDP profile update is complete, the following status is reported:

```shell
$ kubectl get -n intel-ethernet-operator enc <nodename> -o jsonpath={.status.conditions}  | jq
[
  {
  "lastTransitionTime": "2021-12-17T15:25:32Z",
  "message": "Updated successfully",
  "observedGeneration": 3,
  "reason": "Succeeded",
  "status": "True",
  "type": "Updated"
  }
]
```

The user can observe the change of the NICs DDP:

```shell
$ kubectl get -n intel-ethernet-operator enc <nodename> -o jsonpath={.status.devices[<index _of_device>].DDP} | jq
{
  "packageName": "ICE COMMS Package",
  "trackId": "0xc0000002",
  "version": "1.3.30.0"
}
```

#### Updating DDP using profiles preconfigured on node filesystem (newer OCP versions, E810 NICs only)

If using recent versions of OCP, there is an option to switch DDP profile to one already located on node. These are
provided and verified by Red Hat. In order to do so, first check if profiles were discovered properly.

```shell
$ oc get enc -n intel-ethernet-operator enc <nodename> -o jsonpath={.status.discoveredDDPs} | jq
[
  "/host/discover-ddp/ddp/ice-1.3.30.0.pkg.xz",
  "/host/discover-ddp/ddp-comms/ice_comms-1.3.31.0.pkg.xz",
  "/host/discover-ddp/ddp-wireless_edge/ice_wireless_edge-1.3.7.0.pkg.xz"
]
```

Create `EthernetClusterConfig` CR file and provide one of selected paths:

```yaml
apiVersion: ethernet.intel.com/v1
kind: EthernetClusterConfig
metadata:
  name: <name>
  namespace: <namespace>
spec:
  nodeSelectors:
    kubernetes.io/hostname: <hostname>
  deviceSelector:                 # Use cases:  
    deviceIds:                    #   If you want to target specific cards on your node, use pciAddresses list.
      - "<deviceID>"              #   If you want to target all E810 devices on your node, you might want to use deviceIDs list.
      - "<deviceID>"              # Also check list of notes above to create ECC that is matching your needs.
    pciAddresses:           
      - "<pci-address>"
      - "<pci-address>"
  deviceConfig:
    discoveredDDPPath: "/host/discover-ddp/ddp-comms/ice_comms-1.3.31.0.pkg.xz"
```

```shell
$ kubectl apply -f <filename>
```

Once the DDP profile update is complete, the following status is reported:

```shell
$ kubectl get -n intel-ethernet-operator enc <nodename> -o jsonpath={.status.conditions}  | jq
[
  {
  "lastTransitionTime": "2024-01-31T11:49:27Z",
  "message": "Updated successfully",
  "observedGeneration": 2,
  "reason": "Succeeded",
  "status": "True",
  "type": "Updated"
  }
]
```

The user can observe the change of the NICs DDP:

```shell
$ kubectl get -n intel-ethernet-operator enc <nodename> -o jsonpath={.status.devices[<index _of_device>].DDP} | jq
{
  "packageName": "ICE COMMS Package",
  "trackId": "0xc0000002",
  "version": "1.3.31.0"
}
```

### Updating Firmware and DDP at once (E810 NICs only)

You can also update Firmware and DDP using one `EthernetClusterConfig` CR. [Rules](#note-that) are the same as for
Firmware and DDP updates.

Create a CR `yaml` file:

```yaml
apiVersion: ethernet.intel.com/v1
kind: EthernetClusterConfig
metadata:
  name: <name>
  namespace: <namespace>
spec:
  nodeSelectors:
    kubernetes.io/hostname: <hostname>
  deviceSelector:                 # Use cases:  
    deviceIds:                    #   If you want to target specific cards on your node, use pciAddresses list.
      - "<deviceID>"              #   If you want to target all E810 devices on your node, you might want to use deviceIDs list.
      - "<deviceID>"              # Also check list of notes above to create ECC that is matching your needs.
    pciAddresses:           
      - "<pci-address>"
      - "<pci-address>"
  deviceConfig:
    ddpURL: "<URL_to_DDP>"
    ddpChecksum: "<file_checksum_SHA-1_hash>"
    fwURL: "<URL_to_firmware>"
    fwChecksum: "<file_checksum_SHA-1_hash>"
```

If `fwURL` or `ddpUrl` points to external location, then depending on the environment, proxy configuration might be
required. Steps describing how to configure proxy in Operator are present in
[deployment guides](#deploying-the-operator).

The CR can be applied by running:

```shell
$ kubectl apply -f <filename>
```

Once the Firmware version and DDP profile update is complete, the following status is reported:

```shell
$ kubectl get -n intel-ethernet-operator enc <nodename> -o jsonpath={.status.conditions}  | jq
[
  {
    "lastTransitionTime": "2024-12-12T12:36:36Z",
    "message": "Updated successfully",
    "observedGeneration": 3,
    "reason": "Succeeded",
    "status": "True",
    "type": "Updated"
  }
]
```

The user can observe the change of the NICs status:

```shell
`$ kubectl get -n intel-ethernet-operator enc <nodename> -o jsonpath='{.status.devices[<index _of_device>]}'  | jq
{
  "DDP": {
    "packageName": "ICE COMMS Package",
    "trackId": "0xc0000002",
    "version": "1.3.37.0"
  },
  ...
  "firmware": {
    "MAC": "b4:96:91:af:6e:14",
    "version": "4.01 0x800135ed 1.3256.0"
  },
  ...
}
```

### Deploying Flow Configuration Agent (E810 NICs only)

The Flow Configuration Agent Pod runs Unified Flow Tool (UFT) to configure Flow rules for a PF. UFT requires that trust
mode is enabled for the first VF (VF0) of a PF so that it has the capability of creating/modifying flow rules for that
PF. This VF also needs to be bound to `vfio-pci` driver. The SRIOV VFs pools are K8s extended resources that are exposed
via **SRIOV Network Operator**.

The VF pool consists of VF0 from all available Intel E810 series NICs PF which, in this context, we call the
**Admin VF pool**. The **Admin VF pool** is associated with a NetworkAttachmentDefinition that enables these VFs trust
mode 'on'. The SRIOV Network Operator can be used to create the **Admin VF pool** and the
**NetworkAttachmentDefinition** needed by UFT. You can find more information on creating VFs pools with
**SRIOV Network Operator**
[here](https://docs.openshift.com/container-platform/4.14/networking/hardware_networks/configuring-sriov-device.html)
and creating **NetworkAttachmentDefinition**
[here](https://docs.openshift.com/container-platform/4.14/networking/hardware_networks/configuring-sriov-net-attach.html).
This documentation is valid for both OCP and K8s versions of **SRIOV Network Operator**.

The following steps will guide you through how to create the **Admin VF pool** and the **NetworkAttachmentDefinition**
needed for Flow Configuration Agent Pod.

#### Creating Trusted VF using SRIOV Network Operator (E810 NICs only)

Once SRIOV Network operator is up and running we can examine the `SriovNetworkNodeStates` to view available Intel E810
Series NICs as shown below:

```shell
$ kubectl get sriovnetworknodestates -n intel-ethernet-operator
NAME              AGE
worker-01   1d
```

```text
$ kubectl describe sriovnetworknodestates worker-01 -n intel-ethernet-operator
Name:         worker-01
Namespace:    intel-ethernet-operator
Labels:       <none>
Annotations:  <none>
API Version:  sriovnetwork.openshift.io/v1
Kind:         SriovNetworkNodeState
Metadata:
Spec:
  Dp Config Version:  42872603
Status:
  Interfaces:
    Device ID:      165f
    Driver:         tg3
    Link Speed:     100 Mb/s
    Link Type:      ETH
    Mac:            b0:7b:25:de:3f:be
    Mtu:            1500
    Name:           eno8303
    Pci Address:    0000:04:00.0
    Vendor:         14e4
    Device ID:      165f
    Driver:         tg3
    Link Speed:     -1 Mb/s
    Link Type:      ETH
    Mac:            b0:7b:25:de:3f:bf
    Mtu:            1500
    Name:           eno8403
    Pci Address:    0000:04:00.1
    Vendor:         14e4
    Device ID:      159b
    Driver:         ice
    Link Speed:     -1 Mb/s
    Link Type:      ETH
    Mac:            b4:96:91:cd:de:38
    Mtu:            1500
    Name:           eno12399
    Pci Address:    0000:31:00.0
    Vendor:         8086
    Device ID:      159b
    Driver:         ice
    Link Speed:     -1 Mb/s
    Link Type:      ETH
    Mac:            b4:96:91:cd:de:39
    Mtu:            1500
    Name:           eno12409
    Pci Address:    0000:31:00.1
    Vendor:         8086
    Device ID:      1592
    Driver:         ice
    E Switch Mode:  legacy
    Link Speed:     -1 Mb/s
    Link Type:      ETH
    Mac:            b4:96:91:aa:d8:40
    Mtu:            1500
    Name:           ens1f0
    Pci Address:    0000:18:00.0
    Totalvfs:       128
    Vendor:         8086
    Device ID:      1592
    Driver:         ice
    E Switch Mode:  legacy
    Link Speed:     -1 Mb/s
    Link Type:      ETH
    Mac:            b4:96:91:aa:d8:41
    Mtu:            1500
    Name:           ens1f1
    Pci Address:    0000:18:00.1
    Totalvfs:       128
    Vendor:         8086
  Sync Status:      Succeeded
Events:             <none>

```

By looking at the `SriovNetworkNodeStates` status we can find the NICs information such as PCI addresses and Interface
names. Those will be used to define `SriovNetworkNodePolicy` required VFs pools.

For example, the following three `SriovNetworkNodePolicy` CRs will create a trusted VF pool name with resourceName
`cvl_uft_admin` along with two additional VFs pools for application.

> Please note, that the `SriovNetworkNodePolicy` named `uft-admin-policy` below uses `pfNames:` with VF index range
selectors to target VF0 only of Intel E810 series NIC. More information on using VF partitioning can be found
[here](https://docs.openshift.com/container-platform/4.14/networking/hardware_networks/configuring-sriov-device.html).
>
> Also note, that the `nodeSelector` label can be modified to match your preferences.
`ethernet.intel.com/intel-ethernet-cvl-present` label is automatically created by Operator on nodes with E810 NICs present.
You can also use labels created by solutions like
[Node Feature Discovery](https://kubernetes-sigs.github.io/node-feature-discovery/stable/get-started/index.html).

Save the yaml shown below to a file named `sriov-network-policy.yaml` and then apply to create the VFs pools.

```yaml
apiVersion: sriovnetwork.openshift.io/v1
kind: SriovNetworkNodePolicy
metadata:
  name: uft-admin-policy
  namespace: sriov-network-operator
spec:
  deviceType: vfio-pci
  nicSelector:
    pfNames:
    - ens1f0#0-0
    - ens1f1#0-0
    vendor: "8086"
  nodeSelector:
    ethernet.intel.com/intel-ethernet-cvl-present: ""
  numVfs: 8
  priority: 99
  resourceName: cvl_uft_admin
---
apiVersion: sriovnetwork.openshift.io/v1
kind: SriovNetworkNodePolicy
metadata:
  name: cvl-vfio-policy
  namespace: sriov-network-operator
spec:
  deviceType: vfio-pci
  nicSelector:
    pfNames:
    - ens1f0#1-3
    - ens1f1#1-3
    vendor: "8086"
  nodeSelector:
    ethernet.intel.com/intel-ethernet-cvl-present: ""
  numVfs: 8
  priority: 89
  resourceName: cvl_vfio
---
apiVersion: sriovnetwork.openshift.io/v1
kind: SriovNetworkNodePolicy
metadata:
  name: cvl-iavf-policy
  namespace: sriov-network-operator
spec:
  deviceType: netdevice
  nicSelector:
    pfNames:
    - ens1f0#4-7
    - ens1f1#4-7
    vendor: "8086"
  nodeSelector:
    ethernet.intel.com/intel-ethernet-cvl-present: ""
  numVfs: 8
  priority: 79
  resourceName: cvl_iavf

```

```shell
$ kubectl create -f sriov-network-policy.yaml
```

#### Checking node status (E810 NICs only)

Check the node status to confirm that `cvl_uft_admin` resource pool registered DCF capable VFs on the node.

```shell
$ kubectl describe node worker-01 -n intel-ethernet-operator | grep -i allocatable -A 20
Allocatable:
  bridge.network.kubevirt.io/cni-podman0:  1k
  cpu:                                     108
  devices.kubevirt.io/kvm:                 1k
  devices.kubevirt.io/tun:                 1k
  devices.kubevirt.io/vhost-net:           1k
  ephemeral-storage:                       468315972Ki
  hugepages-1Gi:                           0
  hugepages-2Mi:                           8Gi
  memory:                                  518146752Ki
  openshift.io/cvl_iavf:                   8
  openshift.io/cvl_uft_admin:              2
  openshift.io/cvl_vfio:                   6
  pods:                                    250
```

#### Creating DCF capable SRIOV Network (E810 NICs only)

Next, we will need to create SRIOV network attachment definition for the DCF VF pool as shown below:

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: sriovnetwork.openshift.io/v1
kind: SriovNetwork
metadata:
  name: sriov-cvl-dcf
  namespace: sriov-network-operator
spec:
  trust: 'on'
  networkNamespace: intel-ethernet-operator
  resourceName: cvl_uft_admin
EOF
```

>Note: If the above does not successfully set trust mode to 'on' for VF0, you can do it manually using this command:

```shell
$ ip link set <PF_NAME> vf 0 trust on
```

#### Building UFT image (E810 NICs only)

```shell
$ export IMAGE_REGISTRY=<OCP Image registry>
$ git clone https://github.com/intel/UFT.git
$ git checkout v22.07
$ make dcf-image
$ docker tag uft:v22.07 $IMAGE_REGISTRY/uft:v22.07
$ docker push $IMAGE_REGISTRY/uft:v22.07
```

> :warning: If you are using a version of
[sriov-network-device-plugin](https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin) newer than v3.5.1,
you will need to apply the "``patches/uft-fix.patch``" file from the IEO repository on "``images/entrypoint.sh``"
located in the UFT repository before building the UFT image (running ``make dcf-image``).

```shell
$ patch -u <UFT_repo>/images/entrypoint.sh < <IEO_repo>/patches/uft-fix.patch
```

#### Creating FlowConfig Node Agent Deployment CR (E810 NICs only)

> Note: The Admin VF pool prefix in `DCFVfPoolName` should match how it is shown on node description in
[Checking node status](#checking-node-status-e810-nics-only) section.

```yaml
apiVersion: flowconfig.intel.com/v1
kind: FlowConfigNodeAgentDeployment
metadata:
  labels:
    control-plane: flowconfig-daemon
  name: flowconfig-daemon-deployment
  namespace: intel-ethernet-operator
spec:
  DCFVfPoolName: openshift.io/cvl_uft_admin
  NADAnnotation: sriov-cvl-dcf

```

#### Verifying that FlowConfig Daemon is running on available nodes (E810 NICs only)

```shell
$ kubectl get pods -n intel-ethernet-operator
NAME                                                          READY   STATUS    RESTARTS   AGE
cvl-discovery-kwjkb                                           1/1     Running   0          44h
cvl-discovery-tpqzb                                           1/1     Running   0          44h
flowconfig-daemon-worker-01                                   2/2     Running   0          44h
fwddp-daemon-m8d4w                                            1/1     Running   0          44h
intel-ethernet-operator-controller-manager-79c4d5bf6d-bjlr5   1/1     Running   0          44h
intel-ethernet-operator-controller-manager-79c4d5bf6d-txj5q   1/1     Running   0          44h

$ kubectl logs -n intel-ethernet-operator flowconfig-daemon-worker-01 -c uft
Generating server_conf.yaml file...
Done!
server :
    ld_lib : "/usr/local/lib64"
ports_info :
    - pci  : "0000:18:01.0"
      mode : dcf
do eal init ...
[{'pci': '0000:18:01.0', 'mode': 'dcf'}]
[{'pci': '0000:18:01.0', 'mode': 'dcf'}]
the dcf cmd line is: a.out -c 0x30 -n 4 -a 0000:18:01.0,cap=dcf -d /usr/local/lib64 --file-prefix=dcf --
EAL: Detected 96 lcore(s)
EAL: Detected 2 NUMA nodes
EAL: Detected shared linkage of DPDK
EAL: Multi-process socket /var/run/dpdk/dcf/mp_socket
EAL: Selected IOVA mode 'VA'
EAL: No available 1048576 kB hugepages reported
EAL: VFIO support initialized
EAL: Using IOMMU type 1 (Type 1)
EAL: Probe PCI driver: net_iavf (8086:1889) device: 0000:18:01.0 (socket 0)
EAL: Releasing PCI mapped resource for 0000:18:01.0
EAL: Calling pci_unmap_resource for 0000:18:01.0 at 0x2101000000
EAL: Calling pci_unmap_resource for 0000:18:01.0 at 0x2101020000
EAL: Using IOMMU type 1 (Type 1)
EAL: Probe PCI driver: net_ice_dcf (8086:1889) device: 0000:18:01.0 (socket 0)
ice_load_pkg_type(): Active package is: 1.3.30.0, ICE COMMS Package (double VLAN mode)
TELEMETRY: No legacy callbacks, legacy socket not created
grpc server start ...
now in server cycle
```

### Creating Flow Configuration rules with ClusterFlowConfig (E810 NICs only)

With trusted VFs and application VFs ready to be configured, create a sample `ClusterFlowConfig` CR:

Please see the [ClusterFlowConfig Spec](flowconfig-daemon/creating-rules.md) for detailed specification of supported
rules. Also, please note that this `ClusterFlowConfig` CR will create `NodeFlowConfig` with rules on nodes, on which
pods matching `podSelector` will be present.

>Note: By default, first additional network interface added to pod from Multus will be named `net1`. That is why
`spec.rules.pattern[0].action[0].conf.podInterface` field in provided example is set to `net1`.

```yaml
apiVersion: flowconfig.intel.com/v1
kind: ClusterFlowConfig
metadata:
  name: pppoes-sample
  namespace: intel-ethernet-operator
spec:
  rules:
    - pattern:
        - type: RTE_FLOW_ITEM_TYPE_ETH
        - type: RTE_FLOW_ITEM_TYPE_IPV4
          spec:
            hdr:
              src_addr: 10.56.217.9
          mask:
            hdr:
              src_addr: 255.255.255.255
        - type: RTE_FLOW_ITEM_TYPE_END
      action:
        - type: to-pod-interface
          conf:
            podInterface: net1
      attr:
        ingress: 1
        priority: 0
  podSelector:
      matchLabels:
        app: vagf
        role: controlplane

```

To verify if the flow rules has been applied, sample pod that meets the criteria needs to be created.

First create SRIOV pod network to which sample pod will be attached.

```yaml
apiVersion: sriovnetwork.openshift.io/v1
kind: SriovNetwork
metadata:
  name: sriov-podnet
  namespace: sriov-network-operator
spec:
  networkNamespace: intel-ethernet-operator
  resourceName: cvl_iavf
  ipam: |-
    {
      "type": "host-local",
      "subnet": "10.56.217.0/24",
      "rangeStart": "10.56.217.171",
      "rangeEnd": "10.56.217.181",
      "routes": [
      {
        "dst": "0.0.0.0/0"
      }
    ],
    "gateway": "10.56.217.1"
    }

```

When SRIOV pod network is ready to be utilized by pods, create sample pod itself.

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: example-pod
  namespace: intel-ethernet-operator
  labels:
    app: vagf
    role: controlplane
  annotations:
    k8s.v1.cni.cncf.io/networks: sriov-podnet
spec:
  containers:
    - name: appcntr
      image: alpine
      command:
        - /bin/sh
        - '-c'
        - '--'
      args:
        - ' while true; do sleep 30; done '
      resources:
        limits:
          openshift.io/cvl_iavf: '1'
        requests:
          openshift.io/cvl_iavf: '1'
      imagePullPolicy: IfNotPresent

```

Validate that Flow Rules are applied by the controller from UFT logs.

```text
$ kubectl logs -n intel-ethernet-operator flowconfig-daemon-worker-01 -c uft
Generating server_conf.yaml file...
Done!
server :
    ld_lib : "/usr/local/lib64"
ports_info :
    - pci  : "0000:18:01.0"
      mode : dcf
do eal init ...
[{'pci': '0000:18:01.0', 'mode': 'dcf'}]
[{'pci': '0000:18:01.0', 'mode': 'dcf'}]
the dcf cmd line is: a.out -c 0x30 -n 4 -a 0000:18:01.0,cap=dcf -d /usr/local/lib64 --file-prefix=dcf --
EAL: Detected 96 lcore(s)
EAL: Detected 2 NUMA nodes
EAL: Detected shared linkage of DPDK
EAL: Multi-process socket /var/run/dpdk/dcf/mp_socket
EAL: Selected IOVA mode 'VA'
EAL: No available 1048576 kB hugepages reported
EAL: VFIO support initialized
EAL: Using IOMMU type 1 (Type 1)
EAL: Probe PCI driver: net_iavf (8086:1889) device: 0000:18:01.0 (socket 0)
EAL: Releasing PCI mapped resource for 0000:18:01.0
EAL: Calling pci_unmap_resource for 0000:18:01.0 at 0x2101000000
EAL: Calling pci_unmap_resource for 0000:18:01.0 at 0x2101020000
EAL: Using IOMMU type 1 (Type 1)
EAL: Probe PCI driver: net_ice_dcf (8086:1889) device: 0000:18:01.0 (socket 0)
ice_load_pkg_type(): Active package is: 1.3.30.0, ICE COMMS Package (double VLAN mode)
TELEMETRY: No legacy callbacks, legacy socket not created
grpc server start ...
now in server cycle
flow.rte_flow_attr
flow.rte_flow_item
flow.rte_flow_item
flow.rte_flow_item_ipv4
flow.rte_ipv4_hdr
flow.rte_flow_item_ipv4
flow.rte_ipv4_hdr
flow.rte_flow_item
flow.rte_flow_action
flow.rte_flow_action_vf
flow.rte_flow_action
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0) [rte_flow_item(type_=9, spec=None, last=None, mask=None), rte_flow_item(type_=11, spec=rte_flow_item_ipv4(hdr=rte_ipv4_hdr(version_ihl=0, type_of_service=0, total_length=0, packet_id=0, fragment_offset=0, time_to_live=0, next_proto_id=0, hdr_checksum=0, src_addr=171497737, dst_addr=0)), last=None, mask=rte_flow_item_ipv4(hdr=rte_ipv4_hdr(version_ihl=0, type_of_service=0, total_length=0, packet_id=0, fragment_offset=0, time_to_live=0, next_proto_id=0, hdr_checksum=0, src_addr=4294967295, dst_addr=0))), rte_flow_item(type_=0, spec=None, last=None, mask=None)] [rte_flow_action(type_=11, conf=rte_flow_action_vf(reserved=0, original=0, id=2)), rte_flow_action(type_=0, conf=None)]
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0)
1
Finish ipv4: {'hdr': {'version_ihl': 0, 'type_of_service': 0, 'total_length': 0, 'packet_id': 0, 'fragment_offset': 0, 'time_to_live': 0, 'next_proto_id': 0, 'hdr_checksum': 0, 'src_addr': 165230602, 'dst_addr': 0}}
Finish ipv4: {'hdr': {'version_ihl': 0, 'type_of_service': 0, 'total_length': 0, 'packet_id': 0, 'fragment_offset': 0, 'time_to_live': 0, 'next_proto_id': 0, 'hdr_checksum': 0, 'src_addr': 4294967295, 'dst_addr': 0}}
rte_flow_action(type_=11, conf=rte_flow_action_vf(reserved=0, original=0, id=2))
rte_flow_action_vf(reserved=0, original=0, id=2)
Action vf:  {'reserved': 0, 'original': 0, 'id': 2}
rte_flow_action(type_=0, conf=None)
Validate ok...
flow.rte_flow_attr
flow.rte_flow_item
flow.rte_flow_item
flow.rte_flow_item_ipv4
flow.rte_ipv4_hdr
flow.rte_flow_item_ipv4
flow.rte_ipv4_hdr
flow.rte_flow_item
flow.rte_flow_action
flow.rte_flow_action_vf
flow.rte_flow_action
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0) [rte_flow_item(type_=9, spec=None, last=None, mask=None), rte_flow_item(type_=11, spec=rte_flow_item_ipv4(hdr=rte_ipv4_hdr(version_ihl=0, type_of_service=0, total_length=0, packet_id=0, fragment_offset=0, time_to_live=0, next_proto_id=0, hdr_checksum=0, src_addr=171497737, dst_addr=0)), last=None, mask=rte_flow_item_ipv4(hdr=rte_ipv4_hdr(version_ihl=0, type_of_service=0, total_length=0, packet_id=0, fragment_offset=0, time_to_live=0, next_proto_id=0, hdr_checksum=0, src_addr=4294967295, dst_addr=0))), rte_flow_item(type_=0, spec=None, last=None, mask=None)] [rte_flow_action(type_=11, conf=rte_flow_action_vf(reserved=0, original=0, id=2)), rte_flow_action(type_=0, conf=None)]
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0)
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0)
1
Finish ipv4: {'hdr': {'version_ihl': 0, 'type_of_service': 0, 'total_length': 0, 'packet_id': 0, 'fragment_offset': 0, 'time_to_live': 0, 'next_proto_id': 0, 'hdr_checksum': 0, 'src_addr': 165230602, 'dst_addr': 0}}
Finish ipv4: {'hdr': {'version_ihl': 0, 'type_of_service': 0, 'total_length': 0, 'packet_id': 0, 'fragment_offset': 0, 'time_to_live': 0, 'next_proto_id': 0, 'hdr_checksum': 0, 'src_addr': 4294967295, 'dst_addr': 0}}
rte_flow_action(type_=11, conf=rte_flow_action_vf(reserved=0, original=0, id=2))
rte_flow_action_vf(reserved=0, original=0, id=2)
Action vf:  {'reserved': 0, 'original': 0, 'id': 2}
rte_flow_action(type_=0, conf=None)
free attr
free item ipv4
free item ipv4
free list item
free action vf conf
free list action
Flow rule #0 created on port 0
```

### Creating Flow Configuration rules with NodeFlowConfig (E810 NICs only)

If `ClusterFlowConfig` does not satisfy your use case, you can use `NodeFlowConfig`. Create a sample, node specific
`NodeFlowConfig` CR with the same name as a target node. It should have empty spec:

```yaml
apiVersion: flowconfig.intel.com/v1
kind: NodeFlowConfig
metadata:
  name: worker-01
spec:
```

Check status of CR:

```shell
$ kubectl describe nodeflowconfig worker-01

Name:         worker-01
Namespace:    intel-ethernet-operator
Labels:       <none>
Annotations:  <none>
API Version:  flowconfig.intel.com/v1
Kind:         NodeFlowConfig
Metadata:
Status:
  Port Info:
    Port Id:    0
    Port Mode:  dcf
    Port Pci:   0000:18:01.0
Events:         <none>

```

You can see the DCF port information from NodeFlowConfig CR status for a node. These port information can be used to
identify for which port on a node the Flow rules should be applied.

#### Updating a sample Node Flow Configuration rule (E810 NICs only)

Please see the [NodeFlowConfig Spec](flowconfig-daemon/creating-rules.md) for detailed specification of supported rules.
You can update the `NodeFlowConfig` with a sample rule for a target port as shown below:

```yaml
apiVersion: flowconfig.intel.com/v1
kind: NodeFlowConfig
metadata:
  name: worker-01
  namespace: intel-ethernet-operator
spec:
  rules:
    - pattern:
        - type: RTE_FLOW_ITEM_TYPE_ETH
        - type: RTE_FLOW_ITEM_TYPE_IPV4
          spec:
            hdr:
              src_addr: 10.56.217.9
          mask:
            hdr:
              src_addr: 255.255.255.255
        - type: RTE_FLOW_ITEM_TYPE_END
      action:
        - type: RTE_FLOW_ACTION_TYPE_DROP
        - type: RTE_FLOW_ACTION_TYPE_END
      portId: 0
      attr:
        ingress: 1

```

Validate that Flow Rules are applied by the controller from UFT logs.

```text
$ kubectl logs -n intel-ethernet-operator flowconfig-daemon-worker-01 -c uft
Generating server_conf.yaml file...
Done!
server :
    ld_lib : "/usr/local/lib64"
ports_info :
    - pci  : "0000:5e:01.0"
      mode : dcf
server's pid=14
do eal init ...
[{'pci': '0000:5e:01.0', 'mode': 'dcf'}]
[{'pci': '0000:5e:01.0', 'mode': 'dcf'}]
the dcf cmd line is: a.out -v -c 0x30 -n 4 -a 0000:5e:01.0,cap=dcf -d /usr/local/lib64 --file-prefix=dcf --
EAL: Detected CPU lcores: 88
EAL: Detected NUMA nodes: 2
EAL: RTE Version: 'DPDK 22.07.0'
EAL: Detected shared linkage of DPDK
EAL: Multi-process socket /var/run/dpdk/dcf/mp_socket
EAL: Selected IOVA mode 'VA'
EAL: VFIO support initialized
EAL: Using IOMMU type 1 (Type 1)
EAL: Probe PCI driver: net_ice_dcf (8086:1889) device: 0000:5e:01.0 (socket 0)
ice_load_pkg_type(): Active package is: 1.3.37.0, ICE COMMS Package (double VLAN mode)
TELEMETRY: No legacy callbacks, legacy socket not created
grpc server start ...
now in server cycle
flow.rte_flow_attr
flow.rte_flow_item
flow.rte_flow_item
flow.rte_flow_item_ipv4
flow.rte_ipv4_hdr
flow.rte_flow_item_ipv4
flow.rte_ipv4_hdr
flow.rte_flow_item
flow.rte_flow_action
flow.rte_flow_action
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0) [rte_flow_item(type_=9, spec=None, last=None, mask=None), rte_flow_item(type_=11, spec=rte_flow_item_ipv4(hdr=rte_ipv4_hdr(version_ihl=0, type_of_service=0, total_length=0, packet_id=0, fragment_offset=0, time_to_live=0, next_proto_id=0, hdr_checksum=0, src_addr=171497737, dst_addr=0)), last=None, mask=rte_flow_item_ipv4(hdr=rte_ipv4_hdr(version_ihl=0, type_of_service=0, total_length=0, packet_id=0, fragment_offset=0, time_to_live=0, next_proto_id=0, hdr_checksum=0, src_addr=4294967295, dst_addr=0))), rte_flow_item(type_=0, spec=None, last=None, mask=None)] [rte_flow_action(type_=7, conf=None), rte_flow_action(type_=0, conf=None)]
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0)
1
Finish ipv4: {'hdr': {'version_ihl': 0, 'type_of_service': 0, 'total_length': 0, 'packet_id': 0, 'fragment_offset': 0, 'time_to_live': 0, 'next_proto_id': 0, 'hdr_checksum': 0, 'src_addr': 165230602, 'dst_addr': 0}}
Finish ipv4: {'hdr': {'version_ihl': 0, 'type_of_service': 0, 'total_length': 0, 'packet_id': 0, 'fragment_offset': 0, 'time_to_live': 0, 'next_proto_id': 0, 'hdr_checksum': 0, 'src_addr': 4294967295, 'dst_addr': 0}}
rte_flow_action(type_=7, conf=None)
rte_flow_action(type_=0, conf=None)
Validate ok...
flow.rte_flow_attr
flow.rte_flow_item
flow.rte_flow_item
flow.rte_flow_item_ipv4
flow.rte_ipv4_hdr
flow.rte_flow_item_ipv4
flow.rte_ipv4_hdr
flow.rte_flow_item
flow.rte_flow_action
flow.rte_flow_action
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0) [rte_flow_item(type_=9, spec=None, last=None, mask=None), rte_flow_item(type_=11, spec=rte_flow_item_ipv4(hdr=rte_ipv4_hdr(version_ihl=0, type_of_service=0, total_length=0, packet_id=0, fragment_offset=0, time_to_live=0, next_proto_id=0, hdr_checksum=0, src_addr=171497737, dst_addr=0)), last=None, mask=rte_flow_item_ipv4(hdr=rte_ipv4_hdr(version_ihl=0, type_of_service=0, total_length=0, packet_id=0, fragment_offset=0, time_to_live=0, next_proto_id=0, hdr_checksum=0, src_addr=4294967295, dst_addr=0))), rte_flow_item(type_=0, spec=None, last=None, mask=None)] [rte_flow_action(type_=7, conf=None), rte_flow_action(type_=0, conf=None)]
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0)
rte_flow_attr(group=0, priority=0, ingress=1, egress=0, transfer=0, reserved=0)
1
Finish ipv4: {'hdr': {'version_ihl': 0, 'type_of_service': 0, 'total_length': 0, 'packet_id': 0, 'fragment_offset': 0, 'time_to_live': 0, 'next_proto_id': 0, 'hdr_checksum': 0, 'src_addr': 165230602, 'dst_addr': 0}}
Finish ipv4: {'hdr': {'version_ihl': 0, 'type_of_service': 0, 'total_length': 0, 'packet_id': 0, 'fragment_offset': 0, 'time_to_live': 0, 'next_proto_id': 0, 'hdr_checksum': 0, 'src_addr': 4294967295, 'dst_addr': 0}}
rte_flow_action(type_=7, conf=None)
rte_flow_action(type_=0, conf=None)
free attr
free item ipv4
free item ipv4
free list item
free list action
Flow rule #0 created on port 0
```

## Uninstalling operator (X710/E810 NICs)

### Deployed by OLM (Operator Lifecycle Manager, X710/E810 NICs)

>Note: If you installed Operator from RedHat Certified Operators catalog or OperatorHub catalog, it means that OLM was
used as deployment method.

Uninstalling of the Operator should be done by deleting OLM `ClusterServiceVersion` and `Subscription` CRs from
`intel-ethernet-operator` namespace. In case flowconfig-daemon was deployed, ```FlowConfigNodeAgentDeployment``` CR also
needs to be deleted prior to uninstalling of operator itself:

```shell
$ kubectl -n intel-ethernet-operator delete flowconfignodeagentdeployments flowconfig-daemon-flowconfig-daemon
```

To uninstall operator execute following commands:

```shell
$ CSV=$(kubectl get subscription intel-ethernet-subscription -n intel-ethernet-operator -o json | jq -r '.status.installedCSV')
$ kubectl delete subscription intel-ethernet-subscription -n intel-ethernet-operator
$ kubectl delete csv $CSV -n intel-ethernet-operator
```

Replace names of resources according to ones that were used. Deleting namespace without prior deleting of resources
inside it can lead to namespace being stuck at termination state. To delete namespace:

```shell
$ kubectl delete ns intel-ethernet-operator
```

More information can be found here <https://olm.operatorframework.io/docs/tasks/uninstall-operator/>

### Deployed by Helm (X710/E810 NICs)

If you used `make helm-deploy` to deploy Operator, you can uninstall it by executing:

```shell
$ make helm-undeploy
```

If you used different release name or namespace, you need to execute following command:

```shell
$ helm uninstall <release-name> --namespace=<release-namespace>
```

Be sure to substitute `<release-name` and `release-namespace` with values used in deployment process.

Then you can delete Operator namespace:

```shell
$ kubectl delete ns <release-namespace>
```

## Hardware Validation Environment (X710/E810 NICs)

- Intel® Ethernet Network Adapter X710-DA2
- Intel® Ethernet Network Adapter E810-XXVDA2
- 3nd Generation Intel® Xeon® processor platform

## Summary

The Intel Ethernet Operator is a functional tool to manage the update of Intel® X710/E810 NICs FW and E810 NICs DDP profile, as well as
the programming of the E810 NICs VFs Flow Configuration autonomously in a Cloud Native OpenShift environment based on user
input. It is easy in use by providing simple steps to apply the Custom Resources to configure various aspects of the
device.
