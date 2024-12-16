// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ec2pricing_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/samber/lo"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/ec2pricing"
	h "github.com/aws/amazon-ec2-instance-selector/v3/pkg/test"
)

const (
	getProducts              = "GetProducts"
	describeSpotPriceHistory = "DescribeSpotPriceHistory"
	mockFilesPath            = "../../test/static"
)

// Mocking helpers

type mockedPricing struct {
	pricing.GetProductsAPIClient
	GetProductsResp pricing.GetProductsOutput
	GetProductsErr  error
}

func (m mockedPricing) GetProducts(_ context.Context, input *pricing.GetProductsInput, optFns ...func(*pricing.Options)) (*pricing.GetProductsOutput, error) {
	return &m.GetProductsResp, m.GetProductsErr
}

type mockedSpotEC2 struct {
	ec2.DescribeSpotPriceHistoryAPIClient
	DescribeSpotPriceHistoryPagesResp ec2.DescribeSpotPriceHistoryOutput
	DescribeSpotPriceHistoryPagesErr  error
}

func (m mockedSpotEC2) DescribeSpotPriceHistory(_ context.Context, input *ec2.DescribeSpotPriceHistoryInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSpotPriceHistoryOutput, error) {
	return &m.DescribeSpotPriceHistoryPagesResp, m.DescribeSpotPriceHistoryPagesErr
}

func setupOdMock(t *testing.T, api string, file string) mockedPricing {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
	mockFile, err := os.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+mockFilename)
	switch api {
	case getProducts:
		priceList := []string{string(mockFile)}
		productsOutput := pricing.GetProductsOutput{
			PriceList: priceList,
		}
		return mockedPricing{
			GetProductsResp: productsOutput,
		}

	default:
		h.Assert(t, false, "Unable to mock the provided API type "+api)
	}
	return mockedPricing{}
}

func setupEc2Mock(t *testing.T, api string, file string) mockedSpotEC2 {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
	mockFile, err := os.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+mockFilename)
	switch api {
	case describeSpotPriceHistory:
		dspho := ec2.DescribeSpotPriceHistoryOutput{}
		err = json.Unmarshal(mockFile, &dspho)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		return mockedSpotEC2{
			DescribeSpotPriceHistoryPagesResp: dspho,
		}

	default:
		h.Assert(t, false, "Unable to mock the provided API type "+api)
	}
	return mockedSpotEC2{}
}

func TestGetOndemandInstanceTypeCost_m5large(t *testing.T) {
	pricingMock := setupOdMock(t, getProducts, "m5_large.json")
	ctx := context.Background()
	ec2pricingClient := ec2pricing.EC2Pricing{
		ODPricing: lo.Must(ec2pricing.LoadODCacheOrNew(ctx, pricingMock, "us-east-1", 0, "")),
	}
	price, err := ec2pricingClient.GetOnDemandInstanceTypeCost(ctx, ec2types.InstanceTypeM5Large)
	h.Ok(t, err)
	h.Equals(t, float64(0.096), price)
}

func TestRefreshOnDemandCache(t *testing.T) {
	pricingMock := setupOdMock(t, getProducts, "m5_large.json")
	ctx := context.Background()
	ec2pricingClient := ec2pricing.EC2Pricing{
		ODPricing: lo.Must(ec2pricing.LoadODCacheOrNew(ctx, pricingMock, "us-east-1", 0, "")),
	}
	err := ec2pricingClient.RefreshOnDemandCache(ctx)
	h.Ok(t, err)

	price, err := ec2pricingClient.GetOnDemandInstanceTypeCost(ctx, ec2types.InstanceTypeM5Large)
	h.Ok(t, err)
	h.Equals(t, float64(0.096), price)
}

func TestGetSpotInstanceTypeNDayAvgCost(t *testing.T) {
	ec2Mock := setupEc2Mock(t, describeSpotPriceHistory, "m5_large.json")
	ctx := context.Background()
	ec2pricingClient := ec2pricing.EC2Pricing{
		SpotPricing: lo.Must(ec2pricing.LoadSpotCacheOrNew(ctx, ec2Mock, "us-east-1", 0, "", 30)),
	}
	price, err := ec2pricingClient.GetSpotInstanceTypeNDayAvgCost(ctx, ec2types.InstanceTypeM5Large, []string{"us-east-1a"}, 30)
	h.Ok(t, err)
	h.Equals(t, float64(0.041486231229302666), price)
}

func TestRefreshSpotCache(t *testing.T) {
	ec2Mock := setupEc2Mock(t, describeSpotPriceHistory, "m5_large.json")
	ctx := context.Background()
	ec2pricingClient := ec2pricing.EC2Pricing{
		SpotPricing: lo.Must(ec2pricing.LoadSpotCacheOrNew(ctx, ec2Mock, "us-east-1", 0, "", 30)),
	}
	err := ec2pricingClient.RefreshSpotCache(ctx, 30)
	h.Ok(t, err)

	price, err := ec2pricingClient.GetSpotInstanceTypeNDayAvgCost(ctx, ec2types.InstanceTypeM5Large, []string{"us-east-1a"}, 30)
	h.Ok(t, err)
	h.Equals(t, float64(0.041486231229302666), price)
}
