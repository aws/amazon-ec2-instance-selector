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
	"github.com/aws/aws-sdk-go-v2/aws"
)

// Tests

func TestDefaultRegistry(t *testing.T) {
	registry := selector.NewRegistry()
	registry.RegisterAWSServices()

	emr := "emr"
	filters := selector.Filters{
		Service: &emr,
	}

	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
}

func TestRegister_LazyInit(t *testing.T) {
	registry := selector.ServiceRegistry{}
	registry.RegisterAWSServices()

	emr := "emr"
	filters := selector.Filters{
		Service: &emr,
	}

	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
}

func TestExecuteTransforms_OnUnrecognizedService(t *testing.T) {
	registry := selector.NewRegistry()
	registry.RegisterAWSServices()

	nes := "nonexistentservice"
	filters := selector.Filters{
		Service: &nes,
	}

	_, err := registry.ExecuteTransforms(filters)
	h.Nok(t, err)
}

func TestRegister_CustomService(t *testing.T) {
	registry := selector.NewRegistry()
	customServiceFn := func(version string) (filters selector.Filters, err error) {
		filters.BareMetal = aws.Bool(true)
		return filters, nil
	}

	registry.Register("myservice", selector.ServiceFiltersFn(customServiceFn))

	myService := "myservice"
	filters := selector.Filters{
		Service: &myService,
	}

	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, *transformedFilters.BareMetal == true, "custom service should have transformed BareMetal to true")
}

func TestExecuteTransforms_ShortCircuitOnEmptyService(t *testing.T) {
	registry := selector.NewRegistry()
	registry.RegisterAWSServices()

	emr := ""
	filters := selector.Filters{
		Service: &emr,
	}

	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters == filters, " Filters should not be modified")
}

func TestExecuteTransforms_ValidVersionParsing(t *testing.T) {
	registry := selector.NewRegistry()
	customServiceFn := func(version string) (filters selector.Filters, err error) {
		h.Assert(t, version == "myversion", "version should have been parsed as myversion but got %s", version)
		return filters, nil
	}

	registry.Register("myservice", selector.ServiceFiltersFn(customServiceFn))

	myService := "myservice-myversion"
	filters := selector.Filters{
		Service: &myService,
	}

	_, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
}

func TestExecuteTransforms_LongVersionWithExtraDash(t *testing.T) {
	registry := selector.NewRegistry()
	customServiceFn := func(version string) (filters selector.Filters, err error) {
		h.Assert(t, version == "myversion-test", "version should have been parsed as myversion-test but got %s", version)
		return filters, nil
	}

	registry.Register("myservice", selector.ServiceFiltersFn(customServiceFn))

	myService := "myservice-myversion-test"
	filters := selector.Filters{
		Service: &myService,
	}

	_, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
}
