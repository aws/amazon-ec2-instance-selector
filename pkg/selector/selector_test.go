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
	"log"
	"strconv"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/pkg/test"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

const (
	describeInstanceTypesPages    = "DescribeInstanceTypesPages"
	describeInstanceTypes         = "DescribeInstanceTypes"
	describeInstanceTypeOfferings = "DescribeInstanceTypeOfferings"
	mockFilesPath                 = "../../test/static"
)

// Mocking helpers

type itFn = func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool
type ioFn = func(page *ec2.DescribeInstanceTypeOfferingsOutput, lastPage bool) bool

type mockedEC2 struct {
	ec2iface.EC2API
	DescribeInstanceTypesPagesResp    ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesPagesErr     error
	DescribeInstanceTypesResp         ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesErr          error
	DescribeInstanceTypeOfferingsResp ec2.DescribeInstanceTypeOfferingsOutput
	DescribeInstanceTypeOfferingsErr  error
}

func (m mockedEC2) DescribeInstanceTypes(input *ec2.DescribeInstanceTypesInput) (*ec2.DescribeInstanceTypesOutput, error) {
	return &m.DescribeInstanceTypesResp, m.DescribeInstanceTypesErr
}

func (m mockedEC2) DescribeInstanceTypesPages(input *ec2.DescribeInstanceTypesInput, fn itFn) error {
	fn(&m.DescribeInstanceTypesPagesResp, true)
	return m.DescribeInstanceTypesPagesErr
}

func (m mockedEC2) DescribeInstanceTypeOfferingsPages(input *ec2.DescribeInstanceTypeOfferingsInput, fn ioFn) error {
	fn(&m.DescribeInstanceTypeOfferingsResp, true)
	return m.DescribeInstanceTypeOfferingsErr
}

// Tests

func TestNew(t *testing.T) {
	itf := selector.New(session.Must(session.NewSession()))
	h.Assert(t, itf != nil, "selector instance created without error")
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
	default:
		h.Assert(t, false, "Unable to mock the provided API type "+api)
	}
	return mockedEC2{}
}

func TestFilterVerbose(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2: ec2Mock,
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
		EC2: ec2Mock,
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
	}
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	filters := selector.Filters{
		VCpusRange:       &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
		AvailabilityZone: aws.String("us-east-2a"),
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
	}
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	filters := selector.Filters{
		AvailabilityZone: aws.String("us-east-2a"),
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
		VCpusRange:       &selector.IntRangeFilter{LowerBound: 2, UpperBound: 2},
		AvailabilityZone: aws.String("blah"),
	}
	_, err := itf.FilterVerbose(filters)
	h.Assert(t, err != nil, "Should error since bad zone was passed in")
}

func TestFilterVerbose_Gpus(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro_and_p3_16xl.json")
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	filters := selector.Filters{
		GpusRange:      &selector.IntRangeFilter{LowerBound: 8, UpperBound: 8},
		GpuMemoryRange: &selector.IntRangeFilter{LowerBound: 131072, UpperBound: 131072},
	}
	results, err := itf.FilterVerbose(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus but actually returned "+strconv.Itoa(len(results)))
	h.Assert(t, *results[0].InstanceType == "p3.16xlarge", "Should return p3.16xlarge, got %s instead", *results[0].InstanceType)
}

func TestFilter(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "t3_micro.json")
	itf := selector.Selector{
		EC2: ec2Mock,
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
		EC2: ec2Mock,
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
	log.Println(results)
	h.Assert(t, len(results) == 1, "Should only return 1 instance type with 2 vcpus")
	h.Assert(t, results[0] == "t3.micro", "Should return t3.micro, got %s instead", results[0])
}

func TestFilter_TruncateToMaxResults(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypesPages, "25_instances.json")
	itf := selector.Selector{
		EC2: ec2Mock,
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
		EC2: ec2Mock,
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
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocation("us-east-2a")
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in us-east-2a golden file w/ no resource filters applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithZoneID(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocation("use2-az1")
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in use2-az2 golden file w/ no resource filter applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithRegion(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocation("us-east-2")
	h.Ok(t, err)
	h.Assert(t, len(results) == 228, "Should return 228 entries in us-east-2 golden file w/ no resource filter applied")
}

func TestRetrieveInstanceTypesSupportedInAZ_WithBadZone(t *testing.T) {
	ec2Mock := setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json")
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocation("blah")
	h.Assert(t, err != nil, "Should return an error since a bad zone was passed in")
	h.Assert(t, results == nil, "Should return nil results due to error")
}

func TestRetrieveInstanceTypesSupportedInAZ_Error(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypeOfferingsErr: errors.New("error"),
	}

	itf := selector.Selector{
		EC2: ec2Mock,
	}
	results, err := itf.RetrieveInstanceTypesSupportedInLocation("us-east-2a")
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
	filters, err := itf.AggregateFilterTransform(filters, 0.8, 1.2)
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
	_, err := itf.AggregateFilterTransform(filters, 0.8, 1.2)
	h.Nok(t, err)
}

func TestFilter_InstanceTypeBase(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "c4_large.json").DescribeInstanceTypesResp,
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	c4Large := "c4.large"
	filters := selector.Filters{
		InstanceTypeBase: &c4Large,
	}
	results, err := itf.Filter(filters)
	h.Ok(t, err)
	h.Assert(t, len(results) == 3, "c4.large should return 3 similar instance types")
}
