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

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
)

// Tests
var emr = "emr"

func TestEMRDefaultService(t *testing.T) {
	registry := selector.NewRegistry()
	registry.Register("emr", &selector.EMR{})

	filters := selector.Filters{
		Service: &emr,
	}

	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, *transformedFilters.RootDeviceType == "ebs", "emr should only supports ebs")
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "emr should only support hvm")

	emrWithVersion := "emr-" + "5.20.0"
	filters.Service = &emrWithVersion
	transformedFilters, err = registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, *transformedFilters.RootDeviceType == "ebs", "emr should only supports ebs")
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "emr should only support hvm")
}

func TestFilters_Version5_25_0(t *testing.T) {
	registry := selector.NewRegistry()
	registry.Register("emr", &selector.EMR{})

	filters := selector.Filters{
		Service: &emr,
	}

	emrWithVersion := "emr-" + "5.25.0"
	filters.Service = &emrWithVersion
	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, *transformedFilters.RootDeviceType == "ebs", "emr should only supports ebs")
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "emr should only support hvm")
	h.Assert(t, contains(*transformedFilters.InstanceTypes, "i3en.xlarge"), "emr version 5.25.0 should include i3en.xlarge")
}

func TestFilters_Version5_15_0(t *testing.T) {
	registry := selector.NewRegistry()
	registry.Register("emr", &selector.EMR{})

	filters := selector.Filters{
		Service: &emr,
	}

	emrWithVersion := "emr-" + "5.15.0"
	filters.Service = &emrWithVersion
	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, *transformedFilters.RootDeviceType == "ebs", "emr should only supports ebs")
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "emr should only support hvm")
	h.Assert(t, !contains(*transformedFilters.InstanceTypes, "c1.medium"), "emr version 5.15.0 should not include c1.medium")
}

func TestFilters_Version5_13_0(t *testing.T) {
	registry := selector.NewRegistry()
	registry.Register("emr", &selector.EMR{})

	filters := selector.Filters{
		Service: &emr,
	}

	emrWithVersion := "emr-" + "5.13.0"
	filters.Service = &emrWithVersion
	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, *transformedFilters.RootDeviceType == "ebs", "emr should only supports ebs")
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "emr should only support hvm")
	h.Assert(t, !contains(*transformedFilters.InstanceTypes, "m5a.xlarge"), "emr version 5.13.0 should not include m5a.xlarge")
}

func TestFilters_Version5_9_0(t *testing.T) {
	registry := selector.NewRegistry()
	registry.Register("emr", &selector.EMR{})

	filters := selector.Filters{
		Service: &emr,
	}

	emrWithVersion := "emr-" + "5.9.0"
	filters.Service = &emrWithVersion
	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, *transformedFilters.RootDeviceType == "ebs", "emr should only supports ebs")
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "emr should only support hvm")
	h.Assert(t, !contains(*transformedFilters.InstanceTypes, "m5a.xlarge"), "emr version 5.9.0 should not include m5a.xlarge")
}

func TestFilters_Version5_8_0(t *testing.T) {
	registry := selector.NewRegistry()
	registry.Register("emr", &selector.EMR{})

	filters := selector.Filters{
		Service: &emr,
	}

	emrWithVersion := "emr-" + "5.8.0"
	filters.Service = &emrWithVersion
	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, *transformedFilters.RootDeviceType == "ebs", "emr should only supports ebs")
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "emr should only support hvm")
	h.Assert(t, !contains(*transformedFilters.InstanceTypes, "i3.xlarge"), "emr version 5.8.0 should not include i3.xlarge")
}

func contains(arr []string, input string) bool {
	for _, entry := range arr {
		if entry == input {
			return true
		}
	}
	return false
}
