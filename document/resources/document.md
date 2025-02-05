{{ range . -}}
{{ .MarkdownHeading }} {{ .RegoPackageName }} - {{ .Annotations.Title }}

{{ .Annotations.Description }}
{{ if .Annotations.RelatedResources }}
Related Resources:
{{ range .Annotations.RelatedResources }}
{{ if .Description -}}
* [{{.Description}}]({{ .Ref }})
{{- else -}}
* {{ .Ref }}
{{- end -}}
{{ end }}
{{ end }}
{{ end -}}
