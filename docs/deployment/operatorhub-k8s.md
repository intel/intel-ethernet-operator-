```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2024 Intel Corporation
```

# Deploy Intel Ethernet Operator from OperatorHub Catalog on K8s (X710/E810)

## Technical Requirements and Prerequisites (X710/E810)

- Intel® Ethernet Network Adapter X710/E810
- Kubernetes version 1.21 or newer
- Kubernetes is running on top of distribution with in-tree ICE driver supported by
  [NVM Update tool](../intel-ethernet-operator.md#ice-driver-variant-e810-nics-only) (second paragraph, E810 only)
- Operator Lifecycle Manager (OLM) is installed on cluster.

>Note: Ensure that `/lib/firmware` directory is not immutable on used Linux distribution(E810 only). This is not the case for
large majority of the systems, but if `/lib/firmware` is read-only on your nodes, please refer to
[alternative firmware search path section](../intel-ethernet-operator.md#warning-alternative-firmware-search-path-on-nodes-with-libfirmware-read-only)
for more information.

### Install OLM (X710/E810)

Install Operator Lifecycle Manager (OLM), a tool to help manage the Operators running on your cluster.

```shell
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.25.0/install.sh | bash -s 0.25.0
```

Verify OLM installation:

```shell
$ kubectl get pods -n olm
NAME                               READY   STATUS    RESTARTS   AGE
catalog-operator-9457dd57c-x58br   1/1     Running   0          2d
olm-operator-67fdb4b99d-hmdpm      1/1     Running   0          2d
operatorhubio-catalog-d2h92        1/1     Running   0          17h
packageserver-655574f96c-6g5g9     1/1     Running   0          2d
packageserver-655574f96c-dx5bs     1/1     Running   0          2d
```

### Deploy systemd ICE driver reload service (needed only for DDP configuration functionality) (E810 only)

To enable ICE driver reload on boot, which is needed for DDP update functionality, please create suitable systemd
service. If you seek more information, see
[ICE driver reload section](../intel-ethernet-operator.md#warning-ice-driver-reload-after-reboot) on main page.

If you need guidence, follow steps below:

Create `ice-driver-reload.sh` script on every node on which you need DDP configuration functionality. You can find the
script [here](../../extras/ice-driver-reload.sh). Place it in a directory of your choice, in this example it
will be `/home/user/ice-driver-reload.sh`.

Make sure script is executable.

```shell
$ chmod +x /home/user/ice-driver-reload.sh
```

Create systemd service by creating file named `ice-driver-reload.service` with content shown below in
`/etc/systemd/system/` directory:

>Note: Depending on used Linux distribution and its version, systemd service file location might be different.

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
ExecStart=/usr/bin/bash /home/user/ice-driver-reload.sh
StandardOutput=journal+console
[Install]
WantedBy=default.target
```

Enable systemd service.

```shell
$ systemctl daemon-reload
$ systemctl enable ice-driver-reload.service
```

On the next reboot of the node, you can check if service has been executed properly by using following command:

```shell
$ systemctl status ice-driver-reload.service
● ice-driver-reload.service - ice driver reload
     Loaded: loaded (/etc/systemd/system/ice-driver-reload.service; enabled; vendor preset: enabled)
     Active: active (exited) since Wed 2024-11-22 16:43:11 UTC; 2 weeks 0 days ago
   Main PID: 2086 (code=exited, status=0/SUCCESS)
        CPU: 261ms

Nov 22 16:43:08 silpixa00401329a.ir.intel.com systemd[1]: Starting ice driver reload...
Nov 22 16:43:11 silpixa00401329a.ir.intel.com systemd[1]: Finished ice driver reload.
```

## Deploy the Operator (X710/E810)

### Deploying from manifests located on OperatorHub website (X710/E810)

Install the operator by running the following command:

```shell
$ kubectl create -f https://operatorhub.io/install/intel-ethernet-operator.yaml
```

Check if the operator has deployed successfully:

```shell
$ kubectl get pods -n my-intel-ethernet-operator
NAME                                                          READY   STATUS    RESTARTS      AGE
cvl-discovery-db6j7                                           1/1     Running   0             23h
cvl-discovery-fl5n6                                           1/1     Running   0             23h
fwddp-daemon-4cmn7                                            1/1     Running   0             23h
fwddp-daemon-5jjzw                                            1/1     Running   0             23h
intel-ethernet-operator-controller-manager-75d4449bfb-cx65b   1/1     Running   0             23h
intel-ethernet-operator-controller-manager-75d4449bfb-dhqv5   1/1     Running   0             23h
```

### Deploying from CLI (X710/E810)

Create a namespace for the operator:

```shell
$ kubectl create ns intel-ethernet-operator
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
$ kubectl apply -f <filename>
```

>Note: If your cluster is behind proxy and you want to use external `fwURL`(X710/E810) and `ddpURL`(E810 only) in
`EthernetClusterConfig`, then you need to configure proxy on cluster. You can configure it by setting HTTP_PROXY,
HTTPS_PROXY and NO_PROXY environmental variables in
[operator's subscription](https://docs.openshift.com/container-platform/4.14/operators/admin/olm-configuring-proxy-support.html).
Be aware that operator will ignore lowercase `http_proxy` variables and will accept only uppercase variables.

Then create and apply `Subscription` yaml file:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: intel-ethernet-operator
  namespace: intel-ethernet-operator
spec:
  channel: alpha
  name: intel-ethernet-operator
  source: operatorhubio-catalog
  sourceNamespace: olm
```

```shell
$ kubectl apply -f <filename>
```

Check if the Operator has deployed successfully:

```shell
$ kubectl get pods -n intel-ethernet-operator
NAME                                                          READY   STATUS    RESTARTS      AGE
cvl-discovery-db6j7                                           1/1     Running   0             23h
cvl-discovery-fl5n6                                           1/1     Running   0             23h
fwddp-daemon-4cmn7                                            1/1     Running   0             23h
fwddp-daemon-5jjzw                                            1/1     Running   0             23h
intel-ethernet-operator-controller-manager-75d4449bfb-cx65b   1/1     Running   0             23h
intel-ethernet-operator-controller-manager-75d4449bfb-dhqv5   1/1     Running   0             23h
```

## Uninstalling Operator (X710/E810)

Steps to uninstall Intel Ethernet Operator can be found [here](../intel-ethernet-operator.md#uninstalling-operator-x710e810-nics).
