// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outputs_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector/outputs"
	h "github.com/aws/amazon-ec2-instance-selector/v3/pkg/test"
)

const (
	describeInstanceTypes = "DescribeInstanceTypes"
	mockFilesPath         = "../../../test/static"
)

func getInstanceTypes(t *testing.T, file string) []*instancetypes.Details {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, describeInstanceTypes, file)
	mockFile, err := os.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+string(mockFilename))
	dito := ec2.DescribeInstanceTypesOutput{}
	err = json.Unmarshal(mockFile, &dito)
	h.Assert(t, err == nil, "Error parsing mock json file contents"+mockFilename)
	instanceTypesDetails := []*instancetypes.Details{}
	for _, it := range dito.InstanceTypes {
		odPrice := float64(0.53)
		instanceTypesDetails = append(instanceTypesDetails, &instancetypes.Details{InstanceTypeInfo: it, OndemandPricePerHour: &odPrice})
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
