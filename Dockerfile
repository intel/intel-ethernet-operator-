# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020-2024 Intel Corporation

# Build the manager binary
FROM docker.io/golang:alpine3.19 as builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY main.go main.go
COPY apis/ apis/

COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
    go build -trimpath -mod=readonly -gcflags="all=-spectre=all -N -l" \
    -asmflags="all=-spectre=all" -ldflags="all=-s -w" -a -o \
    manager main.go

FROM docker.io/redhat/ubi9-micro:9.3-15

ARG VERSION
### Required OpenShift Labels
LABEL name="Intel Ethernet Operator" \
    vendor="Intel Corporation" \
    version=$VERSION \
    release="1" \
    summary="Intel Ethernet Operator for E810 NICs" \
    description="The role of the Intel Ethernet Operator is to orchestrate and manage the configuration of the capabilities exposed by the Intel E810 Series network interface cards (NICs)"

USER 1001
WORKDIR /
COPY --from=builder /workspace/manager .
COPY assets/ assets/

USER root
RUN mkdir /licenses
COPY LICENSE /licenses/LICENSE.txt
RUN chown 1001 /licenses
USER 1001

ENTRYPOINT ["/manager"]
