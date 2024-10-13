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

package selector_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/awsapi"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	describeInstanceTypes         = "DescribeInstanceTypes"
	describeInstanceTypeOfferings = "DescribeInstanceTypeOfferings"
	describeAvailabilityZones     = "DescribeAvailabilityZones"
	mockFilesPath                 = "../../test/static"
)

// Mocking helpers
type mockedEC2 struct {
	awsapi.SelectorInterface
	DescribeInstanceTypesResp           ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesRespFn         func(instanceType []ec2types.InstanceType) ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesErr            error
	DescribeInstanceTypeOfferingsRespFn func(zone string) ec2.DescribeInstanceTypeOfferingsOutput
	DescribeInstanceTypeOfferingsResp   ec2.DescribeInstanceTypeOfferingsOutput
	DescribeInstanceTypeOfferingsErr    error
	DescribeAvailabilityZonesResp       ec2.DescribeAvailabilityZonesOutput
	DescribeAvailabilityZonesErr        error
}

func (m mockedEC2) DescribeAvailabilityZones(ctx context.Context, input *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error) {
	return &m.DescribeAvailabilityZonesResp, m.DescribeAvailabilityZonesErr
}

func (m mockedEC2) DescribeInstanceTypes(ctx context.Context, input *ec2.DescribeInstanceTypesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceTypesOutput, error) {
	var response ec2.DescribeInstanceTypesOutput
	if m.DescribeInstanceTypesRespFn != nil {
		response = m.DescribeInstanceTypesRespFn(input.InstanceTypes)
	} else {
		response = m.DescribeInstanceTypesResp
	}

	return &response, m.DescribeInstanceTypesErr
}

func (m mockedEC2) DescribeInstanceTypeOfferings(ctx context.Context, input *ec2.DescribeInstanceTypeOfferingsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceTypeOfferingsOutput, error) {
	var response ec2.DescribeInstanceTypeOfferingsOutput
	if m.DescribeInstanceTypeOfferingsRespFn != nil {
		response = m.DescribeInstanceTypeOfferingsRespFn(input.Filters[0].Values[0])
	} else {
		response = m.DescribeInstanceTypeOfferingsResp
	}

	return &response, m.DescribeInstanceTypeOfferingsErr
}

func mockMultiRespDescribeInstanceTypesOfferings(t *testing.T, locationToFile map[string]string) mockedEC2 {
	api := describeInstanceTypeOfferings
	locationToResp := map[string]ec2.DescribeInstanceTypeOfferingsOutput{}
	for zone, file := range locationToFile {
		mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
		mockFile, err := os.ReadFile(mockFilename)
		h.Assert(t, err == nil, "Error reading mock file "+string(mockFilename))
		ditoo := ec2.DescribeInstanceTypeOfferingsOutput{}
		err = json.Unmarshal(mockFile, &ditoo)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		locationToResp[zone] = ditoo
	}
	return mockedEC2{
		DescribeInstanceTypeOfferingsRespFn: func(input string) ec2.DescribeInstanceTypeOfferingsOutput {
			resp := locationToResp[input]
			return resp
		},
	}
}

func setupMock(t *testing.T, api string, file string) mockedEC2 {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
	mockFile, err := os.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+string(mockFilename))
	switch api {
	case describeInstanceTypes:
		dito := ec2.DescribeInstanceTypesOutput{}
		err = json.Unmarshal(mockFile, &dito)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		return mockedEC2{
			DescribeInstanceTypesResp: dito,
		}
	case describeInstanceTypeOfferings:
		ditoo := ec2.DescribeInstanceTypeOfferingsOutput{}
		err = json.Unmarshal(mockFile, &ditoo)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		return mockedEC2{
			DescribeInstanceTypeOfferingsResp: ditoo,
		}
	case describeAvailabilityZones:
		dazo := ec2.DescribeAvailabilityZonesOutput{}
		err = json.Unmarshal(mockFile, &dazo)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		return mockedEC2{
			DescribeAvailabilityZonesResp: dazo,
		}
	default:
		h.Assert(t, false, "Unable to mock the provided API type "+api)
	}
	return mockedEC2{}
}

func getSelector(ec2Mock mockedEC2) selector.Selector {
	return selector.Selector{
		EC2:                   ec2Mock,
		EC2Pricing:            &ec2PricingMock{},
		InstanceTypesProvider: instancetypes.NewProvider("", "us-east-1", 0, ec2Mock),
	}
}

// Tests

func TestNew(t *testing.T) {
	ctx := context.Background()
	cfg, _ := config.LoadDefaultConfig(ctx)
	itf, _ := selector.New(ctx, cfg)
	h.Assert(t, itf != nil, "selector instance created without error")
}

func TestFilterVerbose(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	filters := selector.Filters{
		VCpusRange: &selector.Int32RangeFilter{LowerBound: 2, UpperBound: 2},
	}
	ctx := context.Background()
	results, err := itf.FilterVerbose(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	h.Assert(t, results[0].InstanceType == "t3.micro", "Should return t3.micro, got %s instead", results[0].InstanceType)
}

func TestFilterVerbose_NoResults(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	filters := selector.Filters{
		VCpusRange: &selector.Int32RangeFilter{LowerBound: 4, UpperBound: 4},
	}
	ctx := context.Background()
	results, err := itf.FilterVerbose(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance type with 4 vcpus")
}

func TestFilterVerbose_Failure(t *testing.T) {
	ctx := context.Background()
	itf := getSelector(mockedEC2{DescribeInstanceTypesErr: errors.New("error")})
	filters := selector.Filters{
		VCpusRange: &selector.Int32RangeFilter{LowerBound: 4, UpperBound: 4},
	}
	results, err := itf.FilterVerbose(ctx, filters)
	h.Assert(t, results == nil, "Results should be nil")
	h.Assert(t, err != nil, "An error should be returned")
}

func TestFilterVerbose_AZFilteredIn(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "t3_micro.json").DescribeInstanceTypesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	filters := selector.Filters{
		VCpusRange:        &selector.Int32RangeFilter{LowerBound: 2, UpperBound: 2},
		AvailabilityZones: &[]string{"us-east-2a"},
	}
	ctx := context.Background()
	results, err := itf.FilterVerbose(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	h.Assert(t, results[0].InstanceType == "t3.micro", "Should return t3.micro, got %s instead", results[0].InstanceType)
}

func TestFilterVerbose_AZFilteredOut(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "t3_micro.json").DescribeInstanceTypesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a_only_c5d12x.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	filters := selector.Filters{
		AvailabilityZones: &[]string{"us-east-2a"},
	}
	ctx := context.Background()
	results, err := itf.FilterVerbose(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance types in us-east-2a but actually returned "+strconv.Itoa(len(results)))
}

func TestFilterVerboseAZ_FilteredErr(t *testing.T) {
	itf := getSelector(mockedEC2{})
	filters := selector.Filters{
		VCpusRange:        &selector.Int32RangeFilter{LowerBound: 2, UpperBound: 2},
		AvailabilityZones: &[]string{"blah"},
	}
	ctx := context.Background()
	_, err := itf.FilterVerbose(ctx, filters)
	h.Assert(t, err != nil, "Should error since bad zone was passed in")
}

func TestFilterVerbose_Gpus(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro_and_p3_16xl.json"))
	gpuMemory, err := bytequantity.ParseToByteQuantity("128g")
	h.Ok(t, err)
	filters := selector.Filters{
		GpusRange: &selector.Int32RangeFilter{LowerBound: 8, UpperBound: 8},
		GpuMemoryRange: &selector.ByteQuantityRangeFilter{
			LowerBound: gpuMemory,
			UpperBound: gpuMemory,
		},
	}
	ctx := context.Background()
	results, err := itf.FilterVerbose(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	h.Assert(t, results[0].InstanceType == "p3.16xlarge", "Should return p3.16xlarge, got %s instead", results[0].InstanceType)
}

func TestFilter(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	filters := selector.Filters{
		VCpusRange: &selector.Int32RangeFilter{LowerBound: 2, UpperBound: 2},
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus")
	h.Assert(t, results[0] == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilter_MoreFilters(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	X8664Architecture := ec2types.ArchitectureTypeX8664
	NitroInstanceType := ec2types.InstanceTypeHypervisorNitro
	filters := selector.Filters{
		VCpusRange:      &selector.Int32RangeFilter{LowerBound: 2, UpperBound: 2},
		BareMetal:       aws.Bool(false),
		CPUArchitecture: &X8664Architecture,
		Hypervisor:      &NitroInstanceType,
		EnaSupport:      aws.Bool(true),
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus")
	h.Assert(t, results[0] == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilter_TruncateToMaxResults(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "25_instances.json"))
	filters := selector.Filters{
		VCpusRange: &selector.Int32RangeFilter{LowerBound: 0, UpperBound: 100},
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) > 1, "Should return > 1 instance types since max results is not set")

	filters = selector.Filters{
		VCpusRange: &selector.Int32RangeFilter{LowerBound: 0, UpperBound: 100},
		MaxResults: aws.Int(1),
	}
	results, err = itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should return 1 instance types since max results is set")

	filters = selector.Filters{
		VCpusRange: &selector.Int32RangeFilter{LowerBound: 0, UpperBound: 100},
		MaxResults: aws.Int(30),
	}
	results, err = itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 25, fmt.Sprintf("Should return 25 instance types since max results is set to 30 but only %d are returned in total", len(results)))
}

func TestFilter_Failure(t *testing.T) {
	itf := getSelector(mockedEC2{DescribeInstanceTypesErr: errors.New("error")})
	filters := selector.Filters{
		VCpusRange: &selector.Int32RangeFilter{LowerBound: 4, UpperBound: 4},
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Assert(t, results == nil, "Results should be nil")
	h.Assert(t, err != nil, "An error should be returned")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithZoneName(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	ctx := context.Background()
	results, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"us-east-2a"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in us-east-2a golden file w/ no resource filters applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithZoneID(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	ctx := context.Background()
	results, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"use2-az1"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in use2-az2 golden file w/ no resource filter applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithRegion(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	ctx := context.Background()
	results, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"us-east-2"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in us-east-2 golden file w/ no resource filter applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithBadZone(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	ctx := context.Background()
	results, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"blah"})
	h.Assert(t, err != nil, "Should return an error since a bad zone was passed in")
	h.Assert(t, results == nil, "Should return nil results due to error")
}

func TestRetrieveInstanceTypesSupportedInAZ_Error(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsErr: errors.New("error"),
		DescribeAvailabilityZonesResp:    setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	ctx := context.Background()
	results, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"us-east-2a"})
	h.Assert(t, err != nil, "Should return an error since ec2 api mock is configured to return an error")
	h.Assert(t, results == nil, "Should return nil results due to error")
}

func TestAggregateFilterTransform(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "g2_2xlarge.json"))
	g22Xlarge := "g2.2xlarge"
	filters := selector.Filters{
		InstanceTypeBase: &g22Xlarge,
	}
	ctx := context.Background()
	filters, err := itf.AggregateFilterTransform(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, filters.GpusRange != nil, "g2.2Xlarge as a base instance type should filter out non-GPU instances")
	h.Assert(t, *filters.BareMetal == false, "g2.2Xlarge as a base instance type should filter out bare metal instances")
	h.Assert(t, *filters.Fpga == false, "g2.2Xlarge as a base instance type should filter out FPGA instances")
	h.Assert(t, *filters.CPUArchitecture == "x86_64", "g2.2Xlarge as a base instance type should only return x86_64 instance types")
}

func TestAggregateFilterTransform_InvalidInstanceType(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "empty.json"))
	t3Micro := "t3.microoon"
	filters := selector.Filters{
		InstanceTypeBase: &t3Micro,
	}
	ctx := context.Background()
	_, err := itf.AggregateFilterTransform(ctx, filters)
	h.Nok(t, err)
}

func TestFilter_InstanceTypeBase(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesRespFn: func(instanceTypes []ec2types.InstanceType) ec2.DescribeInstanceTypesOutput {
			if len(instanceTypes) == 1 {
				return setupMock(t, describeInstanceTypes, "c4_large.json").DescribeInstanceTypesResp
			} else {
				return setupMock(t, describeInstanceTypes, "25_instances.json").DescribeInstanceTypesResp
			}
		},
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := getSelector(ec2Mock)
	c4Large := "c4.large"
	filters := selector.Filters{
		InstanceTypeBase: &c4Large,
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "c4.large should return 3 similar instance types")
}

func TestRetrieveInstanceTypesSupportedInAZs_Intersection(t *testing.T) {
	ec2Mock := mockMultiRespDescribeInstanceTypesOfferings(t, map[string]string{
		"us-east-2a": "us-east-2a.json",
		"us-east-2b": "us-east-2b.json",
	})
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	ctx := context.Background()
	results, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"us-east-2a", "us-east-2b"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "Should return instance types that are included in both files")

	// Check reversed zones to ensure order does not matter
	results, err = itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"us-east-2b", "us-east-2a"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "Should return instance types that are included in both files when passed in reverse order")
}

func TestRetrieveInstanceTypesSupportedInAZs_Duplicates(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2b.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	ctx := context.Background()
	results, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"us-east-2b", "us-east-2b"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "Should return instance types that are included in both files")
}

func TestRetrieveInstanceTypesSupportedInAZs_GoodAndBadZone(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	ctx := context.Background()
	_, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"us-weast-2k", "us-east-2a"})
	h.Nok(t, err)
}

func TestRetrieveInstanceTypesSupportedInAZs_DescribeAZErr(t *testing.T) {
	itf := getSelector(mockedEC2{DescribeAvailabilityZonesErr: fmt.Errorf("error")})
	ctx := context.Background()
	_, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, []string{"us-east-2a"})
	h.Nok(t, err)
}

func TestFilter_AllowList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "25_instances.json").DescribeInstanceTypesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := getSelector(ec2Mock)
	allowRegex, err := regexp.Compile("c4.large")
	h.Ok(t, err)
	filters := selector.Filters{
		AllowList: allowRegex,
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Allow List Regex: 'c4.large' should return 1 instance type")
}

func TestFilter_DenyList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "25_instances.json").DescribeInstanceTypesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := getSelector(ec2Mock)
	denyRegex, err := regexp.Compile("c4.large")
	h.Ok(t, err)
	filters := selector.Filters{
		DenyList: denyRegex,
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 24, "Deny List Regex: 'c4.large' should return 24 instance type matching regex but returned %d", len(results))
}

func TestFilter_AllowAndDenyList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "25_instances.json").DescribeInstanceTypesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := getSelector(ec2Mock)
	allowRegex, err := regexp.Compile("c4.*")
	h.Ok(t, err)
	denyRegex, err := regexp.Compile("c4.large")
	h.Ok(t, err)
	filters := selector.Filters{
		AllowList: allowRegex,
		DenyList:  denyRegex,
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 4, "Allow/Deny List Regex: 'c4.large' should return 4 instance types matching the regex but returned %d", len(results))
}

func TestFilter_X8664_AMD64(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	ArchitectureType := selector.ArchitectureTypeAMD64
	filters := selector.Filters{
		CPUArchitecture: &ArchitectureType,
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with x86_64/amd64 cpu architecture")
	h.Assert(t, results[0] == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilter_VirtType_PV(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "pv_instances.json"))
	pvType := selector.VirtualizationTypePv
	filters := selector.Filters{
		VirtualizationType: &pvType,
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) > 0, "Should return at least 1 instance type when filtering with VirtualizationType: pv")

	paravirtualType := ec2types.VirtualizationTypeParavirtual
	filters = selector.Filters{
		VirtualizationType: &paravirtualType,
	}
	results, err = itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) > 0, "Should return at least 1 instance type when filtering with VirtualizationType: paravirtual")
}

type ec2PricingMock struct {
	GetOndemandInstanceTypeCostResp    float64
	GetOndemandInstanceTypeCostErr     error
	GetSpotInstanceTypeNDayAvgCostResp float64
	GetSpotInstanceTypeNDayAvgCostErr  error
	RefreshOnDemandCacheErr            error
	RefreshSpotCacheErr                error
	onDemandCacheCount                 int
	spotCacheCount                     int
}

func (p *ec2PricingMock) GetOnDemandInstanceTypeCost(ctx context.Context, instanceType ec2types.InstanceType) (float64, error) {
	return p.GetOndemandInstanceTypeCostResp, p.GetOndemandInstanceTypeCostErr
}

func (p *ec2PricingMock) GetSpotInstanceTypeNDayAvgCost(ctx context.Context, instanceType ec2types.InstanceType, availabilityZones []string, days int) (float64, error) {
	return p.GetSpotInstanceTypeNDayAvgCostResp, p.GetSpotInstanceTypeNDayAvgCostErr
}

func (p *ec2PricingMock) RefreshOnDemandCache(ctx context.Context) error {
	return p.RefreshOnDemandCacheErr
}

func (p *ec2PricingMock) RefreshSpotCache(ctx context.Context, days int) error {
	return p.RefreshSpotCacheErr
}

func (p *ec2PricingMock) OnDemandCacheCount() int {
	return p.onDemandCacheCount
}

func (p *ec2PricingMock) SpotCacheCount() int {
	return p.spotCacheCount
}

func (p *ec2PricingMock) Save() error {
	return nil
}
func (p *ec2PricingMock) SetLogger(_ *log.Logger) {}

func TestFilter_PricePerHour(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	itf.EC2Pricing = &ec2PricingMock{
		GetOndemandInstanceTypeCostResp: 0.0104,
		onDemandCacheCount:              1,
	}
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0104,
			UpperBound: 0.0104,
		},
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, fmt.Sprintf("Should return 1 instance type; got %d", len(results)))
}

func TestFilter_PricePerHour_NoResults(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	itf.EC2Pricing = &ec2PricingMock{
		GetOndemandInstanceTypeCostResp: 0.0104,
		onDemandCacheCount:              1,
	}
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0105,
			UpperBound: 0.0105,
		},
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance types")
}

func TestFilter_PricePerHour_OD(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	itf.EC2Pricing = &ec2PricingMock{
		GetOndemandInstanceTypeCostResp: 0.0104,
		onDemandCacheCount:              1,
	}
	onDemandUsage := ec2types.UsageClassTypeOnDemand
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0104,
			UpperBound: 0.0104,
		},
		UsageClass: &onDemandUsage,
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, fmt.Sprintf("Should return 1 instance type; got %d", len(results)))
}

func TestFilter_PricePerHour_Spot(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "t3_micro.json"))
	itf.EC2Pricing = &ec2PricingMock{
		GetSpotInstanceTypeNDayAvgCostResp: 0.0104,
		spotCacheCount:                     1,
	}
	spotUsage := ec2types.UsageClassTypeSpot
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0104,
			UpperBound: 0.0104,
		},
		UsageClass: &spotUsage,
	}
	ctx := context.Background()
	results, err := itf.Filter(ctx, filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, fmt.Sprintf("Should return 1 instance type; got %d", len(results)))
}
