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
	"reflect"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/muesli/termenv"
)

const (
	// table formatting
	headerAndFooterPadding = 7
	headerPadding          = 2

	// verbose view formatting
	outlinePadding = 8

	// controls
	tableControls   = "Controls: ↑/↓ - up/down • ←/→  - left/right • shift + ←/→ - pg up/down • enter - expand • q - quit"
	verboseControls = "Controls: ↑/↓ or scroll wheel - up/down • enter - return to table • q - quit"

	// can't get terminal dimensions on startup, so use this
	initialDimensionVal = 30
)

const (
	// table states
	stateTable   = "table"
	stateVerbose = "verbose"
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

// verboseState represents the current state of the verbose view
type verboseState struct {
	// model for verbose output viewport
	viewport viewport.Model

	instanceTypes []*instancetypes.Details

	// the instance which the verbose output is focused on
	focusedInstanceName *string
}

// BubbleTeaModel is used to hold the state of the bubble tea TUI
type BubbleTeaModel struct {
	// holds the output state of the model
	state string

	// the model for the table output
	TableModel table.Model

	// holds state for the verbose view
	verboseView verboseState
}

// NewBubbleTeaModel initializes a new bubble tea Model which represents
// a stylized table to display instance types
func NewBubbleTeaModel(instanceTypes []*instancetypes.Details) BubbleTeaModel {
	return BubbleTeaModel{
		state:       stateTable,
		TableModel:  createTable(instanceTypes),
		verboseView: *initVerboseView(instanceTypes),
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
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			switch m.state {
			case stateTable:
				// switch from table state to verbose state
				m.state = stateVerbose

				// get focused instance type
				rowIndex := m.TableModel.GetHighlightedRowIndex()
				focusedInstance := m.verboseView.instanceTypes[rowIndex]

				// set content of view
				m.verboseView.focusedInstanceName = focusedInstance.InstanceType
				m.verboseView.viewport.SetContent(VerboseInstanceTypeOutput([]*instancetypes.Details{focusedInstance})[0])

				// move viewport to top of printout
				m.verboseView.viewport.SetYOffset(0)
			case stateVerbose:
				// switch from verbose state to table state
				m.state = stateTable
			}
		}
	case tea.WindowSizeMsg:
		// handle screen resizing
		m = resizeTableView(m, msg)
		m = resizeVerboseView(m, msg)
	}

	switch m.state {
	case stateTable:
		// update table
		var cmd tea.Cmd
		m.TableModel, cmd = m.TableModel.Update(msg)

		// update footer
		controlsStr := lipgloss.NewStyle().Faint(true).Render(tableControls)
		footerStr := fmt.Sprintf("Page: %d/%d | %s", m.TableModel.CurrentPage(), m.TableModel.MaxPages(), controlsStr)
		m.TableModel = m.TableModel.WithStaticFooter(footerStr)

		return m, cmd
	case stateVerbose:
		// update viewport
		var cmd tea.Cmd
		m.verboseView.viewport, cmd = m.verboseView.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View is used by bubble tea to render the bubble tea model
func (m BubbleTeaModel) View() string {
	outputStr := strings.Builder{}

	switch m.state {
	case stateTable:
		outputStr.WriteString(m.TableModel.View())
		outputStr.WriteString("\n")
	case stateVerbose:
		// format header for viewport
		instanceName := titleStyle.Render(*m.verboseView.focusedInstanceName)
		line := strings.Repeat("─", int(math.Max(0, float64(m.verboseView.viewport.Width-lipgloss.Width(instanceName)))))
		outputStr.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, instanceName, line))
		outputStr.WriteString("\n")

		outputStr.WriteString(m.verboseView.viewport.View())
		outputStr.WriteString("\n")

		// format footer for viewport
		pagePercentage := infoStyle.Render(fmt.Sprintf("%3.f%%", m.verboseView.viewport.ScrollPercent()*100))
		line = strings.Repeat("─", int(math.Max(0, float64(m.verboseView.viewport.Width-lipgloss.Width(pagePercentage)))))
		outputStr.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, line, pagePercentage))
		outputStr.WriteString("\n")

		// controls
		outputStr.WriteString(lipgloss.NewStyle().Faint(true).Render(verboseControls))
		outputStr.WriteString("\n")
	}

	return outputStr.String()
}

// table helpers:

// createRows creates a row for each instance type in the passed in list
func createRows(columnsData []*wideColumnsData) *[]table.Row {
	rows := []table.Row{}

	// create a row for each instance type
	for _, data := range columnsData {
		rowData := table.RowData{}

		// create a new row by iterating through the column data
		// struct and using struct tags as column keys
		structType := reflect.TypeOf(*data)
		structValue := reflect.ValueOf(*data)
		for i := 0; i < structType.NumField(); i++ {
			currField := structType.Field(i)
			columnName := currField.Tag.Get(columnTag)
			colValue := structValue.Field(i)
			rowData[columnName] = getUnderlyingValue(colValue)
		}

		newRow := table.NewRow(rowData)

		rows = append(rows, newRow)
	}

	return &rows
}

// maxColWidth finds the maximum width element in the given column
func maxColWidth(columnsData []*wideColumnsData, columnHeader string) int {
	// default max width is the width of the header itself with padding
	maxWidth := len(columnHeader) + headerPadding

	for _, data := range columnsData {
		// get data at given column
		structType := reflect.TypeOf(*data)
		structValue := reflect.ValueOf(*data)
		var underlyingValue interface{}
		for i := 0; i < structType.NumField(); i++ {
			currField := structType.Field(i)
			columnName := currField.Tag.Get(columnTag)
			if columnName == columnHeader {
				colValue := structValue.Field(i)
				underlyingValue = getUnderlyingValue(colValue)
				break
			}
		}

		// see if the width of the current column element exceeds
		// the previous max width
		currWidth := len(fmt.Sprintf("%v", underlyingValue))
		if currWidth > maxWidth {
			maxWidth = currWidth
		}
	}

	return maxWidth
}

// createColumns creates columns based on the tags in the wideColumnsData
// struct
func createColumns(columnsData []*wideColumnsData) *[]table.Column {
	columns := []table.Column{}

	// iterate through wideColumnsData struct and create a new column for each field tag
	columnDataStruct := wideColumnsData{}
	structType := reflect.TypeOf(columnDataStruct)
	for i := 0; i < structType.NumField(); i++ {
		columnHeader := structType.Field(i).Tag.Get(columnTag)
		newCol := table.NewColumn(columnHeader, columnHeader, maxColWidth(columnsData, columnHeader))

		columns = append(columns, newCol)
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
	// calculate and fetch all column data from instance types
	columnsData := getWideColumnsData(instanceTypes)

	newTable := table.New(*createColumns(columnsData)).
		WithRows(*createRows(columnsData)).
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

// resizeTableView will change the dimensions of the table in order to accommodate
// the new window dimensions represented by the given tea.WindowSizeMsg
func resizeTableView(model BubbleTeaModel, msg tea.WindowSizeMsg) BubbleTeaModel {
	// This is needed to handle a bug with bubble tea
	// where resizing causes misprints (https://github.com/Evertras/bubble-table/issues/121)
	termenv.ClearScreen()

	// handle width changes
	model.TableModel = model.TableModel.WithMaxTotalWidth(msg.Width)

	// handle height changes
	if headerAndFooterPadding >= msg.Height {
		// height too short to fit rows
		model.TableModel = model.TableModel.WithPageSize(0)
	} else {
		newRowsPerPage := msg.Height - headerAndFooterPadding
		model.TableModel = model.TableModel.WithPageSize(newRowsPerPage)
	}

	return model
}

// verbose helpers:

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

// initVerboseView initializes and returns a new verboseState based on the given
// instance type details
func initVerboseView(instanceTypes []*instancetypes.Details) *verboseState {
	viewportModel := viewport.New(initialDimensionVal, initialDimensionVal)
	viewportModel.MouseWheelEnabled = true

	return &verboseState{
		viewport:      viewportModel,
		instanceTypes: instanceTypes,
	}
}

// resizeVerboseView will change the dimensions of the verbose viewport in order to accommodate
// the new window dimensions represented by the given tea.WindowSizeMsg
func resizeVerboseView(model BubbleTeaModel, msg tea.WindowSizeMsg) BubbleTeaModel {
	// This is needed to handle a bug with bubble tea
	// where resizing causes misprints (https://github.com/Evertras/bubble-table/issues/121)
	termenv.ClearScreen()

	// handle width changes
	model.verboseView.viewport.Width = msg.Width

	// handle height changes
	if outlinePadding >= msg.Height {
		// height too short to fit viewport
		model.verboseView.viewport.Height = 0
	} else {
		newHeight := msg.Height - outlinePadding
		model.verboseView.viewport.Height = newHeight
	}

	return model
}
