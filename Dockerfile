FROM golang:1.13-alpine as base
ENV GOOS=linux CGO_ENABLED=0 GOARCH=amd64
RUN apk --no-cache add git
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

## BUILDER STAGE ##
FROM base as builder
RUN go build -o conftest -ldflags="-w -s" main.go

## TEST STAGE ##
FROM base as test
RUN go test -v ./...

## ACCEPTANCE STAGE ##
FROM base as acceptance
COPY --from=builder /app/conftest /app/conftest

RUN apk add --update npm bash
RUN npm install -g bats

RUN bats acceptance.bats

## EXAMPLES STAGE ##
FROM base as examples
ENV TERRAFORM_VERSION=0.12.0-rc1 \
    KUSTOMIZE_VERSION=2.0.3

COPY --from=builder /app/conftest /usr/local/bin
COPY examples /examples

RUN apk add --update npm make git jq ca-certificates openssl unzip wget && \
    cd /tmp && \
    wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin

RUN wget -O /usr/local/bin/kustomize https://github.com/kubernetes-sigs/kustomize/releases/download/v${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64 && \
    chmod +x /usr/local/bin/kustomize

RUN go get -u cuelang.org/go/cmd/cue

WORKDIR /examples

## RELEASE ##
FROM alpine:latest
COPY --from=builder /app/conftest /
RUN ln -s /conftest /usr/local/bin/conftest
WORKDIR /project

ENTRYPOINT ["/conftest"]
CMD ["--help"]
