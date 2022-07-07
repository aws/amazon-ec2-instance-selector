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
	"strings"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector/outputs"
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
			resp := locationToResp[input]
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

func getSelector(ec2Mock mockedEC2) selector.Selector {
	return selector.Selector{
		EC2:                   ec2Mock,
		EC2Pricing:            &ec2PricingMock{},
		InstanceTypesProvider: instancetypes.NewProvider("", "us-east-1", 0, ec2Mock),
	}
}

// Tests

func TestNew(t *testing.T) {
	itf := selector.New(session.Must(session.NewSession()))
	h.Assert(t, itf != nil, "selector instance created without error")
}

func TestFilterInstanceTypes(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
	filter := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
	}

	results, err := itf.FilterInstanceTypes(filter)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 intance type with 2 vcpus")
	instanceTypeName := results[0].InstanceType
	h.Assert(t, instanceTypeName != nil, "Instance type name should not be nil")
	h.Assert(t, *instanceTypeName == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilterInstanceTypes_NoResults(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
	filters := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 4, UpperBound: 4},
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance type with 4 vcpus")
}

func TestFilterInstanceTypes_AZFilteredIn(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "t3_micro.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	filters := selector.Filters{
		VCpusRange:        &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
		AvailabilityZones: &[]string{"us-east-2a"},
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	instanceTypeName := results[0].InstanceType
	h.Assert(t, instanceTypeName != nil, "Instance type name should not be nil")
	h.Assert(t, *instanceTypeName == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilterInstanceTypes_AZFilteredOut(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "t3_micro.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a_only_c5d12x.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	filters := selector.Filters{
		AvailabilityZones: &[]string{"us-east-2a"},
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance types in us-east-2a but actually returned "+strconv.Itoa(len(results)))
}

func TestFilterInstanceTypes_AZFilteredErr(t *testing.T) {
	itf := getSelector(mockedEC2{})
	filters := selector.Filters{
		VCpusRange:        &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
		AvailabilityZones: &[]string{"blah"},
	}

	_, err := itf.FilterInstanceTypes(filters)

	h.Assert(t, err != nil, "Should error since bad zone was passed in")
}

func TestFilterInstanceTypes_Gpus(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro_and_p3_16xl.json"))
	gpuMemory, err := bytequantity.ParseToByteQuantity("128g")
	h.Ok(t, err)
	filters := selector.Filters{
		GpusRange: &selector.IntRangeFilter{LowerBound: 8, UpperBound: 8},
		GpuMemoryRange: &selector.ByteQuantityRangeFilter{
			LowerBound: gpuMemory,
			UpperBound: gpuMemory,
		},
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	instanceTypeName := results[0].InstanceType
	h.Assert(t, instanceTypeName != nil, "Instance type name should not be nil")
	h.Assert(t, *instanceTypeName == "p3.16xlarge", "Should return p3.16xlarge, got %s instead", *results[0].InstanceType)
}

func TestFilterInstanceTypes_MoreFilters(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
	filters := selector.Filters{
		VCpusRange:      &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
		BareMetal:       aws.Bool(false),
		CPUArchitecture: aws.String("x86_64"),
		Hypervisor:      aws.String("nitro"),
		EnaSupport:      aws.Bool(true),
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus")
	instanceTypeName := results[0].InstanceType
	h.Assert(t, instanceTypeName != nil, "Instance type name should not be nil")
	h.Assert(t, *instanceTypeName == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilterInstanceTypes_Failure(t *testing.T) {
	itf := getSelector(mockedEC2{DescribeInstanceTypesPagesErr: errors.New("error")})
	filters := selector.Filters{
		VCpusRange: &selector.IntRangeFilter{LowerBound: 4, UpperBound: 4},
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Assert(t, results == nil, "Results should be nil")
	h.Assert(t, err != nil, "An error should be returned")
}

func TestFilterInstanceTypes_InstanceTypeBase(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "c4_large.json").DescribeInstanceTypesResp,
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := getSelector(ec2Mock)
	c4Large := "c4.large"
	filters := selector.Filters{
		InstanceTypeBase: &c4Large,
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "c4.large should return 3 similar instance types")
}

func TestFilterInstanceTypes_AllowList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := getSelector(ec2Mock)
	allowRegex, err := regexp.Compile("c4.large")
	h.Ok(t, err)
	filters := selector.Filters{
		AllowList: allowRegex,
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Allow List Regex: 'c4.large' should return 1 instance type")
}

func TestFilterInstanceTypes_DenyList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := getSelector(ec2Mock)
	denyRegex, err := regexp.Compile("c4.large")
	h.Ok(t, err)
	filters := selector.Filters{
		DenyList: denyRegex,
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 24, "Deny List Regex: 'c4.large' should return 24 instance type matching regex but returned %d", len(results))
}

func TestFilterInstanceTypes_AllowAndDenyList(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
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

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 4, "Allow/Deny List Regex: 'c4.large' should return 4 instance types matching the regex but returned %d", len(results))
}

func TestFilterInstanceTypes_X8664_AMD64(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
	filters := selector.Filters{
		CPUArchitecture: aws.String("amd64"),
	}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with x86_64/amd64 cpu architecture")
	instanceTypeName := results[0].InstanceType
	h.Assert(t, instanceTypeName != nil, "Instance type name should not be nil")
	h.Assert(t, *instanceTypeName == "t3.micro", "Should return t3.micro, got %s instead", results[0])

}

func TestFilterInstanceTypes_VirtType_PV(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "pv_instances.json"))
	filters := selector.Filters{
		VirtualizationType: aws.String("pv"),
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) > 0, "Should return at least 1 instance type when filtering with VirtualizationType: pv")

	filters = selector.Filters{
		VirtualizationType: aws.String("paravirtual"),
	}

	results, err = itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) > 0, "Should return at least 1 instance type when filtering with VirtualizationType: paravirtual")
}

func TestFilterInstanceTypes_PricePerHour(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
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

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, fmt.Sprintf("Should return 1 instance type; got %d", len(results)))
}

func TestFilterInstanceTypes_PricePerHour_NoResults(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
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

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 0, "Should return 0 instance types")
}

func TestFilterInstanceTypes_PricePerHour_OD(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
	itf.EC2Pricing = &ec2PricingMock{
		GetOndemandInstanceTypeCostResp: 0.0104,
		onDemandCacheCount:              1,
	}
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0104,
			UpperBound: 0.0104,
		},
		UsageClass: aws.String("on-demand"),
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, fmt.Sprintf("Should return 1 instance type; got %d", len(results)))
}

func TestFilterInstanceTypes_PricePerHour_Spot(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
	itf.EC2Pricing = &ec2PricingMock{
		GetSpotInstanceTypeNDayAvgCostResp: 0.0104,
		spotCacheCount:                     1,
	}
	filters := selector.Filters{
		PricePerHour: &selector.Float64RangeFilter{
			LowerBound: 0.0104,
			UpperBound: 0.0104,
		},
		UsageClass: aws.String("spot"),
	}

	results, err := itf.FilterInstanceTypes(filters)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, fmt.Sprintf("Should return 1 instance type; got %d", len(results)))
}

// checkSortResults is a helper function for comparing the results of sorting tests. Returns true if
// the order of instance types in the instanceTypes list matches the the order of instance type names
// in the expectedResullt list, and returns false otherwise.
func checkSortResults(instaceTypes []*instancetypes.Details, expectedResult []string) bool {
	if len(instaceTypes) != len(expectedResult) {
		return false
	}

	for i := 0; i < len(instaceTypes); i++ {
		actualName := instaceTypes[i].InstanceTypeInfo.InstanceType
		expectedName := expectedResult[i]

		if actualName == nil || *actualName != expectedName {
			return false
		}
	}

	return true
}

func TestSortInstanceTypes_OneElement(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "t3_micro.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	sortFilter := selector.NameSortFlag
	sortDirection := selector.SortAscendingFlag
	results, err = itf.SortInstanceTypes(results, &sortFilter, &sortDirection)

	h.Ok(t, err)
	h.Assert(t, len(results) == 1, fmt.Sprintf("Should only return one element, but it returned: %d elements", len(results)))
}

func TestSortInstanceTypes_Emptylist(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "empty.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	sortFilter := selector.NameSortFlag
	sortDirection := selector.SortAscendingFlag
	results, err = itf.SortInstanceTypes(results, &sortFilter, &sortDirection)

	h.Ok(t, err)
	h.Assert(t, len(results) == 0, fmt.Sprintf("Should return empty list, but it returned: %d elements", len(results)))
}

func TestSortInstanceTypes_Name(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "3_instances.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	// test ascending
	sortFilter := selector.NameSortFlag
	sortDirection := selector.SortAscendingFlag
	sortedResults, err := itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults := []string{"a1.2xlarge", "a1.4xlarge", "a1.large"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))

	// test descending
	sortDirection = selector.SortDescendingFlag
	sortedResults, err = itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults = []string{"a1.large", "a1.4xlarge", "a1.2xlarge"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))
}

func TestSortInstanceTypes_Memory(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "3_instances.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	// test ascending
	sortFilter := selector.MemorySortFlag
	sortDirection := selector.SortAscendingFlag
	sortedResults, err := itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults := []string{"a1.large", "a1.2xlarge", "a1.4xlarge"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))

	// test descending
	sortDirection = selector.SortDescendingFlag
	sortedResults, err = itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults = []string{"a1.4xlarge", "a1.2xlarge", "a1.large"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))
}

func TestSortInstanceTypes_Vcpu(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "3_instances.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	// test ascending
	sortFilter := selector.VcpuSortFlag
	sortDirection := selector.SortAscendingFlag
	sortedResults, err := itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults := []string{"a1.large", "a1.2xlarge", "a1.4xlarge"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))

	// test descending
	sortDirection = selector.SortDescendingFlag
	sortedResults, err = itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults = []string{"a1.4xlarge", "a1.2xlarge", "a1.large"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))
}

func TestSortInstanceTypes_SpotPrice(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "3_instances.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	// add spot prices
	prices := make(map[string]*float64)
	price1, price2, price3 := 1.0, 2.0, 3.0
	prices["a1.large"] = &price1
	prices["a1.2xlarge"] = &price2
	prices["a1.4xlarge"] = &price3
	for _, v := range results {
		name := v.InstanceTypeInfo.InstanceType
		v.SpotPrice = prices[*name]
	}

	// test ascending
	sortFilter := selector.SpotPriceSortFlag
	sortDirection := selector.SortAscendingFlag
	sortedResults, err := itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults := []string{"a1.large", "a1.2xlarge", "a1.4xlarge"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))

	// test descending
	sortDirection = selector.SortDescendingFlag
	sortedResults, err = itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults = []string{"a1.4xlarge", "a1.2xlarge", "a1.large"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))
}

func TestSortInstanceTypes_OnDemandPrice(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "3_instances.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	// add on demand prices
	prices := make(map[string]*float64)
	price1, price2, price3 := 1.0, 2.0, 3.0
	prices["a1.large"] = &price1
	prices["a1.2xlarge"] = &price2
	prices["a1.4xlarge"] = &price3
	for _, v := range results {
		name := v.InstanceTypeInfo.InstanceType
		v.OndemandPricePerHour = prices[*name]
	}

	// sort ascending
	sortFilter := selector.ODPriceSortFlag
	sortDirection := selector.SortAscendingFlag
	sortedResults, err := itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults := []string{"a1.large", "a1.2xlarge", "a1.4xlarge"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))

	// sort descending
	sortDirection = selector.SortDescendingFlag
	sortedResults, err = itf.SortInstanceTypes(results, &sortFilter, &sortDirection)
	expectedResults = []string{"a1.4xlarge", "a1.2xlarge", "a1.large"}

	h.Ok(t, err)
	h.Assert(t, len(sortedResults) == len(results), fmt.Sprintf("Should return %d elements, but returned %d", len(results), len(sortedResults)))
	h.Assert(t, checkSortResults(sortedResults, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedResults)))
}

func TestSortInstanceTypes_InvalidFilter(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "3_instances.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	sortFilter := "blah blah blah"
	sortDirection := selector.SortAscendingFlag
	sortedResults, err := itf.SortInstanceTypes(results, &sortFilter, &sortDirection)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, sortedResults == nil, fmt.Sprintf("Returned instance types details should be nil, but instead got: %s", outputs.OneLineOutput(sortedResults)))
}

func TestSortInstanceTypes_InvalidDirection(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypesPages, "3_instances.json"))
	filters := selector.Filters{}
	results, err := itf.FilterInstanceTypes(filters)
	h.Ok(t, err)

	sortFilter := selector.NameSortFlag
	sortDirection := "fdsfds"
	sortedResults, err := itf.SortInstanceTypes(results, &sortFilter, &sortDirection)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, sortedResults == nil, fmt.Sprintf("Returned instance types details should be nil, but instead got: %s", outputs.OneLineOutput(sortedResults)))
}

func TestRetrieveInstanceTypesSupportedInAZ_WithZoneName(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2a"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in us-east-2a golden file w/ no resource filters applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithZoneID(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"use2-az1"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in use2-az2 golden file w/ no resource filter applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithRegion(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in us-east-2 golden file w/ no resource filter applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithBadZone(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"blah"})
	h.Assert(t, err != nil, "Should return an error since a bad zone was passed in")
	h.Assert(t, results == nil, "Should return nil results due to error")
}

func TestRetrieveInstanceTypesSupportedInAZ_Error(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsErr: errors.New("error"),
		DescribeAvailabilityZonesResp:    setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2a"})
	h.Assert(t, err != nil, "Should return an error since ec2 api mock is configured to return an error")
	h.Assert(t, results == nil, "Should return nil results due to error")
}

func TestAggregateFilterTransform(t *testing.T) {
	itf := getSelector(setupMock(t, describeInstanceTypes, "g2_2xlarge.json"))
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
	itf := getSelector(setupMock(t, describeInstanceTypes, "empty.json"))
	t3Micro := "t3.microoon"
	filters := selector.Filters{
		InstanceTypeBase: &t3Micro,
	}
	_, err := itf.AggregateFilterTransform(filters)
	h.Nok(t, err)
}

func TestRetrieveInstanceTypesSupportedInAZs_Intersection(t *testing.T) {
	ec2Mock := mockMultiRespDescribeInstanceTypesOfferings(t, map[string]string{
		"us-east-2a": "us-east-2a.json",
		"us-east-2b": "us-east-2b.json",
	})
	ec2Mock.DescribeAvailabilityZonesResp = setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp
	itf := getSelector(ec2Mock)
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
	itf := getSelector(ec2Mock)
	results, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2b", "us-east-2b"})
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "Should return instance types that are included in both files")
}

func TestRetrieveInstanceTypesSupportedInAZs_GoodAndBadZone(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
		DescribeAvailabilityZonesResp:     setupMock(t, describeAvailabilityZones, "us-east-2.json").DescribeAvailabilityZonesResp,
	}
	itf := getSelector(ec2Mock)
	_, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-weast-2k", "us-east-2a"})
	h.Nok(t, err)
}

func TestRetrieveInstanceTypesSupportedInAZs_DescribeAZErr(t *testing.T) {
	itf := getSelector(mockedEC2{DescribeAvailabilityZonesErr: fmt.Errorf("error")})
	_, err := itf.RetrieveInstanceTypesSupportedInLocations([]string{"us-east-2a"})
	h.Nok(t, err)
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

func (p *ec2PricingMock) GetOnDemandInstanceTypeCost(instanceType string) (float64, error) {
	return p.GetOndemandInstanceTypeCostResp, p.GetOndemandInstanceTypeCostErr
}

func (p *ec2PricingMock) GetSpotInstanceTypeNDayAvgCost(instanceType string, availabilityZones []string, days int) (float64, error) {
	return p.GetSpotInstanceTypeNDayAvgCostResp, p.GetSpotInstanceTypeNDayAvgCostErr
}

func (p *ec2PricingMock) RefreshOnDemandCache() error {
	return p.RefreshOnDemandCacheErr
}

func (p *ec2PricingMock) RefreshSpotCache(days int) error {
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
