# Generate Policy Documentations

## Document your policies

OPA has introduced a standard way to document policies called [Metadata](https://www.openpolicyagent.org/docs/latest/policy-language/#metadata). 
This format allows for structured in code documentation of policies.

```opa
# METADATA
# title: My rule
# description: A rule that determines if x is allowed.
# authors:
# - John Doe <john@example.com>
# entrypoint: true
allow if {
  ...
}
```

For the generated documentation to make sense your `packages` should be documented with at least the `title` field 
and `rules` should have both `title` and `description`. This will ensure that no section is empty in your 
documentations.

## Generate the documentation

In code documentation is great but what we often want it to later generated an actual static reference documentation.
The `doc` command will retrieve all annotation of a targeted module and generate a markdown documentation for it.

```bash
conftest doc path/to/policy
```

## Use your own template

You can override the [default template](../document/resources/document.md) with your own template

```aiignore
conftest -t template.md path/tp/policies
```

All annotation are returned as a sorted list of all annotations, grouped by the path and location of their targeted 
package or rule. For instance using this template

```bash
{{ range . -}}
{{ .Path }} has annotations {{ .Annotations }}
{{ end -}}
```

for the following module

```yaml
# METADATA
# scope: subpackages
# organizations:
# - Acme Corp.
package foo
---
# METADATA
# description: A couple of useful rules
package foo.bar

# METADATA
# title: My Rule P
p := 7
```

You will obtain the following rendered documentation:

```bash
data.foo has annotations {"organizations":["Acme Corp."],"scope":"subpackages"}
data.foo.bar has annotations {"description":"A couple of useful rules","scope":"package"}
data.foo.bar.p has annotations {"scope":"rule","title":"My Rule P"}
```
