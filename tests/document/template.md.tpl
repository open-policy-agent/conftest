{{ range . -}}
{{ .Path }} has annotations {{ .Annotations }}
{{ end -}}
