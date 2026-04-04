{{- define "inventory-back.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "inventory-back.fullname" -}}
{{- printf "%s" (include "inventory-back.name" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "inventory-back.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "inventory-back.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "inventory-back.selectorLabels" -}}
app.kubernetes.io/name: {{ include "inventory-back.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
