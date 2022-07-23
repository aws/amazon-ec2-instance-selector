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

package sorter_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector/outputs"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/sorter"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
)

const (
	mockFilesPath              = "../../test/static"
	describeInstanceTypesPages = "DescribeInstanceTypesPages"
)

// Helpers

// getInstanceTypeDetails unmarshalls the json file in the given testing folder
// and returns a list of instance type details
func getInstanceTypeDetails(t *testing.T, file string) []*instancetypes.Details {
	folder := "FilterVerbose"
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, folder, file)
	mockFile, err := ioutil.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+string(mockFilename))

	instanceTypes := []*instancetypes.Details{}
	err = json.Unmarshal(mockFile, &instanceTypes)
	h.Assert(t, err == nil, fmt.Sprintf("Error parsing mock json file contents %s. Error: %v", mockFilename, err))
	return instanceTypes
}

// checkSortResults is a helper function for comparing the results of sorting tests. Returns true if
// the order of instance types in the instanceTypes list matches the the order of instance type names
// in the expectedResult list, and returns false otherwise.
func checkSortResults(instanceTypes []*instancetypes.Details, expectedResult []string) bool {
	if len(instanceTypes) != len(expectedResult) {
		return false
	}

	for i := 0; i < len(instanceTypes); i++ {
		actualName := instanceTypes[i].InstanceTypeInfo.InstanceType
		expectedName := expectedResult[i]

		if actualName == nil || *actualName != expectedName {
			return false
		}
	}

	return true
}

// Tests

func TestNewSorter_JSONPath(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)

	h.Ok(t, err)
	h.Assert(t, result != nil, "Returned sorter should not be nil")
}

func TestNewSorter_SpecialCases(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test gpus flag
	sortField := "gpus"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)

	h.Ok(t, err)
	h.Assert(t, result != nil, "Returned sorter should not be nil")

	// test inference accelerators flag
	sortField = "inference-accelerators"

	result, err = sorter.NewSorter(instanceTypes, sortField, sortDirection)

	h.Ok(t, err)
	h.Assert(t, result != nil, "Returned sorter should not be nil")
}

func TestNewSorter_OneElement(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "1_instance.json")

	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)

	h.Ok(t, err)
	h.Assert(t, result != nil, "Returned sorter should not be nil")
}

func TestNewSorter_EmptyList(t *testing.T) {
	instanceTypes := []*instancetypes.Details{}

	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)

	h.Ok(t, err)
	h.Assert(t, result != nil, "Returned sorter should not be nil")
}

func TestNewSorter_InvalidSortField(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := "fdsafdsafdjskalfjlsf #@"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, result == nil, "Returned sorter should be nil")
}

func TestNewSorter_InvalidDirection(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "fdsa hfd j2 $#21"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, result == nil, "Returned sorter should be nil")
}

func TestInstanceTypes(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	sorterInstances := result.InstanceTypes()

	h.Assert(t,
		len(sorterInstances) == len(instanceTypes),
		fmt.Sprintf("returned instance types list should have %d elements but actually has %d elements.",
			len(instanceTypes),
			len(sorterInstances),
		),
	)
	for i := range instanceTypes {
		h.Assert(t,
			*instanceTypes[i].InstanceType == *sorterInstances[i].InstanceType,
			fmt.Sprintf("Instance types in sorter (%s) should be the same as original list (%s).",
				strings.Join(outputs.OneLineOutput(instanceTypes), ", "),
				strings.Join(outputs.OneLineOutput(sorterInstances), ", "),
			),
		)
	}
	h.Assert(t,
		len(sorterInstances) == len(instanceTypes),
		fmt.Sprintf("Instance types in sorter (%s) should be the same as original list (%s).",
			strings.Join(outputs.OneLineOutput(instanceTypes), ", "),
			strings.Join(outputs.OneLineOutput(sorterInstances), ", "),
		),
	)
}

func TestSort_Number(t *testing.T) {
	// All numbers (ints and floats) are evaluated as floats
	// due to the way that json unmarshalling must be done
	// in order to match json path library input format

	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test ascending
	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances := result.InstanceTypes()
	expectedResults := []string{
		"a1.large",
		"a1.2xlarge",
		"a1.4xlarge",
	}
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))

	// test descending
	sortDirection = "desc"

	result, err = sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances = result.InstanceTypes()
	expectedResults = []string{
		"a1.4xlarge",
		"a1.2xlarge",
		"a1.large",
	}
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))
}

func TestSort_String(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test ascending
	sortField := ".InstanceType"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances := result.InstanceTypes()
	expectedResults := []string{
		"a1.2xlarge",
		"a1.4xlarge",
		"a1.large",
	}
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))

	// test descending
	sortDirection = "desc"

	result, err = sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances = result.InstanceTypes()
	expectedResults = []string{
		"a1.large",
		"a1.4xlarge",
		"a1.2xlarge",
	}
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))
}

func TestSort_Invalid(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test ascending
	sortField := ".SpotPrice"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances := result.InstanceTypes()
	expectedResults := []string{
		"a1.large",
		"a1.2xlarge",
		"a1.4xlarge",
	}
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))

	// test descending
	sortDirection = "desc"

	result, err = sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances = result.InstanceTypes()
	expectedResults = []string{
		"a1.2xlarge",
		"a1.large",
		"a1.4xlarge",
	}
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))
}

func TestSort_Unsortable(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".NetworkInfo"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Assert(t, err != nil, "An error should be returned")
}

func TestSort_Pointer(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".EbsInfo"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Assert(t, err != nil, "An error should be returned")
}

func TestSort_Bool(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test ascending
	sortField := ".HibernationSupported"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances := result.InstanceTypes()
	expectedResults := []string{
		"a1.4xlarge",
		"a1.2xlarge",
		"a1.large",
	}
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))

	// test descending
	sortDirection = "desc"

	result, err = sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances = result.InstanceTypes()
	expectedResults = []string{
		"a1.large",
		"a1.2xlarge",
		"a1.4xlarge",
	}
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))
}

func TestSort_EmptyList(t *testing.T) {
	instanceTypes := []*instancetypes.Details{}

	sortField := ".HibernationSupported"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances := result.InstanceTypes()
	h.Assert(t, len(sorterInstances) == 0, fmt.Sprintf("sorter instance types list should be empty but actually has %d elements", len(sorterInstances)))
}

func TestSort_OneElement(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "1_instance.json")

	sortField := ".HibernationSupported"
	sortDirection := "asc"

	result, err := sorter.NewSorter(instanceTypes, sortField, sortDirection)
	h.Ok(t, err)

	err = result.Sort()
	h.Ok(t, err)

	sorterInstances := result.InstanceTypes()
	expectedResults := []string{"a1.2xlarge"}
	h.Assert(t, len(sorterInstances) == 1, fmt.Sprintf("sorter instance types list should have 1 element actually has %d elements", len(sorterInstances)))
	h.Assert(t, checkSortResults(sorterInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sorterInstances)))
}
