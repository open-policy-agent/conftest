# METADATA
# title: My package foo
# description: the package with rule A and subpackage bar
# scope: subpackages
# organizations:
# - Acme Corp.
package foo
import rego.v1

# METADATA
# title: My Rule A
# description: the rule A = 3
# related_resources:
# - ref: https://example.com
# - ref: https://example.com/more
#   description: Yet another link
a := 3
