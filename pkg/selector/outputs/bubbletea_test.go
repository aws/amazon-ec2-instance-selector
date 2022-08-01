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
	"github.com/evertras/bubble-table/table"
)

// helpers

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

// getRowsInstances reformats the given table rows into a list of instance type names
func getRowsInstances(rows []table.Row) string {
	instances := []string{}

	for _, row := range rows {
		instances = append(instances, fmt.Sprintf("%v", row.Data["Instance Type"]))
	}

	return strings.Join(instances, ", ")
}

// tests

func TestNewBubbleTeaModel_Hypervisor(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "g3_16xlarge.json")

	// test non nil Hypervisor
	model := outputs.NewBubbleTeaModel(instanceTypes)
	rows := model.TableModel.GetVisibleRows()
	expectedHypervisor := "xen"
	actualHypervisor := rows[0].Data["Hypervisor"]

	h.Assert(t, actualHypervisor == expectedHypervisor, fmt.Sprintf("Hypervisor should be %s but instead is %s", expectedHypervisor, actualHypervisor))

	// test nil Hypervisor
	instanceTypes[0].Hypervisor = nil
	model = outputs.NewBubbleTeaModel(instanceTypes)
	rows = model.TableModel.GetVisibleRows()
	expectedHypervisor = "none"
	actualHypervisor = rows[0].Data["Hypervisor"]

	h.Assert(t, actualHypervisor == expectedHypervisor, fmt.Sprintf("Hypervisor should be %s but instead is %s", expectedHypervisor, actualHypervisor))
}

func TestNewBubbleTeaModel_CPUArchitectures(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "g3_16xlarge.json")
	model := outputs.NewBubbleTeaModel(instanceTypes)
	rows := model.TableModel.GetVisibleRows()

	actualGPUArchitectures := "x86_64"
	expectedGPUArchitectures := rows[0].Data["CPU Arch"]

	h.Assert(t, actualGPUArchitectures == expectedGPUArchitectures, "CPU architecture should be (%s), but actually (%s)", expectedGPUArchitectures, actualGPUArchitectures)
}

func TestNewBubbleTeaModel_GPU(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "g3_16xlarge.json")
	model := outputs.NewBubbleTeaModel(instanceTypes)
	rows := model.TableModel.GetVisibleRows()

	// test GPU count
	expectedGPUCount := "4"
	actualGPUCount := fmt.Sprintf("%v", rows[0].Data["GPUs"])

	h.Assert(t, expectedGPUCount == actualGPUCount, "GPU count should be %s, but is actually %s", expectedGPUCount, actualGPUCount)

	// test GPU memory
	expectedGPUMemory := "32"
	actualGPUMemory := rows[0].Data["GPU Mem (GiB)"]

	h.Assert(t, expectedGPUMemory == actualGPUMemory, "GPU memory should be %s, but is actually %s", expectedGPUMemory, actualGPUMemory)

	// test GPU info
	expectedGPUInfo := "NVIDIA M60"
	actualGPUInfo := rows[0].Data["GPU Info"]

	h.Assert(t, expectedGPUInfo == actualGPUInfo, "GPU info should be (%s), but is actually (%s)", expectedGPUInfo, actualGPUInfo)
}

func TestNewBubbleTeaModel_ODPricing(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "g3_16xlarge.json")

	// test non nil OD price
	model := outputs.NewBubbleTeaModel(instanceTypes)
	rows := model.TableModel.GetVisibleRows()
	expectedODPrice := "$4.56"
	actualODPrice := fmt.Sprintf("%v", rows[0].Data["On-Demand Price/Hr"])

	h.Assert(t, actualODPrice == expectedODPrice, "Actual OD price should be %s, but is actually %s", expectedODPrice, actualODPrice)

	// test nil OD price
	instanceTypes[0].OndemandPricePerHour = nil
	model = outputs.NewBubbleTeaModel(instanceTypes)
	rows = model.TableModel.GetVisibleRows()
	expectedODPrice = "-Not Fetched-"
	actualODPrice = fmt.Sprintf("%v", rows[0].Data["On-Demand Price/Hr"])

	h.Assert(t, actualODPrice == expectedODPrice, "Actual OD price should be %s, but is actually %s", expectedODPrice, actualODPrice)
}

func TestNewBubbleTeaModel_SpotPricing(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "g3_16xlarge.json")

	// test non nil spot price
	model := outputs.NewBubbleTeaModel(instanceTypes)
	rows := model.TableModel.GetVisibleRows()
	expectedODPrice := "$1.368"
	actualODPrice := fmt.Sprintf("%v", rows[0].Data["Spot Price/Hr (30d avg)"])

	h.Assert(t, actualODPrice == expectedODPrice, "Actual spot price should be %s, but is actually %s", expectedODPrice, actualODPrice)

	// test nil spot price
	instanceTypes[0].SpotPrice = nil
	model = outputs.NewBubbleTeaModel(instanceTypes)
	rows = model.TableModel.GetVisibleRows()
	expectedODPrice = "-Not Fetched-"
	actualODPrice = fmt.Sprintf("%v", rows[0].Data["Spot Price/Hr (30d avg)"])

	h.Assert(t, actualODPrice == expectedODPrice, "Actual spot price should be %s, but is actually %s", expectedODPrice, actualODPrice)
}

func TestNewBubbleTeaModel_Rows(t *testing.T) {
	instanceTypes := getInstanceTypeDetails(t, "3_instances.json")
	model := outputs.NewBubbleTeaModel(instanceTypes)
	rows := model.TableModel.GetVisibleRows()

	h.Assert(t, len(rows) == len(instanceTypes), "Number of rows should be %d, but is actually %d", len(instanceTypes), len(rows))

	// test that order of instance types is retained
	for i := range instanceTypes {
		currInstanceName := instanceTypes[i].InstanceType
		currRowName := rows[i].Data["Instance Type"]

		h.Assert(t, *currInstanceName == currRowName, "Rows should be in following order: %s. Actual order: [%s]", outputs.OneLineOutput(instanceTypes), getRowsInstances(rows))
	}
}
