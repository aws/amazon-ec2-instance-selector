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

package cli

import (
	"github.com/evertras/bubble-table/table"
)

const (
	// column keys
	colKeyInstanceType       = "Instance Type"
	colKeyVCPU               = "VCPUs"
	colKeyMemory             = "Memory (GiB)"
	colKeyHypervisor         = "Hypervisor"
	colKeyCurrentGen         = "Current Gen"
	colKeyHibernationSupport = "Hibernation Support"
	colKeyCPUArch            = "CPU Architecture"
	colKeyNetworkPerformance = "Network Performance"
	colKeyENI                = "ENIs"
	colKeyGPU                = "GPUs"
	colKeyGPUMemory          = "GPU Memory (GiB)"
	colKeyGPUInfo            = "GPU Info"
	colKeyODPrice            = "On-Demand Price/Hr"
	colKeySpotPrice          = "Spot Price/Hr (30 day avg)"
)

// TODO: define custom boarder
// var (customBoarder = table.Boarder{...})

// Model is used to hold the state of the bubble tea UI
type Model struct {
	tableModel table.Model
}

// NewModel initializes a new bubble tea Model which represents
// a stylized table to display instance types
func NewModel() Model {
	// create columns based on column names
	columns := []table.Column{
		table.NewColumn(colKeyInstanceType, colKeyInstanceType, 10),
		table.NewColumn(colKeyVCPU, colKeyVCPU, 5),
		table.NewColumn(colKeyMemory, colKeyMemory, 5),
		table.NewColumn(colKeyHypervisor, colKeyHypervisor, 5),
		table.NewColumn(colKeyCurrentGen, colKeyCurrentGen, 5),
		table.NewColumn(colKeyHibernationSupport, colKeyHibernationSupport, 5),
		table.NewColumn(colKeyCPUArch, colKeyCPUArch, 10),
		table.NewColumn(colKeyNetworkPerformance, colKeyNetworkPerformance, 10),
		table.NewColumn(colKeyENI, colKeyENI, 5),
		table.NewColumn(colKeyGPU, colKeyGPU, 10),
		table.NewColumn(colKeyGPUMemory, colKeyGPUMemory, 5),
		table.NewColumn(colKeyGPUInfo, colKeyGPUInfo, 10),
		table.NewColumn(colKeyODPrice, colKeyODPrice, 10),
		table.NewColumn(colKeySpotPrice, colKeySpotPrice, 10),
	}

	// TODO: remove this
	_ = columns

	return Model{}
}
