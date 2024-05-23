```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2024 Intel Corporation
```

# Deploy Intel Ethernet Operator from source code on OCP

## Technical Requirements and Prerequisites (X710/E810)

- Intel® Ethernet Network Adapter X710/E810
- OCP version 4.11 or newer

### Deploy MachineConfig that configures firmware search path (X710/E810)

In order to set the kernel parameter needed for correct functioning of Operator, please use the following command:

```shell
$ oc apply -f extras/fw-search-path-machine-config.yaml
```

### Deploy systemd ICE driver reload service (needed only for DDP configuration functionality, E810 only)

To enable ICE driver reload on boot, which is needed for DDP update functionality, please use the following command:

```shell
$ oc apply -f extras/ice-driver-reload-machine-config.yaml
```

### Additional prerequisites only for if you want to use **Flow Configuration** functionality (E810 only)

- Out-of-tree ICE driver 1.9.11
- IOMMU enabled
- Hugepage memory configured
- SRIOV Network Operator deployed in the cluster

### Out-of-tree ICE driver update (E810 only)

In order for the Flow Configuration to be possible, the platform needs to provide an OOT ICE driver 1.9.11. More
information can be found [here](../intel-ethernet-operator.md#ice-driver-variant-e810-nics-only). You can provide such driver using
[OCP KMMO ICE driver install](../oot-ice-driver/kmm-ice-install-ocp.md) guide.

### Enabling IOMMU (E810 only)

IOMMU needs to be enabled in order for Flow Configuration to be possible. This usually consists of making changes in
node UEFI/BIOS configuration. For more information refer to
[IOMMU section](../intel-ethernet-operator.md#iommu-needed-only-for-flow-configuration-functionality-e810-nics-only) on main page.

### Configuring Hugepages (E810 only)

Hugepages needs to be configured on nodes. See
[Hugepages](../intel-ethernet-operator.md#hugepages-needed-only-for-flow-configuration-functionality-e810-nics-only) section.

### Deploying SRIOV Network Operator (E810 only)

To deploy SRIOV Network Operator visit
[installation guide](https://docs.openshift.com/container-platform/4.14/networking/hardware_networks/installing-sriov-operator.html).
If you need more information as to why SRIOV Network Operator is needed for Flow Configuration, see this
[section](../intel-ethernet-operator.md#sriov-network-operator-needed-only-for-flow-configuration-functionality-e810-nics-only).

## Building the Operator from source (X710/E810)

Building the operator images will require Go and Operator SDK to be installed.

### Installing Go (X710/E810)

You can install Go following the steps [here](https://go.dev/doc/install).

> Note: Intel Ethernet Operator is based on Go 1.21.

### Installing Operator SDK (X710/E810)

Please install Operator SDK v1.32.0 following the steps below:

```shell
$ export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
$ export OS=$(uname | awk '{print tolower($0)}')
$ export SDK_VERSION=v1.32.0
$ export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/${SDK_VERSION}
$ curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
$ chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk
```

### Building the Operator images (X710/E810)

To build the Operator, the images must be built from source. In order to build, execute the following steps:

> Note: The arguments are to be replaced with the following:
>
- VERSION is the version to be applied to the bundle e.g. `2.0.0`.
- IMAGE_REGISTRY is the address of the registry where the build images are to be pushed to ie. `my.private.registry.com`.
- TLS_VERIFY defines whether connection to registry need TLS verification, default is `false`.
- USE_HTTP defines whether connection to registry uses HTTP protocol, default is `false` (HTTPS will be used)

```shell
$ make VERSION=$(VERSION) IMAGE_REGISTRY=$(IMAGE_REGISTRY) TLS_VERIFY=$(TLS_VERIFY) USE_HTTP=$(USE_HTTP) build_all push_all catalog-build catalog-push
```

## Deploying the Operator (X710/E810)

After building is done, you can deploy Operator using `Operator Lifecycle Manager` or `Helm`.

### Deploy using OLM (X710/E810)

Create a namespace for the operator:

```shell
$ oc create ns intel-ethernet-operator
```

Create and apply the following `Catalog Source` yaml file:

> Note: The REGISTRY_ADDRESS and VERSION need to be replaced:
>
> - VERSION is the version of images built in previous step e.g. `2.0.0`.
> - IMAGE_REGISTRY is the address of the registry where the built images were pushed ie. `my.private.registry.com`.

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: intel-ethernet-operators
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  image: <IMAGE_REGISTRY>/intel-ethernet-operator-catalog:<VERSION>
  publisher: Intel
  displayName: Intel ethernet operators (Local)
```

```shell
$ oc apply -f <filename>
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

>Note: If your cluster is behind proxy and you want to use external `fwURL`(X710/E810) and `ddpURL`(E810 only) in `EthernetClusterConfig`,
then you need to configure proxy on cluster. You can configure it by using
[OCP cluster-wide proxy](https://docs.openshift.com/container-platform/4.14/networking/enable-cluster-wide-proxy.html)
or by setting HTTP_PROXY, HTTPS_PROXY and NO_PROXY environmental variables in
[operator's subscription](https://docs.openshift.com/container-platform/4.14/operators/admin/olm-configuring-proxy-support.html).
Be aware that operator will ignore lowercase `http_proxy` variables and will accept only uppercase variables.

Then create and apply `Subscription` yaml:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: intel-ethernet-subscription
  namespace: intel-ethernet-operator
spec:
  channel: alpha
  name: intel-ethernet-operator
  source: intel-ethernet-operators
  sourceNamespace: openshift-marketplace
  installPlanApproval: Automatic
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

### Deploy using Helm (X710/E810)

Create a namespace for the operator:

```shell
$ oc create ns intel-ethernet-operator
```

Update helm `values.yaml` and `Chart.yaml` files with location and tags of images built in previous step.

```shell
$ make VERSION=$(VERSION) IMAGE_REGISTRY=$(IMAGE_REGISTRY) helm-configure
```

Review [`values.yaml`](../../deployment/intel-ethernet-operator/values.yaml) file. `image.repository` and `image.tag`
should be already updated by command executed in the last step. You can either use default values or customize
deployment to match your needs.

Please pay attention to `image.webhookServerConfiguration`(E810 only). These settings correspond to `Validating Webhook` that
verifies `NodeFlowConfig` and `ClusterFlowConfig` CRs. When you deploy Operator with `Operator Lifecycle Manager` (OLM)
certificates for that `Webhook` are provided by `OLM`. When deploying with `helm`, user is responsible for provision of
certificates.

>:warning: **If you don't plan to use Flow Configuration, you can disable the webhook completly**, as it's used
only for validating CRs associated with Flow Configuration.

To achieve this you can use `CertManager`. If you have `CertManager` installed in the cluster just make sure that
`image.useCertManager` is `true`. If you don't have `CertManager` installed, but you want to use it, see
[official product installation guide](https://cert-manager.io/docs/installation/).

If you would prefer not to use `CertManager`, you can provide server certificate, server key and CA certificate
yourself. First make sure that `image.useCertManager` is set to `false`. Then set variables with location of files
containing certificates.

Be aware that some files cannot be accessed:

- Files in templates/ cannot be accessed.
- Files excluded using .helmignore cannot be accessed.
- Files outside of a helm application subchart, including those of the parent, cannot be accessed

>Note: If your cluster is behind proxy and you want to use external `fwURL`(X710/E810) and `ddpURL`(E810 only) in `EthernetClusterConfig`,
then you need to configure proxy on cluster. You can configure it by using
[OCP cluster-wide proxy](https://docs.openshift.com/container-platform/4.14/networking/enable-cluster-wide-proxy.html)
or by setting `proxy.httpProxy`, `proxy.httpsProxy` and `proxy.noProxy` in `values.yaml`.

When you are sure `values.yaml` has all the settings you want, deploy the Operator.

```shell
$ make helm-deploy
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

## Uninstalling Operator (X710/E810)

Steps to uninstall Intel Ethernet Operator can be found [here](../intel-ethernet-operator.md#uninstalling-operator-x710e810-nics).
