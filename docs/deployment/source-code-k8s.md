```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2024 Intel Corporation
```

# Deploy Intel Ethernet Operator from source code on K8s

## Technical Requirements and Prerequisites (X710/E810)

- Intel® Ethernet Network Adapter X710/E810
- Kubernetes version 1.21 or newer
- Kubernetes is running on top of distribution with in-tree ICE driver supported by
  [NVM Update tool](../intel-ethernet-operator.md#ice-driver-variant-e810-nics-only) (second paragraph, E810 only)

### :warning: If */lib/firmware* directory is read-only on your nodes(E810 only)

If */lib/firmware* directory is read-only on your nodes, necessarily see this
[section](../intel-ethernet-operator.md#warning-alternative-firmware-search-path-on-nodes-with-libfirmware-read-only).

### Deploy systemd ICE driver reload service (needed only for DDP configuration functionality, E810 only)

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

### Additional prerequisites only for if you want to use **Flow Configuration** functionality (E810 only)

- Out-of-tree ICE driver 1.9.11
- IOMMU enabled
- Hugepage memory configured
- SRIOV Network Operator deployed in the cluster

### Out-of-tree ICE driver update (E810 only)

In order for the Flow Configuration to be possible, the platform needs to provide an OOT ICE driver 1.9.11. More
information can be found [here](../intel-ethernet-operator.md#ice-driver-variant-e810-nics-only). You can provide such driver using
[K8s KMMO ICE driver install](../oot-ice-driver/kmm-ice-install-k8s.md) guide. If you don't want to use `KMMO`, you can
of course provide Out-of-tree ICE driver 1.9.11 manually.

### Enabling IOMMU (E810 only)

IOMMU needs to be enabled in order for Flow Configuration to be possible. This usually consists of making changes in
node UEFI/BIOS configuration. For more information refer to
[IOMMU section](../intel-ethernet-operator.md#iommu-needed-only-for-flow-configuration-functionality-e810-nics-only) on main page.

### Configuring Hugepages (E810 only)

Hugepages needs to be configured on nodes. See
[Hugepages](../intel-ethernet-operator.md#hugepages-needed-only-for-flow-configuration-functionality-e810-nics-only) section.

### Deploying SRIOV Network Operator (E810 only)

To deploy SRIOV Network Operator visit
[installation guide](https://github.com/k8snetworkplumbingwg/sriov-network-operator/blob/master/doc/quickstart.md).
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

`Operator Lifecycle Manager` must be deployed in your cluster. You can install it by executing:

```shell
$ curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.25.0/install.sh | bash -s v0.25.0
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

#### Enable mTLS for Flowconfig validation webhook in OLM (optional, X710/E810)

If [client certificate verification](#validation-webhook-mtls-optional-e810-only) in Flowconfig `Validating Webhook` server is
required, create `ConfigMap`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: webhook-config
  namespace: intel-ethernet-operator
data:
  enable-webhook-mtls: "true"
```

and if the self signed certificate will be used for the kube-apiserver, create the secret with the CA certificate:

```shell
$ kubectl create secret generic webhook-client-ca --from-file=ca.crt=<filename> --namespace=intel-ethernet-operator
```

Enabling the mTLS without specifying the `webhook-client-ca` secret, tells the webhook server to verify client
certificates using Kubernetes general CA, by default mounted to pod at
`/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`

---

Create a namespace for the operator:

```shell
$ kubectl create ns intel-ethernet-operator
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
  namespace: intel-ethernet-operator
spec:
  sourceType: grpc
  image: <IMAGE_REGISTRY>/intel-ethernet-operator-catalog:<VERSION>
  publisher: Intel
  displayName: Intel ethernet operators (Local)
```

```shell
$ kubectl apply -f <filename>
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

>Note: If your cluster is behind proxy and you want to use external `fwURL`(X710/E810) and `ddpURL`(E810 only) in `EthernetClusterConfig`,
then you need to configure proxy on cluster. You can configure it by setting HTTP_PROXY, HTTPS_PROXY and NO_PROXY
environmental variables in
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
  sourceNamespace: intel-ethernet-operator
  installPlanApproval: Automatic
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

### Deploy using Helm (X710/E810)

Create a namespace for the operator:

```shell
$ kubectl create ns intel-ethernet-operator
```

Update helm `values.yaml` and `Chart.yaml` files with location and tags of images built in previous step.

```shell
$ make VERSION=$(VERSION) IMAGE_REGISTRY=$(IMAGE_REGISTRY) helm-configure
```

#### Review deployment settings (X710/E810)

Review [`values.yaml`](../../deployment/intel-ethernet-operator/values.yaml) file. `image.repository` and `image.tag`
should be already updated by command executed in the last step. You can either use default values or customize
deployment to match your needs.

Please pay attention to `webhookServerConfiguration`(E810 only). These settings correspond to `Validating Webhook` that
verifies `NodeFlowConfig` and `ClusterFlowConfig` CRs. When you deploy Operator with `Operator Lifecycle Manager` (OLM)
certificates for that `Webhook` are provided by `OLM`. When deploying with `helm`, user is responsible for provision of
certificates.

>:warning: **If you don't plan to use Flow Configuration, you can disable the webhook completly**, as it's used
only for validating CRs associated with Flow Configuration.

To achieve this you can use `CertManager`. If you have `CertManager` installed in the cluster just make sure that
`webhookServerConfiguration.useCertManager` is `true`. If you don't have `CertManager` installed, but you want to use
it, see [official product installation guide](https://cert-manager.io/docs/installation/).

If you would prefer not to use `CertManager`, you can provide server certificate, server key and CA certificate
yourself. First make sure that `webhookServerConfiguration.useCertManager` is set to `false`. Then set variables with
location of files containing certificates.

Be aware that some files cannot be accessed:

- Files in templates/ cannot be accessed.
- Files excluded using .helmignore cannot be accessed.
- Files outside of a helm application subchart, including those of the parent, cannot be accessed

If [client certificate verification](#validation-webhook-mtls-optional-e810-only) in Flowconfig `Validating Webhook` server is
required(E810 only), then you need to set `webhookServerConfiguration.mTLSWebhook.enable` to `true`. If the self signed certificate
will be used for the kube-apiserver, set `clientCaFilepath` variable with location of file containing CA certificate.
Note that restrictions as to location of the file are same as listed above. Enabling the mTLS without specifying the
`clientCaFilepath` tells the webhook server to verify client certificates using Kubernetes general CA, by default
mounted to pod at `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.

>Note: If your cluster is behind proxy and you want to use external `fwURL`(X710/E810) and `ddpURL`(E810 only) in `EthernetClusterConfig`,
then you need to configure proxy on cluster. You can configure it by setting `proxy.httpProxy`, `proxy.httpsProxy` and
`proxy.noProxy` in `values.yaml`.

---

When you are sure `values.yaml` has all the settings you want, deploy the Operator.

```shell
$ make helm-deploy
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

### Validation webhook mTLS (optional, E810 only)

If mTLS is required for kube-apiserver->webhook communication,
[cluster configuration](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/control-plane-flags#customizing-the-control-plane-with-flags-in-clusterconfiguration)
has to be extended with additional steps. Optionally, it is achievable by modifying the kube-apiserver manifest file:

 ```text
 /etc/kubernetes/manifests/kube-apiserver.yaml
 ```

and restarting the kubelet. When client certificate verification is enabled in webhook server, kube-apiserver
configuration has to contain `--admission-control-config-file`
[flag](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/) pointing to the
`AdmissionConfiguration` file (acessible from the inside of the kube-apiserver pod). In example below,
`/etc/kubernetes/pki/` directory was re-used, as it's already mounted to the pod:

```yaml
cat /etc/kubernetes/manifests/kube-apiserver.yaml

apiVersion: v1
kind: Pod
metadata:
  ...
  name: kube-apiserver
  namespace: kube-system
spec:
  containers:
  - command:
    - kube-apiserver
    ...
    - --admission-control-config-file=/etc/kubernetes/pki/admissioncfg.yaml
```

```yaml
cat /etc/kubernetes/pki/admissioncfg.yaml

apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
- name: ValidatingAdmissionWebhook
  configuration:
    apiVersion: apiserver.config.k8s.io/v1
    kind: WebhookAdmissionConfiguration
    kubeConfigFile: "/etc/kubernetes/pki/admission_kubeconfig"
```

The `AdmissionConfiguration` file must point to the custom kubeConfig file that contains the paths to the key/certficate
pair which will be used by the kube-apiserver when reaching the webhook service.

```yaml
cat /etc/kubernetes/pki/admission_kubeconfig

apiVersion: v1
kind: Config
users:
- name: "intel-ethernet-operator-controller-manager-service.intel-ethernet-operator.svc"
  user:
    client-certificate: /etc/kubernetes/pki/apiserver-webhook-client.crt
    client-key: /etc/kubernetes/pki/apiserver-webhook-client.key
```

Certificate above, will be verified by the webhook server with the certificate provided by the `webhook-client-ca`
secret or Kubernetes general CA (`/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`).

In the presented example, cluster admin is responsible for creation custom CA and key/certificate pair. For more
information, please see the
[k8s reference](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#authenticate-apiservers).

## Uninstalling Operator (X710/E810)

Steps to uninstall Intel Ethernet Operator can be found [here](../intel-ethernet-operator.md#uninstalling-operator-x710e810-nics).
