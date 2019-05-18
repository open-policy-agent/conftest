FROM golang:1.12-alpine as builder
RUN apk --no-cache add git
WORKDIR /
COPY . /
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"


FROM bats/bats:v1.1.0 as acceptance
COPY --from=builder /conftest /usr/local/bin
COPY acceptance.bats /acceptance.bats
COPY examples /examples
RUN ./acceptance.bats


FROM golang:1.12-alpine as examples

ENV TERRAFORM_VERSION=0.12.0-rc1 \
    KUSTOMIZE_VERSION=2.0.3

COPY --from=builder /conftest /usr/local/bin
COPY examples /examples

RUN apk add --update npm make git jq ca-certificates openssl unzip wget && \
    cd /tmp && \
    wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin

RUN wget -O /usr/local/bin/kustomize https://github.com/kubernetes-sigs/kustomize/releases/download/v${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64 && \
    chmod +x /usr/local/bin/kustomize

RUN go get -u cuelang.org/go/cmd/cue

WORKDIR /examples


FROM alpine:latest
COPY --from=builder /conftest /
RUN ln -s /conftest /usr/local/bin/conftest
WORKDIR /project
ENTRYPOINT ["/conftest"]
CMD ["--help"]
