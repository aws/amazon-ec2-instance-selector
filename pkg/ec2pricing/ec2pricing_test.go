package ec2pricing_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/ec2pricing"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/pricing/pricingiface"
)

const (
	getProductsPages              = "GetProductsPages"
	describeSpotPriceHistoryPages = "DescribeSpotPriceHistoryPages"
	mockFilesPath                 = "../../test/static"
)

// Mocking helpers

type gpFn = func(page *pricing.GetProductsOutput, lastPage bool) bool
type dspFn = func(page *ec2.DescribeSpotPriceHistoryOutput, lastPage bool) bool

type mockedPricing struct {
	pricingiface.PricingAPI
	ec2iface.EC2API
	GetProductsPagesResp              pricing.GetProductsOutput
	GetProductsPagesErr               error
	DescribeSpotPriceHistoryPagesResp ec2.DescribeSpotPriceHistoryOutput
	DescribeSpotPriceHistoryPagesErr  error
}

func (m mockedPricing) GetProductsPages(input *pricing.GetProductsInput, fn gpFn) error {
	fn(&m.GetProductsPagesResp, true)
	return m.GetProductsPagesErr
}

func (m mockedPricing) DescribeSpotPriceHistoryPages(input *ec2.DescribeSpotPriceHistoryInput, fn dspFn) error {
	fn(&m.DescribeSpotPriceHistoryPagesResp, true)
	return m.DescribeSpotPriceHistoryPagesErr
}

func setupMock(t *testing.T, api string, file string) mockedPricing {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
	mockFile, err := ioutil.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+string(mockFilename))
	switch api {
	case getProductsPages:
		var productsMap map[string]interface{}
		err = json.Unmarshal(mockFile, &productsMap)
		h.Assert(t, err == nil, "Error parsing mock json file contents "+mockFilename)
		productsOutput := pricing.GetProductsOutput{
			PriceList: []aws.JSONValue{productsMap},
		}
		return mockedPricing{
			GetProductsPagesResp: productsOutput,
		}
	case describeSpotPriceHistoryPages:
		dspho := ec2.DescribeSpotPriceHistoryOutput{}
		err = json.Unmarshal(mockFile, &dspho)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		return mockedPricing{
			DescribeSpotPriceHistoryPagesResp: dspho,
		}

	default:
		h.Assert(t, false, "Unable to mock the provided API type "+api)
	}
	return mockedPricing{}
}

func TestGetOndemandInstanceTypeCost_m5large(t *testing.T) {
	sess := session.Session{
		Config: &aws.Config{
			Region: aws.String("us-east-1"),
		},
	}
	pricingMock := setupMock(t, getProductsPages, "m5_large.json")
	ec2pricingClient := ec2pricing.EC2Pricing{
		PricingClient: pricingMock,
		AWSSession:    &sess,
	}
	price, err := ec2pricingClient.GetOndemandInstanceTypeCost("m5.large")
	h.Ok(t, err)
	h.Equals(t, float64(0.096), price)
}

func TestHydrateOndemandCache(t *testing.T) {
	sess := session.Session{
		Config: &aws.Config{
			Region: aws.String("us-east-1"),
		},
	}
	pricingMock := setupMock(t, getProductsPages, "m5_large.json")
	ec2pricingClient := ec2pricing.EC2Pricing{
		PricingClient: pricingMock,
		AWSSession:    &sess,
	}
	err := ec2pricingClient.HydrateOndemandCache()
	h.Ok(t, err)

	price, err := ec2pricingClient.GetOndemandInstanceTypeCost("m5.large")
	h.Ok(t, err)
	h.Equals(t, float64(0.096), price)
}

func TestGetSpotInstanceTypeNDayAvgCost(t *testing.T) {
	sess := session.Session{
		Config: &aws.Config{
			Region: aws.String("us-east-1"),
		},
	}
	ec2Mock := setupMock(t, describeSpotPriceHistoryPages, "m5_large.json")
	ec2pricingClient := ec2pricing.EC2Pricing{
		EC2Client:  ec2Mock,
		AWSSession: &sess,
	}
	price, err := ec2pricingClient.GetSpotInstanceTypeNDayAvgCost("m5.large", []string{"us-east-1a"}, 30)
	h.Ok(t, err)
	h.Equals(t, float64(0.041486231229302666), price)
}

func TestHydrateSpotCache(t *testing.T) {
	sess := session.Session{
		Config: &aws.Config{
			Region: aws.String("us-east-1"),
		},
	}
	ec2Mock := setupMock(t, describeSpotPriceHistoryPages, "m5_large.json")
	ec2pricingClient := ec2pricing.EC2Pricing{
		EC2Client:  ec2Mock,
		AWSSession: &sess,
	}
	err := ec2pricingClient.HydrateSpotCache(30)
	h.Ok(t, err)

	price, err := ec2pricingClient.GetSpotInstanceTypeNDayAvgCost("m5.large", []string{"us-east-1a"}, 30)
	h.Ok(t, err)
	h.Equals(t, float64(0.041486231229302666), price)
}
