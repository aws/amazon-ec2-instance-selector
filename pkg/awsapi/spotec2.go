package awsapi

import "github.com/aws/aws-sdk-go-v2/service/ec2"

type EC2Interface interface {
	ec2.DescribeSpotPriceHistoryAPIClient
}