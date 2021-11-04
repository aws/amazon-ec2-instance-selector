FROM public.ecr.aws/bitnami/golang:latest

WORKDIR /app

COPY license-config.hcl .
ARG GOPROXY="https://proxy.golang.org,direct"
RUN GO111MODULE=on go get github.com/mitchellh/golicense

CMD $GOPATH/bin/golicense 
