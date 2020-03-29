# Build the manager binary
FROM golang:1.14 as builder

## GOLANG env
ARG GOPROXY="https://proxy.golang.org,direct"
ARG GO111MODULE="on"
ARG CGO_ENABLED=0
ARG GOOS=linux 
ARG GOARCH=amd64 

# Copy go.mod and download dependencies
WORKDIR /amazon-ec2-instance-selector
COPY go.mod .
COPY go.sum .
RUN go mod download

# Build
COPY . .
RUN make build
# In case the target is build for testing:
# $ docker build  --target=builder -t test .
CMD ["/amazon-ec2-instance-selector/build/ec2-instance-selector"]

# Copy the binary into a thin image
FROM amazonlinux:2 as amazonlinux
FROM scratch
WORKDIR /
COPY --from=builder /amazon-ec2-instance-selector/build/ec2-instance-selector .
COPY --from=amazonlinux /etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/
COPY THIRD_PARTY_LICENSES .
USER 1000
ENTRYPOINT ["/ec2-instance-selector"]
