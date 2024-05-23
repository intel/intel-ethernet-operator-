```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020-2024 Intel Corporation
```

# Install OOT (out of tree) ICE driver on OCP nodes (E810 only)

## Prerequisites

> Note: This guide was prepared and tested on OCP 4.13.9.

* OCP cluster 4.13.9 `(any OCP version supported by KMM Operator should work, but steps in this guide may require adjustments)`
* Redhat account with right subscription for Redhat registry access
* External image registry and OCP is able to access it `(you can skip this step and use internal OCP registry of which configuration is included in this guide)`

## SSH into cluster

```shell
$ ssh -i <path_to_key> core@<ip>
# or
$ oc debug node/<node_name>
```

## Install [Kernel Module Management Operator](https://docs.openshift.com/container-platform/4.13/hardware_enablement/kmm-kernel-module-management.html)

Install KMM either from OperatorHub or from CLI using following commands.

Create and apply `yaml` file containing Resources needed for installation.

```shell
$ vi kmm.yml
```

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: openshift-kmm
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: kernel-module-management
  namespace: openshift-kmm
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: kernel-module-management
  namespace: openshift-kmm
spec:
  channel: stable
  installPlanApproval: Automatic
  name: kernel-module-management
  source: redhat-operators
  sourceNamespace: openshift-marketplace
```

```shell
$ oc apply -f kmm.yml
```

### Verify that KMM is running in the cluster

```shell
$ oc get pods -n openshift-kmm
NAME                                              READY   STATUS    RESTARTS   AGE
kmm-operator-controller-manager-9b546d464-ghv8v   2/2     Running   0          65s
```

### Get Redhat image pull secret from Redhat subscription

Go to [Pull secret](https://console.redhat.com/openshift/install/pull-secret) page on Redhat OpenShift cluster manager
site and download the pull secret file. You will need to log in with your RH account. Copy the secret to clipboard or
save to a file. Either way create the secret file on client machine. Let's assume it is stored in `./rht_auth.json`
file.

```shell
$ vi ./rht_auth.json # copied from clipboard or file
```

Find out the right driver toolkit image needed for the cluster:
> Note: It is important to provide right cluster info - in case incorrect version is provided, the latest kernel headers
may not be located in the toolkit image.

```shell
$ oc adm release info 4.13.9 --image-for=driver-toolkit
quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:<some-version>
```

Pull this image locally on client machine using Podman with the authfile `./rht_auth.json` downloaded in previous step.

```shell
$ podman pull --authfile=./rht_auth.json quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:<some-version>
```

### Prepare image registry

> Important: To proceed with OOT ICE driver installation, image registry to which OCP cluster has access, needs to be
available. If you already have such registry configured, be it public (like [quay.io](<https://quay.io>)) or private,
you may use it and skip next, optional step. If not, you either need to configure access to existing external registry
or use OCP internal registry. Steps to configure internal registry can be found below.

#### Configure internal registry (optional)

Configuring the registry as per
<https://docs.openshift.com/container-platform/4.13/registry/configuring_registry_storage/configuring-registry-storage-baremetal.html>

```shell
$ oc patch configs.imageregistry.operator.openshift.io cluster --type merge --patch '{"spec":{"managementState":"Managed"}}'
```

> Note: "emptyDir" type of storage is ephemeral - in an event of a node reboot all image cache will be lost.
[See following guide for more info](https://docs.openshift.com/container-platform/4.13/registry/configuring_registry_storage/configuring-registry-storage-baremetal.html)

```shell
$ oc patch configs.imageregistry.operator.openshift.io cluster --type merge --patch '{"spec":{"storage":{"emptyDir":{}}}}'
```

Verify if internal registry pod is available.

```shell
$ oc get pods -n openshift-image-registry
NAME                                               READY   STATUS      RESTARTS   AGE
cluster-image-registry-operator-6c4b6696f4-dpr4s   1/1     Running     42         69d
image-pruner-28327680-bjlcp                        0/1     Completed   0          2d12h
image-pruner-28329120-h7vz7                        0/1     Completed   0          36h
image-pruner-28330560-hlk65                        0/1     Completed   0          12h
image-registry-58fb5d86cc-6wzcj                    1/1     Running     18         49d <--------- This one
node-ca-zpr6k                                      1/1     Running     42         69d
```

Login to internal registry to ensure it is working properly.

```shell
$ sudo podman login -u kubeadmin -p $(oc whoami -t) image-registry.openshift-image-registry.svc:5000
Login Succeeded!
```

### Prepare driver source code image

Create your ICE OOT Dockerfile, replace the target (OCP node) kernel version, ICE version and URL. The first base image
should be the driver toolkit you got by running `oc adm release info 4.13.9 --image-for=driver-toolkit`. You can get
kernel version of OCP nodes by executing `oc get nodes -o wide`. If you are working behind proxy, don't forget proxy
related settings.

```dockerfile
FROM <driver-toolkit> as builder

ENV HTTP_PROXY <your-proxy>
ENV HTTPS_PROXY <your-proxy>

WORKDIR /usr/src
RUN ["curl", "-X", "GET", "https://downloadmirror.intel.com/789309/ice-1.12.7.tar.gz", "--output", "ice-1.12.7.tar.gz"]
RUN ["tar","-xvf", "ice-1.12.7.tar.gz"]
WORKDIR /usr/src/ice-1.12.7/src
RUN ["make", "install"]

FROM registry.redhat.io/ubi9/ubi-minimal

ENV HTTP_PROXY <your-proxy>
ENV HTTPS_PROXY <your-proxy>

RUN microdnf install kmod -y

RUN mkdir -p /opt/lib/modules/<kernel-version>/
COPY --from=builder /usr/lib/modules/<kernel-version>/updates/drivers/net/ethernet/intel/ice/ice.ko /opt/lib/modules/<kernel-version>/
RUN ls /opt/lib/modules/<kernel-version>

RUN depmod -b /opt <kernel-version>
```

Build and push source container to your registry. This might be your preconfigured external registry or internal
registry configured in previous steps.

```shell
$ podman build -t <your-registry>/openshift-kmm/kmm-ice-driver:<kernel-version> .
$ podman push <your-registry>/openshift-kmm/kmm-ice-driver:<kernel-version>
```

If you are getting the following error while pushing to internal registry

```text
Error: trying to reuse blob sha256:c9ac8ed59ad94403e08b349f8fda48ca4a120e90f550186208a4218662062577 at destination: checking whether a blob sha256:c9ac8ed59ad94403e08b349f8fda48ca4a120e90f550186208a4218662062577 exists in image-registry.openshift-image-registry.svc:5000/openshift-kmm/kmm-ice-driver: authentication required
```

create authentication file for `image-registry.openshift-image-registry.svc:5000`.

First create base64 encoded authentication string.

```shell
$ echo -n "kubeadmin:$(oc whoami -t)" | base64 -w 0
```

Then create `auth.json` file with following content

```json
{
  "auths": {
    "image-registry.openshift-image-registry.svc:5000": {
     "auth": "<base64_authentication_string>"
    }
  }
}
```

Push image again with following command

```shell
sudo podman push --authfile=auth.json image-registry.openshift-image-registry.svc:5000/openshift-kmm/kmm-ice-driver:5.14.0-284.25.1.el9_2.x86_64
```

### Unload in-tree ICE module from node

On some systems `irdma` module might be present and that module uses `ice` module as a dependency resulting in failure
of `ice` unloading procedure. Process of unloading `irdma` before `ice` could be performed manually or on boot using
systemd service.

> :warning: There is known issue in `irdma` module dependencies file that could cause unload of `i40e` driver together
> with `irdma` module. To avoid it always use `rmmod` instead of `modprobe -r` as `rmmod` does not unload dependencies.

If `irdma` module is loaded, unload it first.

Either manually.

```shell
$ sudo rmmod irdma
```

Or by using systemd (MachineConfig that will create respective systemd service on nodes)

```yaml
---
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  labels:
    machineconfiguration.openshift.io/role: worker # select role here
  name: unload-irdma-on-boot
spec:
  config:
    ignition:
      version: 3.2.0
    systemd:
      units:
        - contents: |
            [Unit]
            Description=irdma unload on boot
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
            ExecStart=/usr/bin/bash rmmod irdma
            StandardOutput=journal+console
            [Install]
            WantedBy=default.target
          enabled: true
          name: "irdma-driver-unload.service"
```

### Create KMM CR

Copy and edit the below CR resource before applying it, you can find possible values with annotations
[here](https://docs.openshift.com/container-platform/4.13/hardware_enablement/kmm-kernel-module-management.html).

```shell
$ vi kmm-module.yaml
```

* `selector` is the label for nodes you want the driver deployed on
* `regexp` is the regex which should match the kernel versions of nodes you want the driver deployed on
* `containerImage` is the image name as it appears in your registry
* `moduleName` is the name of your kernel module, it has to be ice for this module

```yaml
---
apiVersion: kmm.sigs.x-k8s.io/v1beta1
kind: Module
metadata:
  name: ice
  namespace: openshift-kmm
spec:
  moduleLoader:
    container:
      modprobe:
        moduleName: ice
      inTreeModuleToRemove: ice
      kernelMappings:
        - regexp: "<kernel-version>"
          containerImage: "<your-registry>/openshift-kmm/kmm-ice-driver:<kernel-version>"
  selector:
    node-role.kubernetes.io/worker: ""
```

> Note: To load different version of ice module, first old version of it needs to be unloaded. KMMO supports unloading
> of old module, but when the ModuleLoader pod is terminated, there is a limitation that the old module won't be loaded
> again.

Create the `Module` Resource

```shell
$ oc create -f kmm-module.yaml
```

Once the above KMMO CR is created, new pod/pods will appear.

```shell
oc get pods -n openshift-kmm
NAME                                              READY   STATUS    RESTARTS   AGE
ice-x4mcp-2tzkr                                   1/1     Running   0          6s
kmm-operator-controller-manager-9b546d464-ghv8v   2/2     Running   0          3h10m
```

You can now see the KMM manager logs and deployment of a DaemonSet targeting a node in the cluster.

```shell
$ oc logs -n openshift-kmm kmm-operator-controller-manager-549d9dbc84-f2rl
...
15:17:38.849520       1 module_reconciler.go:359] kmm "msg"="creating new driver container DS" "Module"={"name":"ice","namespace":"openshift-kmm"} "controller"="Module" "controllerGroup"="kmm.sigs.x-k8s.io" "controllerKind"="Module" "image"="image-registry.openshift-image-registry.svc:5000/openshift-kmm/kmm-ice-driver:5.14.0-284.25.1.el9_2.x86_64" "kernel version"="5.14.0-284.25.1.el9_2.x86_64" "name"="ice" "namespace"="openshift-kmm" "reconcileID"="78946ce7-6e73-46fc-9ee6-7b76adcfec6e" "version"=""
...
```

You can verify if ICE driver has been changed to OOT by executing

```shell
$ ethtool -i enp24s0 # name of any E810 NIC on the node, replace according to your environment
driver: ice
version: 1.12.7
firmware-version: 4.40 0x8001beb1 1.3492.0
expansion-rom-version:
bus-info: 0000:18:00.0
supports-statistics: yes
supports-test: yes
supports-eeprom-access: yes
supports-register-dump: yes
supports-priv-flags: yes
```
