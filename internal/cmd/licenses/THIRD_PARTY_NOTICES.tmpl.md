# Third-Party Notices

This distribution includes third-party software. The dependencies below are compiled into the provider binary.
The list covers all supported platforms, so some entries are only part of the binaries of certain platforms.

## Go Runtime

The binary includes the Go standard library and runtime ({{ .GO_VERSION }}), licensed under BSD-3-Clause.

- License: https://github.com/golang/go/blob/{{ .GO_VERSION }}/LICENSE
- Patent grant: https://github.com/golang/go/blob/{{ .GO_VERSION }}/PATENTS

## Go Modules

| Package | Version | License | Details |
| --- | --- | --- | --- |
{{ .GO_LICENSE_REPORT }}
{{- define "ROWS" -}}
{{ `{{- range . }}` }}
| {{ `{{ .Name }}` }} | {{ `{{ .Version }}` }} | {{ `{{ .LicenseName }}` }} | https://pkg.go.dev/{{ `{{ .Name }}` }}@{{ `{{ .Version }}` }}?tab=licenses |
{{ `{{- end }}` }}
{{- end }}
