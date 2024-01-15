package main

import future.keywords.contains
import future.keywords.if

deny contains msg if {
    msg := "foo"
}
