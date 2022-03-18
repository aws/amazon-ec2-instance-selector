FROM public.ecr.aws/bitnami/golang:1.17

WORKDIR /app

COPY license-config.hcl .
ARG GOPROXY="https://proxy.golang.org,direct"
RUN GO111MODULE=on go install github.com/mitchellh/golicense@v0.2.0

CMD $GOPATH/bin/golicense 
