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
	"os"
	"strings"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector/outputs"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/sorter"
	h "github.com/aws/amazon-ec2-instance-selector/v3/pkg/test"
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
	mockFile, err := os.ReadFile(mockFilename)
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

		if string(actualName) != expectedName {
			return false
		}
	}

	return true
}

// Tests

func TestSort_JSONPath(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".EbsInfo.EbsOptimizedInfo.BaselineBandwidthInMbps"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults := []string{
		"a1.large",
		"a1.2xlarge",
		"a1.4xlarge",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))
}

func TestSort_SpecialCases(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "4_special_cases.json")

	// test gpus flag
	sortField := "gpus"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults := []string{
		"g3.4xlarge",
		"g3.16xlarge",
		"inf1.24xlarge",
		"inf1.2xlarge",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected gpus order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))

	// test inference accelerators flag
	sortField = "inference-accelerators"

	sortedInstances, err = sorter.Sort(instanceTypes, sortField, sortDirection)

	expectedResults = []string{
		"inf1.2xlarge",
		"inf1.24xlarge",
		"g3.16xlarge",
		"g3.4xlarge",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected inference accelerators order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))
}

func TestSort_OneElement(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "1_instance.json")

	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults := []string{"a1.2xlarge"}

	h.Ok(t, err)
	h.Assert(t, len(sortedInstances) == 1, fmt.Sprintf("Should only have 1 instance, but have: %d", len(sortedInstances)))
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))
}

func TestSort_EmptyList(t *testing.T) {
	instanceTypes := []*instancetypes.Details{}

	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)

	h.Ok(t, err)
	h.Assert(t, len(sortedInstances) == 0, fmt.Sprintf("Sorted instance types list should be empty but actually has %d elements", len(sortedInstances)))
}

func TestSort_InvalidSortField(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := "fdsafdsafdjskalfjlsf #@"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, sortedInstances == nil, "Returned sorter should be nil")
}

func TestSort_InvalidDirection(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "fdsa hfd j2 $#21"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, sortedInstances == nil, "Returned sorter should be nil")
}

func TestSort_Number(t *testing.T) {
	// All numbers (ints and floats) are evaluated as floats
	// due to the way that json unmarshalling must be done
	// in order to match json path library input format

	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test ascending
	sortField := ".MemoryInfo.SizeInMiB"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults := []string{
		"a1.large",
		"a1.2xlarge",
		"a1.4xlarge",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))

	// test descending
	sortDirection = "desc"

	sortedInstances, err = sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults = []string{
		"a1.4xlarge",
		"a1.2xlarge",
		"a1.large",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))
}

func TestSort_String(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test ascending
	sortField := ".InstanceType"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults := []string{
		"a1.2xlarge",
		"a1.4xlarge",
		"a1.large",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))

	// test descending
	sortDirection = "desc"

	sortedInstances, err = sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults = []string{
		"a1.large",
		"a1.4xlarge",
		"a1.2xlarge",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))
}

func TestSort_Invalid(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test ascending
	sortField := ".SpotPrice"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults := []string{
		"a1.large",
		"a1.2xlarge",
		"a1.4xlarge",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))

	// test descending
	sortDirection = "desc"

	sortedInstances, err = sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults = []string{
		"a1.2xlarge",
		"a1.large",
		"a1.4xlarge",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))
}

func TestSort_Unsortable(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".NetworkInfo"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, sortedInstances == nil, "returned instances list should be nil")
}

func TestSort_Pointer(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	sortField := ".EbsInfo"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, sortedInstances == nil, "returned instances list should be nil")
}

func TestSort_Bool(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")

	// test ascending
	sortField := ".HibernationSupported"
	sortDirection := "asc"

	sortedInstances, err := sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults := []string{
		"a1.4xlarge",
		"a1.2xlarge",
		"a1.large",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected ascending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))

	// test descending
	sortDirection = "desc"

	sortedInstances, err = sorter.Sort(instanceTypes, sortField, sortDirection)
	expectedResults = []string{
		"a1.large",
		"a1.2xlarge",
		"a1.4xlarge",
	}

	h.Ok(t, err)
	h.Assert(t, checkSortResults(sortedInstances, expectedResults), fmt.Sprintf("Expected descending order: [%s], but actual order: %s", strings.Join(expectedResults, ","), outputs.OneLineOutput(sortedInstances)))
}
