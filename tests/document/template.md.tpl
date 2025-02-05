{{ range . -}}
{{ .RegoPackageName }} has annotations {{ .Annotations }}
{{ end -}}
