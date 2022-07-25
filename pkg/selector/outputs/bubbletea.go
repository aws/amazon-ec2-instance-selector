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
	"math"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/muesli/termenv"
)

const (
	// table formatting
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

	// controls
	controlsString = "Controls: ↑/↓ - up/down • ←/→  - left/right • shift + ←/→ - pg up/down • q - quit"

	// table states
	tableState   = "table"
	verboseState = "verbose"
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

// styling for viewport
var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

// BubbleTeaModel is used to hold the state of the bubble tea TUI
type BubbleTeaModel struct {
	instanceTypes []*instancetypes.Details

	// holds the output state of the model
	// TODO: Instead of a string maybe have a "active model"?
	// and then implement a model for each state? Like the verbose printout
	// state would have its own update and view methods?
	// This could be easier to expand perhaps?
	state string

	// the model for the table output
	tableModel table.Model

	// verbose output
	focusedInstance string

	// model for verbose output viewport
	verboseViewport viewport.Model
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
	// TODO: clear up how this also makes verbose output

	// TODO: maybe move table creation to helper method

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
	// TODO: change from having default key map to just the ones listed
	keys := table.DefaultKeyMap()
	keys.ScrollLeft.SetKeys("left")
	keys.ScrollRight.SetKeys("right")
	keys.PageDown.SetKeys("x", "pgdown", "shift+right")
	keys.PageUp.SetKeys("z", "pgup", "shift+left")
	keys.RowUp.SetKeys("up")
	keys.RowDown.SetKeys("down")

	// TODO: create verbose output model
	// probably move this somewhere else
	// TODO: have different constant for height
	viewportModel := viewport.New(maxWidth, rowsPerPage)
	viewportModel.MouseWheelEnabled = true

	model := BubbleTeaModel{
		instanceTypes: instanceTypes,
		state:         tableState,
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
		verboseViewport: viewportModel,
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
	// TODO: check message for screen resizing
	//		maybe can fix bug with printing out verbose output not
	//		clearing screen correctly

	// check for quit
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	}

	// check view state
	switch m.state {
	case tableState:
		// update table
		var cmd tea.Cmd
		m.tableModel, cmd = m.tableModel.Update(msg)

		// update footer
		controlsStr := lipgloss.NewStyle().Faint(true).Render(controlsString)
		footerStr := fmt.Sprintf("Page: %d/%d | %s", m.tableModel.CurrentPage(), m.tableModel.MaxPages(), controlsStr)
		m.tableModel = m.tableModel.WithStaticFooter(footerStr)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			// check for change in state
			switch msg.String() {
			case "enter", "c":
				// clear screen
				//termenv.ClearScreen()

				// change to verbose state
				m.state = verboseState
				rowIndex := m.tableModel.GetHighlightedRowIndex()
				focusedInstance := m.instanceTypes[rowIndex]

				// TODO: cleanup? maybe with substring to remove new line char
				m.focusedInstance = *focusedInstance.InstanceType
				m.verboseViewport.SetContent(VerboseInstanceTypeOutput([]*instancetypes.Details{focusedInstance})[0])

				// move viewport to top of printout
				m.verboseViewport.SetYOffset(0)
			}
		case tea.WindowSizeMsg:
			// handle screen resizing
			// TODO: use to handle bubble tea bug where
			// printing outside of screen bounds causes print bugs
			termenv.ClearScreen()

			// handle width changes
			if msg.Width < maxWidth {
				m.tableModel = m.tableModel.WithMaxTotalWidth(msg.Width)
			} else {
				m.tableModel = m.tableModel.WithMaxTotalWidth(maxWidth)
			}

			// TODO: handle height changes?
		}

		return m, cmd
	case verboseState:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "c":
				// clear screen
				// TODO: use to handle bubble tea bug where
				// printing outside of screen bounds causes print bugs
				//termenv.ClearScreen()

				// change to table state
				m.state = tableState
			}
		case tea.WindowSizeMsg:
			// handle screen resizing
			// TODO: use to handle bubble tea bug where
			// printing outside of screen bounds causes print bugs
			termenv.ClearScreen()

			// handle width changes
			if msg.Width < maxWidth {
				m.verboseViewport.Width = msg.Width
			} else {
				m.verboseViewport.Width = maxWidth
			}

			// TODO: handle height changes?
		}

		// TODO: clean up
		// update viewport model
		var cmd tea.Cmd
		m.verboseViewport, cmd = m.verboseViewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View is used by bubble tea to render the bubble tea model
func (m BubbleTeaModel) View() string {
	outputStr := strings.Builder{}

	switch m.state {
	case tableState:
		outputStr.WriteString(m.tableModel.View())
		outputStr.WriteString("\n")
	case verboseState:
		// TODO: instead of clearing screen on update, look into viewport for
		// verbose printout

		// for _, str := range m.focusedInstanceOutput {
		// 	outputStr.WriteString(str)
		// 	outputStr.WriteString("\n")
		// }

		// format header for viewport
		instanceName := titleStyle.Render(m.focusedInstance)
		line := strings.Repeat("─", int(math.Max(0, float64(m.verboseViewport.Width-lipgloss.Width(instanceName)))))
		outputStr.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, instanceName, line))
		outputStr.WriteString("\n")

		outputStr.WriteString(m.verboseViewport.View())
		outputStr.WriteString("\n")

		// format footer for viewport
		pagePercentage := infoStyle.Render(fmt.Sprintf("%3.f%%", m.verboseViewport.ScrollPercent()*100))
		line = strings.Repeat("─", int(math.Max(0, float64(m.verboseViewport.Width-lipgloss.Width(pagePercentage)))))
		outputStr.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, line, pagePercentage))
		outputStr.WriteString("\n")
	}

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
