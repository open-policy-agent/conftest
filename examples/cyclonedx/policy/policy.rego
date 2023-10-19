package main

deny[msg] {
    expected_shas256 := "sha256:d7ec60cf8390612b360c857688b383068b580d9a6ab78417c9493170ad3f1616"
    input.metadata.component.version != expected_shas256
    msg := sprintf(
        "current SHA256 %s is not equal to expected SHA256 %s", [input.metadata.component.version, expected_shas256]
    )
}