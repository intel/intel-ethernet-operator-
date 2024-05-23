{{/*
Expand the name of the chart.
*/}}
{{- define "intel-ethernet-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "intel-ethernet-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "intel-ethernet-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "intel-ethernet-operator.labels" -}}
helm.sh/chart: {{ include "intel-ethernet-operator.chart" . }}
{{ include "intel-ethernet-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "intel-ethernet-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "intel-ethernet-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
control-plane: controller-manager
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "intel-ethernet-operator.serviceAccount" -}}
{{- printf "%s-controller-manager" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the webhook service to use
*/}}
{{- define "intel-ethernet-operator.webhookService" -}}
{{- printf "%s-webhook-service" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the webhook tls certificates secret to use
*/}}
{{- define "intel-ethernet-operator.webhookCertsSecret" -}}
{{- printf "%s-webhook-server-certs" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the webhook mtls config to use
*/}}
{{- define "intel-ethernet-operator.mTLSWebhookConfig" -}}
{{- printf "%s-webhook-mtls-config" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the webhook mtls ca secret to use
*/}}
{{- define "intel-ethernet-operator.mTLSWebhookCASecret" -}}
{{- printf "%s-webhook-mtls-ca-secret" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the cluster role to use
*/}}
{{- define "intel-ethernet-operator.clusterRole" -}}
{{- printf "%s-manager-cluster-role" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the cluster role binding to use
*/}}
{{- define "intel-ethernet-operator.clusterRoleBinding" -}}
{{- printf "%s-manager-cluster-role-binding" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the manager config map to use
*/}}
{{- define "intel-ethernet-operator.managerConfigMap" -}}
{{- printf "%s-manager-config" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the role to use
*/}}
{{- define "intel-ethernet-operator.role" -}}
{{- printf "%s-manager-leader-election-role" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the role binding to use
*/}}
{{- define "intel-ethernet-operator.roleBinding" -}}
{{- printf "%s-manager-leader-election-role-binding" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the validating webhook configuration to use
*/}}
{{- define "intel-ethernet-operator.validatingWebhookConfiguration" -}}
{{- printf "%s-validating-webhook-configuration" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the cert manager certificate to use
*/}}
{{- define "intel-ethernet-operator.certificate" -}}
{{- printf "%s-serving-cert" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the cert manager self signed issuer to use
*/}}
{{- define "intel-ethernet-operator.selfSignedIssuer" -}}
{{- printf "%s-self-signed-issuer" (include "intel-ethernet-operator.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}
