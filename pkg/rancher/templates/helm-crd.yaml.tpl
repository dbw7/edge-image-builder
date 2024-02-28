apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  repo: {{ .Repo }}
  chart: {{ .Chart }}
  targetNamespace: {{ .TargetNamespace }}
  createNamespace: {{ .CreateNamespace }}
  version: {{ .Version }}
  set:
    {{- range $key, $value := .Set }}
    {{$key}}: {{$value}}
    {{- end }}
  chartContent: {{ .ChartContent }}