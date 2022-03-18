FROM public.ecr.aws/bitnami/golang:latest

WORKDIR /app

ARG GO111MODULE=auto
RUN go install github.com/gojp/goreportcard@latest
RUN cd $GOPATH/src/github.com/gojp/goreportcard && (make install || cd / && curl -L https://git.io/vp6lP | sh)

RUN go install github.com/gojp/goreportcard/cmd/goreportcard-cli@latest

RUN go install golang.org/x/tools/cmd/goimports@latest

CMD $GOPATH/bin/goreportcard-cli -v -t 90
