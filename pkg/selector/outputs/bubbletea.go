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
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/muesli/termenv"
)

const (
	// table formatting
	headerAndFooterPadding = 7
	headerPadding          = 2
)

const (
	// column headers
	colInstanceType       = "Instance Type"
	colVCPU               = "VCPUs"
	colMemory             = "Memory (GiB)"
	colHypervisor         = "Hypervisor"
	colCurrentGen         = "Current Gen"
	colHibernationSupport = "Hibernation Support"
	colCPUArch            = "CPU Architecture"
	colNetworkPerformance = "Network Performance"
	colENI                = "ENIs"
	colGPU                = "GPUs"
	colGPUMemory          = "GPU Memory (GiB)"
	colGPUInfo            = "GPU Info"
	colODPrice            = "On-Demand Price/Hr"
	colSpotPrice          = "Spot Price/Hr (30 day avg)"

	// controls
	controlsString = "Controls: ↑/↓ - up/down • ←/→  - left/right • shift + ←/→ - pg up/down • q - quit"
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
	// the model for the table output
	TableModel table.Model
}

// NewBubbleTeaModel initializes a new bubble tea Model which represents
// a stylized table to display instance types
func NewBubbleTeaModel(instanceTypes []*instancetypes.Details) BubbleTeaModel {
	return BubbleTeaModel{
		TableModel: createTable(instanceTypes),
	}
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
	case tea.WindowSizeMsg:
		// handle screen resizing

		// This is needed to handle a bug with bubble tea
		// where resizing causes misprints (https://github.com/Evertras/bubble-table/issues/121)
		termenv.ClearScreen()

		// handle width changes
		m.TableModel = m.TableModel.WithMaxTotalWidth(msg.Width)

		// handle height changes
		if headerAndFooterPadding >= msg.Height {
			// height too short to fit rows
			m.TableModel = m.TableModel.WithPageSize(0)
		} else {
			newRowsPerPage := msg.Height - headerAndFooterPadding
			m.TableModel = m.TableModel.WithPageSize(newRowsPerPage)
		}
	}

	// update table
	var cmd tea.Cmd
	m.TableModel, cmd = m.TableModel.Update(msg)

	// update footer
	controlsStr := lipgloss.NewStyle().Faint(true).Render(controlsString)
	footerStr := fmt.Sprintf("Page: %d/%d | %s", m.TableModel.CurrentPage(), m.TableModel.MaxPages(), controlsStr)
	m.TableModel = m.TableModel.WithStaticFooter(footerStr)

	return m, cmd
}

// View is used by bubble tea to render the bubble tea model
func (m BubbleTeaModel) View() string {
	outputStr := strings.Builder{}

	outputStr.WriteString(m.TableModel.View())
	outputStr.WriteString("\n")

	return outputStr.String()
}

// table creation helpers:

// createRows creates a row for each instance type in the passed in list
func createRows(instanceTypes []*instancetypes.Details) *[]table.Row {
	rows := []table.Row{}

	for _, instanceType := range instanceTypes {
		none := "none"
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

		onDemandPricePerHourStr := "-Not Fetched-"
		spotPricePerHourStr := "-Not Fetched-"
		if instanceType.OndemandPricePerHour != nil {
			onDemandPricePerHourStr = "$" + formatFloat(*instanceType.OndemandPricePerHour)
		}
		if instanceType.SpotPrice != nil {
			spotPricePerHourStr = "$" + formatFloat(*instanceType.SpotPrice)
		}

		newRow := table.NewRow(table.RowData{
			colInstanceType:       *instanceType.InstanceType,
			colVCPU:               *instanceType.VCpuInfo.DefaultVCpus,
			colMemory:             formatFloat(float64(*instanceType.MemoryInfo.SizeInMiB) / 1024.0),
			colHypervisor:         *hyperisor,
			colCurrentGen:         *instanceType.CurrentGeneration,
			colHibernationSupport: *instanceType.HibernationSupported,
			colCPUArch:            strings.Join(cpuArchitectures, ", "),
			colNetworkPerformance: *instanceType.NetworkInfo.NetworkPerformance,
			colENI:                *instanceType.NetworkInfo.MaximumNetworkInterfaces,
			colGPU:                gpus,
			colGPUMemory:          formatFloat(float64(gpuMemory) / 1024.0),
			colGPUInfo:            strings.Join(gpuType, ", "),
			colODPrice:            onDemandPricePerHourStr,
			colSpotPrice:          spotPricePerHourStr,
		})

		rows = append(rows, newRow)
	}

	return &rows
}

// createColumns creates columns based on the column key constants
func createColumns() *[]table.Column {
	// GPU info field has long strings, so have different buffer length for that field
	gpuInfoBuffer := 10

	// create columns based on column names
	columns := []table.Column{
		table.NewColumn(colInstanceType, colInstanceType, len(colInstanceType)+headerPadding),
		table.NewColumn(colVCPU, colVCPU, len(colVCPU)+headerPadding),
		table.NewColumn(colMemory, colMemory, len(colMemory)+headerPadding),
		table.NewColumn(colHypervisor, colHypervisor, len(colHypervisor)+headerPadding),
		table.NewColumn(colCurrentGen, colCurrentGen, len(colCurrentGen)+headerPadding),
		table.NewColumn(colHibernationSupport, colHibernationSupport, len(colHibernationSupport)+headerPadding),
		table.NewColumn(colCPUArch, colCPUArch, len(colCPUArch)+headerPadding),
		table.NewColumn(colNetworkPerformance, colNetworkPerformance, len(colNetworkPerformance)+headerPadding),
		table.NewColumn(colENI, colENI, len(colENI)+headerPadding),
		table.NewColumn(colGPU, colGPU, len(colGPU)+headerPadding),
		table.NewColumn(colGPUMemory, colGPUMemory, len(colGPUMemory)+headerPadding),
		table.NewColumn(colGPUInfo, colGPUInfo, len(colGPUInfo)+gpuInfoBuffer),
		table.NewColumn(colODPrice, colODPrice, len(colODPrice)+headerPadding),
		table.NewColumn(colSpotPrice, colSpotPrice, len(colSpotPrice)+headerPadding),
	}

	return &columns
}

// createKeyMap creates a KeyMap with the controls for the table
func createKeyMap() *table.KeyMap {
	keys := table.KeyMap{
		RowDown: key.NewBinding(
			key.WithKeys("down"),
		),
		RowUp: key.NewBinding(
			key.WithKeys("up"),
		),
		ScrollLeft: key.NewBinding(
			key.WithKeys("left"),
		),
		ScrollRight: key.NewBinding(
			key.WithKeys("right"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("shift+right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("shift+left"),
		),
	}

	return &keys
}

// createTable creates an intractable table which contains information about all of
// the given instance types
func createTable(instanceTypes []*instancetypes.Details) table.Model {
	// can't get terminal size yet, so set temporary value
	initialDimensionVal := 30

	newTable := table.New(*createColumns()).
		WithRows(*createRows(instanceTypes)).
		WithKeyMap(*createKeyMap()).
		WithPageSize(initialDimensionVal).
		Focused(true).
		Border(customBorder).
		WithMaxTotalWidth(initialDimensionVal).
		WithHorizontalFreezeColumnCount(1).
		WithBaseStyle(
			lipgloss.NewStyle().
				Align((lipgloss.Left)),
		).
		HeaderStyle(lipgloss.NewStyle().Align(lipgloss.Center).Bold(true))

	return newTable
}
