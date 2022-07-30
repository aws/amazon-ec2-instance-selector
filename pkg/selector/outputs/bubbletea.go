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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/muesli/termenv"
)

const (
	// table formatting
	headerAndFooterPadding = 7
	headerPadding          = 2

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

	// calculate and fetch all column data from instance types
	columnsData := getWideColumnsData(instanceTypes)

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

// createColumns creates columns based on the column key constants
func createColumns() *[]table.Column {
	columns := []table.Column{}

	// iterate through wideColumnsData struct and create a new column for each field tag
	columnDataStruct := wideColumnsData{}
	structType := reflect.TypeOf(columnDataStruct)
	for i := 0; i < structType.NumField(); i++ {
		columnHeader := structType.Field(i).Tag.Get(columnTag)
		newCol := table.NewColumn(columnHeader, columnHeader, len(columnHeader)+headerPadding)

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
