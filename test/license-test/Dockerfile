FROM public.ecr.aws/bitnami/golang:latest

WORKDIR /app

COPY license-config.hcl .
ARG GOPROXY="https://proxy.golang.org,direct"
RUN GO111MODULE=on go install github.com/mitchellh/golicense@latest

CMD $GOPATH/bin/golicense 
