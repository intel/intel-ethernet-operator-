```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2024 Intel Corporation
```

# Deploy Intel Ethernet Operator from RedHat Certified Operators Catalog on OCP (X710/E810)

## Technical Requirements and Prerequisites

- Intel® Ethernet Network Adapter X710/E810
- OpenShift 4.11-4.13

### Deploy MachineConfig that configures firmware search path (X710/E810)

In order to set the kernel parameter needed for correct functioning of Operator, please use the following command:

```shell
$ oc apply -f extras/fw-search-path-machine-config.yaml
```

See
[alternative firmware search path section](../intel-ethernet-operator.md#warning-alternative-firmware-search-path-on-nodes-with-libfirmware-read-only)
for more information.

### Deploy systemd ICE driver reload service (needed only for DDP configuration functionality, E810 only)

To enable ICE driver reload on boot, which is needed for DDP update functionality, please use the following command:

```shell
$ oc apply -f extras/ice-driver-reload-machine-config.yaml
```

Please refer to [ICE driver reload section](../intel-ethernet-operator.md#warning-ice-driver-reload-after-reboot) for
more information.

## Deploy the Operator (X710/E810)

Intel Ethernet Operator can be deployed from Red Hat Certified Operators catalog using CLI or OCP cluster web console.

### Deploying from CLI (X710/E810)

Create a namespace for the operator:

```shell
$ oc create ns intel-ethernet-operator
```

Create and apply the following `OperatorGroup` yaml file:

```yaml
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: intel-ethernet-operator
  namespace: intel-ethernet-operator
spec:
  targetNamespaces:
    - intel-ethernet-operator
```

```shell
$ oc apply -f <filename>
```

>Note: If your cluster is behind proxy and you want to use external `fwURL` (X710/E810) and `ddpURL`(E810 only) in `EthernetClusterConfig`,
then you need to configure proxy on cluster. You can configure it by using
[OCP cluster-wide proxy](https://docs.openshift.com/container-platform/4.14/networking/enable-cluster-wide-proxy.html)
or by setting HTTP_PROXY, HTTPS_PROXY and NO_PROXY environmental variables in
[operator's subscription](https://docs.openshift.com/container-platform/4.14/operators/admin/olm-configuring-proxy-support.html).
Be aware that operator will ignore lowercase `http_proxy` variables and will accept only uppercase variables.

Then create and apply `Subscription` yaml file:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: intel-ethernet-subscription
  namespace: intel-ethernet-operator
spec:
  channel: alpha
  name: intel-ethernet-operator
  source: certified-operators
  sourceNamespace: openshift-marketplace
```

```shell
$ oc apply -f <filename>
```

Check if the Operator has deployed successfully:

```shell
$ oc get pods -n intel-ethernet-operator
NAME                                                          READY   STATUS    RESTARTS      AGE
cvl-discovery-db6j7                                           1/1     Running   0             23h
cvl-discovery-fl5n6                                           1/1     Running   0             23h
fwddp-daemon-4cmn7                                            1/1     Running   0             23h
fwddp-daemon-5jjzw                                            1/1     Running   0             23h
intel-ethernet-operator-controller-manager-75d4449bfb-cx65b   1/1     Running   0             23h
intel-ethernet-operator-controller-manager-75d4449bfb-dhqv5   1/1     Running   0             23h
```

### Deploying from OCP web console (X710/E810)

Using `OpenShift Container Platform web console`:

1. In the OpenShift Container Platform web console, click Operators → OperatorHub.
2. Select Intel Ethernet Operator from the list of available Operators, and then click Install.
3. On the Install Operator page, under a specific namespace on the cluster, select intel-ethernet-operator.
4. Click Install.
5. Verify if all Operator resources has deployed successfully either using web console or CLI.

## Uninstalling Operator (X710/E810)

Steps to uninstall Intel Ethernet Operator can be found [here](../intel-ethernet-operator.md#uninstalling-operator-x710e810-nics).
