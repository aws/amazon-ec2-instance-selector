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

package outputs

import (
	"fmt"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	rowsPerPage  = 10
	headerBuffer = 2
	maxWidth     = 140
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

var (
	customBorder = table.Border{
		Top:    "─",
		Left:   "│",
		Right:  "│",
		Bottom: "─",

		TopRight:    "╮",
		TopLeft:     "╭",
		BottomRight: "╯",
		BottomLeft:  "╰",

		TopJunction:    "┬",
		LeftJunction:   "├",
		RightJunction:  "┤",
		BottomJunction: "┴",
		InnerJunction:  "┼",

		InnerDivider: "│",
	}
)

// BubbleTeaModel is used to hold the state of the bubble tea TUI
type BubbleTeaModel struct {
	tableModel table.Model
}

// createRows creates a row for each instance type in the passed in list
func createRows(instanceTypes []*instancetypes.Details) []table.Row {
	rows := []table.Row{}

	none := "none"
	for _, instanceType := range instanceTypes {
		hyperisor := instanceType.Hypervisor
		if hyperisor == nil {
			hyperisor = &none
		}

		cpuArchitectures := []string{}
		for _, cpuArch := range instanceType.ProcessorInfo.SupportedArchitectures {
			cpuArchitectures = append(cpuArchitectures, *cpuArch)
		}

		gpus := int64(0)
		gpuMemory := int64(0)
		gpuType := []string{}
		if instanceType.GpuInfo != nil {
			gpuMemory = *instanceType.GpuInfo.TotalGpuMemoryInMiB
			for _, gpuInfo := range instanceType.GpuInfo.Gpus {
				gpus = gpus + *gpuInfo.Count
				gpuType = append(gpuType, *gpuInfo.Manufacturer+" "+*gpuInfo.Name)
			}
		}

		onDemandPricePerHourStr := "-Not Fetched"
		spotPricePerHourStr := "-Not Fetched-"
		if instanceType.OndemandPricePerHour != nil {
			onDemandPricePerHourStr = "$" + formatFloat(*instanceType.OndemandPricePerHour)
		}
		if instanceType.SpotPrice != nil {
			spotPricePerHourStr = "$" + formatFloat(*instanceType.SpotPrice)
		}

		newRow := table.NewRow(table.RowData{
			colKeyInstanceType:       *instanceType.InstanceType,
			colKeyVCPU:               *instanceType.VCpuInfo.DefaultVCpus,
			colKeyMemory:             formatFloat(float64(*instanceType.MemoryInfo.SizeInMiB) / 1024.0),
			colKeyHypervisor:         *hyperisor,
			colKeyCurrentGen:         *instanceType.CurrentGeneration,
			colKeyHibernationSupport: *instanceType.HibernationSupported,
			colKeyCPUArch:            strings.Join(cpuArchitectures, ", "),
			colKeyNetworkPerformance: *instanceType.NetworkInfo.NetworkPerformance,
			colKeyENI:                *instanceType.NetworkInfo.MaximumNetworkInterfaces,
			colKeyGPU:                gpus,
			colKeyGPUMemory:          formatFloat(float64(gpuMemory) / 1024.0),
			colKeyGPUInfo:            strings.Join(gpuType, ", "),
			colKeyODPrice:            onDemandPricePerHourStr,
			colKeySpotPrice:          spotPricePerHourStr,
		})

		rows = append(rows, newRow)
	}

	return rows
}

// NewBubbleTeaModel initializes a new bubble tea Model which represents
// a stylized table to display instance types
func NewBubbleTeaModel(instanceTypes []*instancetypes.Details) BubbleTeaModel {
	// create columns based on column names
	columns := []table.Column{
		table.NewColumn(colKeyInstanceType, colKeyInstanceType, len(colKeyInstanceType)+headerBuffer),
		table.NewColumn(colKeyVCPU, colKeyVCPU, len(colKeyVCPU)+headerBuffer),
		table.NewColumn(colKeyMemory, colKeyMemory, len(colKeyMemory)+headerBuffer),
		table.NewColumn(colKeyHypervisor, colKeyHypervisor, len(colKeyHypervisor)+headerBuffer),
		table.NewColumn(colKeyCurrentGen, colKeyCurrentGen, len(colKeyCurrentGen)+headerBuffer),
		table.NewColumn(colKeyHibernationSupport, colKeyHibernationSupport, len(colKeyHibernationSupport)+headerBuffer),
		table.NewColumn(colKeyCPUArch, colKeyCPUArch, len(colKeyCPUArch)+headerBuffer),
		table.NewColumn(colKeyNetworkPerformance, colKeyNetworkPerformance, len(colKeyNetworkPerformance)+headerBuffer),
		table.NewColumn(colKeyENI, colKeyENI, len(colKeyENI)+headerBuffer),
		table.NewColumn(colKeyGPU, colKeyGPU, len(colKeyGPU)+headerBuffer),
		table.NewColumn(colKeyGPUMemory, colKeyGPUMemory, len(colKeyGPUMemory)+headerBuffer),
		table.NewColumn(colKeyGPUInfo, colKeyGPUInfo, len(colKeyGPUInfo)+10),
		table.NewColumn(colKeyODPrice, colKeyODPrice, len(colKeyODPrice)+headerBuffer),
		table.NewColumn(colKeySpotPrice, colKeySpotPrice, len(colKeySpotPrice)+headerBuffer),
	}

	// create rows based on instance type details
	rows := createRows(instanceTypes)

	// set keys for traversing table
	keys := table.DefaultKeyMap()
	keys.ScrollLeft.SetKeys("left")
	keys.ScrollRight.SetKeys("right")
	keys.PageDown.SetKeys("shift+down", "l", "pgdown")
	keys.PageUp.SetKeys("shift+up", "h", "pgup")

	model := BubbleTeaModel{
		tableModel: table.New(columns).
			WithRows(rows).
			WithKeyMap(keys).
			WithPageSize(rowsPerPage).
			Focused(true).
			Border(customBorder).
			WithMaxTotalWidth(maxWidth).
			// TODO: fix magic number
			WithHorizontalFreezeColumnCount(1).
			WithBaseStyle(
				lipgloss.NewStyle().
					Align((lipgloss.Left)),
			).
			// TODO: maybe remove this because never used
			WithMissingDataIndicatorStyled(table.StyledCell{
				Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#faa")),
				Data:  "Not Fetched",
			}).
			HeaderStyle(lipgloss.NewStyle().Align(lipgloss.Center).Bold(true)),
	}

	return model
}

// Init is used by bubble tea to initialize a bubble tea table
func (m BubbleTeaModel) Init() tea.Cmd {
	return nil
}

// Update is used by bubble tea to update the state of the bubble
// tea model based on user input
func (m BubbleTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// check for quit
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	}

	// update table
	var cmd tea.Cmd
	m.tableModel, cmd = m.tableModel.Update(msg)

	// update footer
	footerStr := fmt.Sprintf("Page: %d/%d", m.tableModel.CurrentPage(), m.tableModel.MaxPages())
	m.tableModel = m.tableModel.WithStaticFooter(footerStr)

	return m, cmd
}

// View is used by bubble tea to render the bubble tea model
func (m BubbleTeaModel) View() string {
	outputStr := strings.Builder{}

	outputStr.WriteString(m.tableModel.View())
	outputStr.WriteString("\n")

	// TODO: add section explaining controls (similar to
	// lighter colored text in default bubble tea example)
	// TODO: put controls in the footer
	controlsStr := "↑/↓ - up/down • ←/→  - left/right • shift + ↑/↓ - pg up/down"
	controlsStr = lipgloss.NewStyle().Faint(true).Render(controlsStr)
	outputStr.WriteString(controlsStr)

	return outputStr.String()
}

// TODO:
// Possible idea:
// - Make it so that you can select rows
// - when row is selected, table disappears and is
//   replaced with a detailed description of all of
//   the information about that one instance type
//      - similar to verbose output (but with nice ~bubble tea~ stuff)
// - Can leave this "expanded detailed" look by pressing something?
//      - will bring you back to the table view at preferibly the same
//         page you were before
// - will require changing the model struct to have more state
// - maybe look into bubble tea cmds
//      - if not then maybe can just have a flag in this
//		  file which then is checked in update and view to
//		  have different results printed and different keys
//		  listened to
