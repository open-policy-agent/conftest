FROM golang:1.11-alpine as builder
RUN apk --no-cache add git
RUN mkdir -p /go/src/github.com/instrumenta/conftest/
COPY . /go/src/github.com/instrumenta/conftest/
WORKDIR /go/src/github.com/instrumenta/conftest/
RUN go build .

FROM bats/bats:v1.1.0 as acceptance
COPY --from=builder /go/src/github.com/instrumenta/conftest/conftest /usr/local/bin
COPY acceptance.bats /acceptance.bats
COPY testdata /testdata
RUN ./acceptance.bats
