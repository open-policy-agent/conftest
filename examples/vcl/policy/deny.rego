package main

deny[msg] {
    acl := input.acl.purge[_] 
    not acl = "127.0.0.1"
    msg := sprintf("acl purge should be 127.0.0.1 got %s", [acl])
}

deny[msg] {
    app := input.backend.app
    not app.port = "8080"
    msg := "default backend port should be 8080"
}