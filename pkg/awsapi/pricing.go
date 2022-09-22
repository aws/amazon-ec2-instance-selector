package awsapi

import "github.com/aws/aws-sdk-go-v2/service/pricing"

type PricingInterface interface {
	pricing.GetProductsAPIClient
}
