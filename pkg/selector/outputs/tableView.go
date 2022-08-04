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
	"reflect"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	// table formatting
	headerAndFooterPadding = 8
	headerPadding          = 2

	// controls
	tableControls = "Controls: ↑/↓ - up/down • ←/→  - left/right • shift + ←/→ - pg up/down • e - expand • q - quit"
	ellipses      = "..."
)

type tableModel struct {
	// the model for the table output
	table table.Model

	tableWidth int

	// the model for the filtering text input
	filterTextInput textinput.Model
}

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

// initTableModel initializes and returns a new tableModel based on the given
// instance type details
func initTableModel(instanceTypes []*instancetypes.Details) *tableModel {
	return &tableModel{
		table:           createTable(instanceTypes),
		tableWidth:      initialDimensionVal,
		filterTextInput: createFilterTextInput(),
	}
}

// createFilterTextInput creates and styles a text input for filtering
func createFilterTextInput() textinput.Model {
	filterTextInput := textinput.New()
	filterTextInput.Prompt = "Filter: "
	filterTextInput.PromptStyle = lipgloss.NewStyle().Bold(true)

	return filterTextInput
}

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
		newCol := table.NewColumn(columnHeader, columnHeader, maxColWidth(columnsData, columnHeader)).
			WithFiltered(true)

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
		HeaderStyle(lipgloss.NewStyle().Align(lipgloss.Center).Bold(true)).
		Filtered(true)

	return newTable
}

// resizeView will change the dimensions of the table in order to accommodate
// the new window dimensions represented by the given tea.WindowSizeMsg
func (m tableModel) resizeView(msg tea.WindowSizeMsg) tableModel {
	// handle width changes
	m.table = m.table.WithMaxTotalWidth(msg.Width)
	m.tableWidth = msg.Width

	// handle height changes
	if headerAndFooterPadding >= msg.Height {
		// height too short to fit rows
		m.table = m.table.WithPageSize(0)
	} else {
		newRowsPerPage := msg.Height - headerAndFooterPadding
		m.table = m.table.WithPageSize(newRowsPerPage)
	}

	return m
}

// updateFooter updates the page and controls string in the table footer
func (m tableModel) updateFooter() tableModel {
	controlsStr := tableControls

	// prevent controls text from wrapping to avoid table misprints
	pageStr := fmt.Sprintf("Page: %d/%d | ", m.table.CurrentPage(), m.table.MaxPages())
	if m.tableWidth < len(pageStr)+len(controlsStr) {
		controlsWidth := m.tableWidth - len(ellipses) - len(pageStr) - 2
		if controlsWidth < 0 {
			controlsWidth = 0
		} else if controlsWidth > len(tableControls) {
			controlsWidth = len(tableControls)
		}
		controlsStr = tableControls[0:controlsWidth] + ellipses
	}

	renderedControls := lipgloss.NewStyle().Faint(true).Render(controlsStr)
	footerStr := fmt.Sprintf("%s%s", pageStr, renderedControls)
	m.table = m.table.WithStaticFooter(footerStr)

	return m
}

// update updates the state of the tableModel
func (m tableModel) update(msg tea.Msg) (tableModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// update filtering input field
		if m.filterTextInput.Focused() {
			var cmd tea.Cmd
			if msg.String() == "enter" || msg.String() == "esc" {
				// exit filter input and update controls string
				m.filterTextInput.Blur()
				m = m.updateFooter()
			} else {
				m.filterTextInput, cmd = m.filterTextInput.Update(msg)
			}

			m.table = m.table.WithFilterInput(m.filterTextInput)
			return m, cmd
		}

		// listen for specific inputs
		switch msg.String() {
		case "f":
			// focus filter input field
			m.filterTextInput.Focus()
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	// update footer
	m = m.updateFooter()

	return m, cmd
}

// view returns a string representing the table view
func (m tableModel) view() string {
	outputStr := strings.Builder{}

	outputStr.WriteString(m.table.View())
	outputStr.WriteString("\n")

	if m.filterTextInput.Value() != "" || m.filterTextInput.Focused() {
		outputStr.WriteString(m.filterTextInput.View())
		outputStr.WriteString("\n")
	}

	return outputStr.String()
}
