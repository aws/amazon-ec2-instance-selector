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

package outputs_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector/outputs"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	describeInstanceTypes = "DescribeInstanceTypes"
	mockFilesPath         = "../../../test/static"
)

func getInstanceTypes(t *testing.T, file string) []*instancetypes.Details {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, describeInstanceTypes, file)
	mockFile, err := ioutil.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+string(mockFilename))
	dito := ec2.DescribeInstanceTypesOutput{}
	err = json.Unmarshal(mockFile, &dito)
	h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
	instanceTypesDetails := []*instancetypes.Details{}
	for _, it := range dito.InstanceTypes {
		odPrice := float64(0.53)
		instanceTypesDetails = append(instanceTypesDetails, &instancetypes.Details{InstanceTypeInfo: *it, OndemandPricePerHour: &odPrice})
	}
	return instanceTypesDetails
}

func TestSimpleInstanceTypeOutput(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "t3_micro.json")
	instanceTypeOut := outputs.SimpleInstanceTypeOutput(instanceTypes)
	h.Assert(t, len(instanceTypeOut) == len(instanceTypes), "Should return the same number of instance types as the data passed in")
	h.Assert(t, instanceTypeOut[0] == "t3.micro", "Should only return t3.micro")

	instanceTypeOut = outputs.SimpleInstanceTypeOutput([]*instancetypes.Details{})
	h.Assert(t, len(instanceTypeOut) == 0, "Should return 0 instance types when passed empty slice")

	instanceTypeOut = outputs.SimpleInstanceTypeOutput(nil)
	h.Assert(t, len(instanceTypeOut) == 0, "Should return 0 instance types when passed nil")
}

func TestVerboseInstanceTypeOutput(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "t3_micro.json")
	outputExpectation, err := json.MarshalIndent(instanceTypes, "", "    ")
	h.Ok(t, err)

	instanceTypeOut := outputs.VerboseInstanceTypeOutput(instanceTypes)
	h.Assert(t, len(instanceTypeOut) == len(instanceTypes), "Should return the same number of instance types as the data passed in")
	h.Assert(t, instanceTypeOut[0] == string(outputExpectation), "Should only return t3.micro")

	instanceTypeOut = outputs.VerboseInstanceTypeOutput([]*instancetypes.Details{})
	h.Assert(t, len(instanceTypeOut) == 0, "Should return 0 instance types when passed empty slice")

	instanceTypeOut = outputs.VerboseInstanceTypeOutput(nil)
	h.Assert(t, len(instanceTypeOut) == 0, "Should return 0 instance types when passed nil")
}

func TestTableOutputShort(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "t3_micro.json")
	instanceTypeOut := outputs.TableOutputShort(instanceTypes)
	outputStr := strings.Join(instanceTypeOut, "")
	lines := strings.Split(outputStr, "\n")
	h.Assert(t, len(lines) == 3, "table should include a 2 header lines and 1 instance type result line")
	h.Assert(t, strings.Contains(outputStr, "t3.micro"), "short table should include instance type")
}

func TestTableOutputWide(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "g2_2xlarge.json")
	instanceTypeOut := outputs.TableOutputWide(instanceTypes)
	outputStr := strings.Join(instanceTypeOut, "")
	lines := strings.Split(outputStr, "\n")
	h.Assert(t, len(lines) == 3, "table should include a 2 header lines and 1 instance type result line")
	h.Assert(t, strings.Contains(outputStr, "g2.2xlarge"), "table should include instance type")
	h.Assert(t, strings.Contains(outputStr, "Moderate"), "wide table should include network performance")
	h.Assert(t, strings.Contains(outputStr, "NVIDIA K520"), "wide table should include GPU Info")
}

func TestTableOutput_MBtoGB(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "g2_2xlarge.json")
	instanceTypeOut := outputs.TableOutputWide(instanceTypes)
	outputStr := strings.Join(instanceTypeOut, "")
	h.Assert(t, strings.Contains(outputStr, "15"), "table should include 15 GB of memory")
	h.Assert(t, strings.Contains(outputStr, "4"), "wide table should include 4 GB of gpu memory")

	instanceTypeOut = outputs.TableOutputShort(instanceTypes)
	outputStr = strings.Join(instanceTypeOut, "")
	h.Assert(t, strings.Contains(outputStr, "15"), "table should include 15 GB of memory")
}

func TestOneLineOutput(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "t3_micro_and_p3_16xl.json")
	instanceTypeOut := outputs.OneLineOutput(instanceTypes)
	h.Assert(t, len(instanceTypeOut) == 1, "Should always return 1 line")
	h.Assert(t, instanceTypeOut[0] == "t3.micro,p3.16xlarge", "Should return both instance types separated by a comma")

	instanceTypeOut = outputs.OneLineOutput([]*instancetypes.Details{})
	h.Assert(t, len(instanceTypeOut) == 0, "Should return 0 instance types when passed empty slice")

	instanceTypeOut = outputs.OneLineOutput(nil)
	h.Assert(t, len(instanceTypeOut) == 0, "Should return 0 instance types when passed nil")
}

func TestTruncateResults(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "25_instances.json")

	// test 0 for max results
	maxResults := aws.Int(0)
	truncatedResult, numTrucated, err := outputs.TruncateResults(maxResults, instanceTypes)
	h.Ok(t, err)
	h.Assert(t, len(truncatedResult) == 0, fmt.Sprintf("Should return 0 instance types since max results is set to %d, but only %d are returned in total", *maxResults, len(truncatedResult)))
	h.Assert(t, numTrucated == 25, fmt.Sprintf("Should truncate 25 results, but actually truncated: %d results", numTrucated))

	// test 1 for max results
	maxResults = aws.Int(1)
	truncatedResult, numTrucated, err = outputs.TruncateResults(maxResults, instanceTypes)
	h.Ok(t, err)
	h.Assert(t, len(truncatedResult) == 1, fmt.Sprintf("Should return 1 instance type since max results is set to %d, but only %d are returned in total", *maxResults, len(truncatedResult)))
	h.Assert(t, numTrucated == 24, fmt.Sprintf("Should truncate 24 results, but actually truncated: %d results", numTrucated))

	// test 30 for max results
	maxResults = aws.Int(30)
	truncatedResult, numTrucated, err = outputs.TruncateResults(maxResults, instanceTypes)
	h.Ok(t, err)
	h.Assert(t, len(truncatedResult) == 25, fmt.Sprintf("Should return 25 instance types since max results is set to %d but only %d are returned in total", *maxResults, len(truncatedResult)))
	h.Assert(t, numTrucated == 0, fmt.Sprintf("Should truncate 0 results, but actually truncated: %d results", numTrucated))
}

func TestTruncateResults_NegativeMaxResults(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "25_instances.json")

	maxResults := aws.Int(-1)
	formattedResult, numTrucated, err := outputs.TruncateResults(maxResults, instanceTypes)

	h.Assert(t, err != nil, "An error should be returned")
	h.Assert(t, formattedResult == nil, fmt.Sprintf("returned list should be nil, but it is actually: %s", outputs.OneLineOutput(formattedResult)))
	h.Assert(t, numTrucated == 0, fmt.Sprintf("No results should be truncated, but %d results were truncated", numTrucated))
}

func TestTrucateResults_NilMaxResults(t *testing.T) {
	instanceTypes := getInstanceTypes(t, "25_instances.json")

	var maxResults *int = nil
	formattedResult, numTrucated, err := outputs.TruncateResults(maxResults, instanceTypes)

	h.Ok(t, err)
	h.Assert(t, len(formattedResult) == 25, fmt.Sprintf("Should return 25 instance types since max results is set to nil but only %d are returned in total", len(formattedResult)))
	h.Assert(t, numTrucated == 0, fmt.Sprintf("No results should be truncated, but actually truncated: %d results", numTrucated))
}
