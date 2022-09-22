// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package ec2pricing_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/awsapi"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/ec2pricing"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
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
	awsapi.PricingInterface
	awsapi.SelectorInterface
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
		priceList := []string{string(mockFile)}
		productsOutput := pricing.GetProductsOutput{
			PriceList: priceList,
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
	pricingMock := setupMock(t, getProductsPages, "m5_large.json")
	ec2pricingClient := ec2pricing.EC2Pricing{
		ODPricing: ec2pricing.LoadODCacheOrNew(pricingMock, "us-east-1", 0, ""),
	}
	price, err := ec2pricingClient.GetOnDemandInstanceTypeCost("m5.large")
	h.Ok(t, err)
	h.Equals(t, float64(0.096), price)
}

func TestRefreshOnDemandCache(t *testing.T) {
	pricingMock := setupMock(t, getProductsPages, "m5_large.json")
	ec2pricingClient := ec2pricing.EC2Pricing{
		ODPricing: ec2pricing.LoadODCacheOrNew(pricingMock, "us-east-1", 0, ""),
	}
	err := ec2pricingClient.RefreshOnDemandCache()
	h.Ok(t, err)

	price, err := ec2pricingClient.GetOnDemandInstanceTypeCost("m5.large")
	h.Ok(t, err)
	h.Equals(t, float64(0.096), price)
}

func TestGetSpotInstanceTypeNDayAvgCost(t *testing.T) {
	ec2Mock := setupMock(t, describeSpotPriceHistoryPages, "m5_large.json")
	ec2pricingClient := ec2pricing.EC2Pricing{
		SpotPricing: ec2pricing.LoadSpotCacheOrNew(ec2Mock, "us-east-1", 0, "", 30),
	}
	price, err := ec2pricingClient.GetSpotInstanceTypeNDayAvgCost("m5.large", []string{"us-east-1a"}, 30)
	h.Ok(t, err)
	h.Equals(t, float64(0.041486231229302666), price)
}

func TestRefreshSpotCache(t *testing.T) {
	ec2Mock := setupMock(t, describeSpotPriceHistoryPages, "m5_large.json")
	ec2pricingClient := ec2pricing.EC2Pricing{
		SpotPricing: ec2pricing.LoadSpotCacheOrNew(ec2Mock, "us-east-1", 0, "", 30),
	}
	err := ec2pricingClient.RefreshSpotCache(30)
	h.Ok(t, err)

	price, err := ec2pricingClient.GetSpotInstanceTypeNDayAvgCost("m5.large", []string{"us-east-1a"}, 30)
	h.Ok(t, err)
	h.Equals(t, float64(0.041486231229302666), price)
}
