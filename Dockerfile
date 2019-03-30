FROM golang:1.12-alpine as builder
RUN apk --no-cache add git
WORKDIR /
COPY . /
RUN go build .

FROM bats/bats:v1.1.0 as acceptance
COPY --from=builder /conftest /usr/local/bin
COPY acceptance.bats /acceptance.bats
COPY testdata /testdata
RUN ./acceptance.bats
