{{/*
Expand the name of the chart.
*/}}
{{- define "kube-recall.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kube-recall.fullname" -}}
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
{{- define "kube-recall.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "kube-recall.annotations" -}}
meta.helm.sh/release-name: {{ .Release.Name }}
meta.helm.sh/release-namespace: {{ .Release.Namespace }}
{{- with .Values.annotations }}
{{ toYaml . | nindent 4 }}
{{- end }}
{{- end }}

{{/*
{{- with .Values.annotations }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
{{- end }}

{{/*
Common labels
*/}}
{{- define "kube-recall.labels" -}}
helm.sh/chart: {{ include "kube-recall.chart" . }}
{{ include "kube-recall.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
  
{{- end -}}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kube-recall.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-recall.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kube-recall.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kube-recall.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
