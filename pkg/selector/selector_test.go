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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

const (
	describeInstanceTypesPages    = "DescribeInstanceTypesPages"
	describeInstanceTypes         = "DescribeInstanceTypes"
	describeInstanceTypeOfferings = "DescribeInstanceTypeOfferings"
	describeAvailabilityZones     = "DescribeAvailabilityZones"
	mockFilesPath                 = "../../test/static"
)

// Mocking helpers

type itFn = func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool
type ioFn = func(page *ec2.DescribeInstanceTypeOfferingsOutput, lastPage bool) bool

type mockedEC2 struct {
	ec2iface.EC2API
	DescribeInstanceTypesPagesResp      ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesPagesErr       error
	DescribeInstanceTypesResp           ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesErr            error
	DescribeInstanceTypeOfferingsRespFn func(zone string) *ec2.DescribeInstanceTypeOfferingsOutput
	DescribeInstanceTypeOfferingsResp   ec2.DescribeInstanceTypeOfferingsOutput
	DescribeInstanceTypeOfferingsErr    error
	DescribeAvailabilityZonesResp       ec2.DescribeAvailabilityZonesOutput
	DescribeAvailabilityZonesErr        error
}

func (m mockedEC2) DescribeAvailabilityZones(input *ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error) {
	return &m.DescribeAvailabilityZonesResp, m.DescribeAvailabilityZonesErr
}

func (m mockedEC2) DescribeInstanceTypes(input *ec2.DescribeInstanceTypesInput) (*ec2.DescribeInstanceTypesOutput, error) {
	return &m.DescribeInstanceTypesResp, m.DescribeInstanceTypesErr
}

func (m mockedEC2) DescribeInstanceTypesPages(input *ec2.DescribeInstanceTypesInput, fn itFn) error {
	fn(&m.DescribeInstanceTypesPagesResp, true)
	return m.DescribeInstanceTypesPagesErr
}

func (m mockedEC2) DescribeInstanceTypeOfferingsPages(input *ec2.DescribeInstanceTypeOfferingsInput, fn ioFn) error {
	if m.DescribeInstanceTypeOfferingsRespFn != nil {
		fn(m.DescribeInstanceTypeOfferingsRespFn(*input.Filters[0].Values[0]), true)
	} else {
		fn(&m.DescribeInstanceTypeOfferingsResp, true)
	}
	return m.DescribeInstanceTypeOfferingsErr
}

func mockMultiRespDescribeInstanceTypesOfferings(t *testing.T, locationToFile map[string]string) mockedEC2 {
	api := describeInstanceTypeOfferings
	locationToResp := map[string]ec2.DescribeInstanceTypeOfferingsOutput{}
	for zone, file := range locationToFile {
		mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
		mockFile, err := ioutil.ReadFile(mockFilename)
		h.Assert(t, err == nil, "Error reading mock file "+string(mockFilename))
		ditoo := ec2.DescribeInstanceTypeOfferingsOutput{}
		err = json.Unmarshal(mockFile, &ditoo)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		locationToResp[zone] = ditoo
	}
	return mockedEC2{
		DescribeInstanceTypeOfferingsRespFn: func(input string) *ec2.DescribeInstanceTypeOfferingsOutput {
			resp, _ := locationToResp[input]
			return &resp
		},
	}
}

func setupMock(t *testing.T, api string, file string) mockedEC2 {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
	mockFile, err := ioutil.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+string(mockFilename))
	switch api {
	case describeInstanceTypes:
		dito := ec2.DescribeInstanceTypesOutput{}
		err = json.Unmarshal(mockFile, &dito)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		return mockedEC2{
			DescribeInstanceTypesResp: dito,
		}
	case describeInstanceTypesPages:
		dito := ec2.DescribeInstanceTypesOutput{}
		err = json.Unmarshal(mockFile, &dito)
		h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
		return mockedEC2{
			DescribeInstanceTypesPagesResp: dito,
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

// Tests

func TestNew(t *testing.T) {
	itf := selector.New(session.Must(session.NewSession()))
	h.Assert(t, itf != nil, "selector instance created without error")
}

func TestFilterVerbose(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
	}
	results, err := itf.FilterVerbose(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	h.Assert(t, *results[0].InstanceType == "t3.micro", "Should return t3.micro, got %s instead", results[0].InstanceType)
}

func TestFilterVerbose_NoResults(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 4, UpperBound: 4},
	}
	results, err := itf.FilterVerbose(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance type with 4 vcpus")
}

func TestFilterVerbose_Failure(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesErr: errors.New("error"),
	}
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	filters := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 4, UpperBound: 4},
	}
	results, err := itf.FilterVerbose(filters)
	h.Assert(t, results == nil, "Results should be nil")
	h.Assert(t, err != nil, "An error should be returned")
}

func TestFilterVerbose_AZFilteredIn(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "t3_micro.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		VCpusRange:        &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
		AvailabilityZones: &[]string{"us-east-2a"},
	}
	results, err := itf.FilterVerbose(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	h.Assert(t, *results[0].InstanceType == "t3.micro", "Should return t3.micro, got %s instead", results[0].InstanceType)
}

func TestFilterVerbose_AZFilteredOut(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "t3_micro.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a_only_c5d12x.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		AvailabilityZones: &[]string{"us-east-2a"},
	}
	results, err := itf.FilterVerbose(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance types in us-east-2a but actually returned "+strconv.Itoa(len(results)))
}

func TestFilterVerboseAZ_FilteredErr(t *testing.T) {
	ec2Mock := mockedEC2{}
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	filters := selector.Filters{
		VCpusRange:        &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
		AvailabilityZones: &[]string{"blah"},
	}
	_, err := itf.FilterVerbose(filters)
	h.Assert(t, err != nil, "Should error since bad zone was passed in")
}

func TestFilterVerbose_Gpus(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro_and_p3_16xl.json")
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	gpuMemory, err := bytequantity.ParseToByteQuantity("128g")
	h.Ok(t, err)
	filters := selector.Filters{
		GpusRange: &selector.IntRangeFilter{LowerBound: 8, UpperBound: 8},
		GpuMemoryRange: &selector.ByteQuantityRangeFilter{
			LowerBound: gpuMemory,
			UpperBound: gpuMemory,
		},
	}
	results, err := itf.FilterVerbose(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	h.Assert(t, *results[0].InstanceType == "p3.16xlarge", "Should return p3.16xlarge, got %s instead", *results[0].InstanceType)
}

func TestFilter(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus")
	h.Assert(t, results[0] == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilter_MoreFilters(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		VCpusRange:      &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
		BareMetal:       aws.Bool(false),
		CPUArchitecture: aws.String("x86_64"),
		Hypervisor:      aws.String("nitro"),
		EnaSupport:      aws.Bool(true),
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus")
	h.Assert(t, results[0] == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilter_TruncateToMaxResults(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "25_instances.json")
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 0, UpperBound: 100},
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) > 1, "Should return > 1 instance types since max results is not set")

	filters = selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 0, UpperBound: 100},
		MaxResults: aws.Int(1),
	}
	results, err = itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should return 1 instance types since max results is set")

	filters = selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 0, UpperBound: 100},
		MaxResults: aws.Int(30),
	}
	results, err = itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 25, "Should return 25 instance types since max results is set to 30 but only 25 are returned in total")
}

func TestFilter_Failure(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesErr: errors.New("error"),
	}

	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 4, UpperBound: 4},
	}
	results, err := itf.Filter(filters)
	h.Assert(t, results == nil, "Results should be nil")
	h.Assert(t, err != nil, "An error should be returned")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithZoneName(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2a"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in us-east-2a golden file w/ no resource filters applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithZoneID(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"use2-az1"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in use2-az2 golden file w/ no resource filter applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithRegion(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in us-east-2 golden file w/ no resource filter applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithBadZone(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"blah"})
	h.Assert(t, err != nil, "Should return an error since a bad zone was passed in")
	h.Assert(t, results == nil, "Should return nil results due to error")
}

func TestRetrieveInstanceTypesSupportedInAZ_Error(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsErr: errors.New("error"),
		DescribeAvailabilityZonesResp:    setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2a"})
	h.Assert(t, err != nil, "Should return an error since ec2 api mock is configured to return an error")
	h.Assert(t, results == nil, "Should return nil results due to error")
}

func TestAggregateFilterTransform(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypes, "g2_2xlarge.json")
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	g22Xlarge := "g2.2xlarge"
	filters := selector.Filters{
		InstanceTypeBase: &g22Xlarge,
	}
	filters, err := itf.AggregateFilterTransform(filters)
	h.Ok(t, err)
	h.Assert(t, filters.GpusRange != nil, "g2.2Xlarge as a base instance type should filter out non-GPU instances")
	h.Assert(t, *filters.BareMetal == false, "g2.2Xlarge as a base instance type should filter out bare metal instances")
	h.Assert(t, *filters.Fpga == false, "g2.2Xlarge as a base instance type should filter out FPGA instances")
	h.Assert(t, *filters.CPUArchitecture == "x86_64", "g2.2Xlarge as a base instance type should only return x86_64 instance types")
}

func TestAggregateFilterTransform_InvalidInstanceType(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypes, "empty.json")
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	t3Micro := "t3.microoon"
	filters := selector.Filters{
		InstanceTypeBase: &t3Micro,
	}
	_, err := itf.AggregateFilterTransform(filters)
	h.Nok(t, err)
}

func TestFilter_InstanceTypeBase(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "c4_large.json").DescribeInstanceTypesResp,
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	c4Large := "c4.large"
	filters := selector.Filters{
		InstanceTypeBase: &c4Large,
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "c4.large should return 3 similar instance types")
}

func TestRetrieveInstanceTypesSupportedInAZs_Intersection(t *testing.T) {
	ec2Mock := mockMultiRespDescribeInstanceTypesOfferings(t, map[string]string{
		"us-east-2a": "us-east-2a.json",
		"us-east-2b": "us-east-2b.json",
	})
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2a", "us-east-2b"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "Should return instance types that are included in both files")

	// Check reversed zones to ensure order does not matter
	results, err = itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2b", "us-east-2a"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "Should return instance types that are included in both files when passed in reverse order")
}

func TestRetrieveInstanceTypesSupportedInAZs_Duplicates(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2b.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2b", "us-east-2b"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "Should return instance types that are included in both files")
}

func TestRetrieveInstanceTypesSupportedInAZs_GoodAndBadZone(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	_, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-weast-2k", "us-east-2a"})
	h.Nok(t, err)
}

func TestRetrieveInstanceTypesSupportedInAZs_DescribeAZErr(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeAvailabilityZonesErr: fmt.Errorf("error"),
	}

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	_, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2a"})
	h.Nok(t, err)
}

func TestFilter_AllowList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	allowRegex, err := regexp.Compile("c4.large")
	h.Ok(t, err)
	filters := selector.Filters{
		AllowList: allowRegex,
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Allow List Regex: 'c4.large' should return 1 instance type")
}

func TestFilter_DenyList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	denyRegex, err := regexp.Compile("c4.large")
	h.Ok(t, err)
	filters := selector.Filters{
		DenyList: denyRegex,
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 24, "Deny List Regex: 'c4.large' should return 24 instance type matching regex but returned %d", len(results))
}

func TestFilter_AllowAndDenyList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	allowRegex, err := regexp.Compile("c4.*")
	h.Ok(t, err)
	denyRegex, err := regexp.Compile("c4.large")
	h.Ok(t, err)
	filters := selector.Filters{
		AllowList: allowRegex,
		DenyList:  denyRegex,
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 4, "Allow/Deny List Regex: 'c4.large' should return 4 instance types matching the regex but returned %d", len(results))
}

func TestFilter_X8664_AMD64(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		CPUArchitecture: aws.String("amd64"),
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with x86_64/amd64 cpu architecture")
	h.Assert(t, results[0] == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilter_VirtType_PV(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "pv_instances.json")
	itf := selector.Selector{
		EC2:        ec2Mock,
		EC2Pricing: &ec2PricingMock{},
	}
	filters := selector.Filters{
		VirtualizationType: aws.String("pv"),
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) > 0, "Should return at least 1 instance type when filtering with VirtualizationType: pv")

	filters = selector.Filters{
		VirtualizationType: aws.String("paravirtual"),
	}
	results, err = itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) > 0, "Should return at least 1 instance type when filtering with VirtualizationType: paravirtual")
}

type ec2PricingMock struct {
	GetOndemandInstanceTypeCostResp    float64
	GetOndemandInstanceTypeCostErr     error
	GetSpotInstanceTypeNDayAvgCostResp float64
	GetSpotInstanceTypeNDayAvgCostErr  error
	HydrateOndemandCacheErr            error
	HydrateSpotCacheErr                error
	lastOnDemandCacheUTC               *time.Time
	lastSpotCacheUTC                   *time.Time
}

func (p *ec2PricingMock) GetOndemandInstanceTypeCost(instanceType string) (float64, error) {
	return p.GetOndemandInstanceTypeCostResp, p.GetOndemandInstanceTypeCostErr
}

func (p *ec2PricingMock) GetSpotInstanceTypeNDayAvgCost(instanceType string, availabilityZones []string, days int) (float64, error) {
	return p.GetSpotInstanceTypeNDayAvgCostResp, p.GetSpotInstanceTypeNDayAvgCostErr
}

func (p *ec2PricingMock) HydrateOndemandCache() error {
	return p.HydrateOndemandCacheErr
}

func (p *ec2PricingMock) HydrateSpotCache(days int) error {
	return p.HydrateSpotCacheErr
}

func (p *ec2PricingMock) LastOnDemandCacheUTC() *time.Time {
	return p.lastOnDemandCacheUTC
}

func (p *ec2PricingMock) LastSpotCacheUTC() *time.Time {
	return p.lastSpotCacheUTC
}

func TestFilter_PricePerHour(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2: ec2Mock,
		EC2Pricing: &ec2PricingMock{
			GetOndemandInstanceTypeCostResp: 0.0104,
		},
	}
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0104,
			UpperBound: 0.0104,
		},
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should return 1 instance type")
}

func TestFilter_PricePerHour_NoResults(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2: ec2Mock,
		EC2Pricing: &ec2PricingMock{
			GetOndemandInstanceTypeCostResp: 0.0104,
		},
	}
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0105,
			UpperBound: 0.0105,
		},
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance types")
}

func TestFilter_PricePerHour_OD(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2: ec2Mock,
		EC2Pricing: &ec2PricingMock{
			GetOndemandInstanceTypeCostResp: 0.0104,
		},
	}
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0104,
			UpperBound: 0.0104,
		},
		UsageClass: aws.String("on-demand"),
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should return 1 instance type")
}

func TestFilter_PricePerHour_Spot(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2: ec2Mock,
		EC2Pricing: &ec2PricingMock{
			GetSpotInstanceTypeNDayAvgCostResp: 0.0104,
		},
	}
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0104,
			UpperBound: 0.0104,
		},
		UsageClass: aws.String("spot"),
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should return 1 instance type")
}
