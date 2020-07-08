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
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/pkg/test"
)

// Tests

func TestTransformBaseInstanceType(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "c4_large.json").DescribeInstanceTypesResp,
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "25_instances.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	instanceTypeBase := "c4.large"
	filters := selector.Filters{
		InstanceTypeBase: &instanceTypeBase,
	}
	filters, err := itf.TransformBaseInstanceType(filters)
	h.Ok(t, err)
	h.Assert(t, *filters.BareMetal == false, " should filter out bare metal instances")
	h.Assert(t, *filters.Fpga == false, "should filter out FPGA instances")
	h.Assert(t, *filters.CPUArchitecture == "x86_64", "should only return x86_64 instance types")
	h.Assert(t, filters.GpusRange.LowerBound == 0 && filters.GpusRange.UpperBound == 0, "should only return non-gpu instance types")
}

func TestTransformBaseInstanceTypeWithGPU(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesResp:         setupMock(t, describeInstanceTypes, "g2_2xlarge.json").DescribeInstanceTypesResp,
		DescribeInstanceTypesPagesResp:    setupMock(t, describeInstanceTypesPages, "g2_2xlarge_group.json").DescribeInstanceTypesPagesResp,
		DescribeInstanceTypeOfferingsResp: setupMock(t, describeInstanceTypeOfferings, "us-east-2a.json").DescribeInstanceTypeOfferingsResp,
	}
	itf := selector.Selector{
		EC2: ec2Mock,
	}
	instanceTypeBase := "g2.2xlarge"
	filters := selector.Filters{
		InstanceTypeBase: &instanceTypeBase,
	}
	filters, err := itf.TransformBaseInstanceType(filters)
	h.Ok(t, err)
	h.Assert(t, *filters.BareMetal == false, " should filter out bare metal instances")
	h.Assert(t, *filters.Fpga == false, "should filter out FPGA instances")
	h.Assert(t, *filters.CPUArchitecture == "x86_64", "should only return x86_64 instance types")
	h.Assert(t, filters.GpusRange.LowerBound == 1 && filters.GpusRange.UpperBound == 1, "should only return gpu instance types")
}

func TestTransformFamilyFlexibile(t *testing.T) {
	itf := selector.Selector{}
	flexible := true
	filters := selector.Filters{
		Flexible: &flexible,
	}
	filters, err := itf.TransformFlexible(filters)
	h.Ok(t, err)
	h.Assert(t, *filters.BareMetal == false, " should filter out bare metal instances")
	h.Assert(t, *filters.Fpga == false, "should filter out FPGA instances")
	h.Assert(t, *filters.CPUArchitecture == "x86_64", "should only return x86_64 instance types")
}
