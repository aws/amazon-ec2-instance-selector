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
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

const (
	// can't get terminal dimensions on startup, so use this
	initialDimensionVal = 30
)

const (
	// table states
	stateTable   = "table"
	stateVerbose = "verbose"
)

// BubbleTeaModel is used to hold the state of the bubble tea TUI
type BubbleTeaModel struct {
	// holds the output currentState of the model
	currentState string

	// the model for the table view
	tableModel tableModel

	// holds state for the verbose view
	verboseModel verboseModel
}

// NewBubbleTeaModel initializes a new bubble tea Model which represents
// a stylized table to display instance types
func NewBubbleTeaModel(instanceTypes []*instancetypes.Details) BubbleTeaModel {
	return BubbleTeaModel{
		currentState: stateTable,
		tableModel:   *initTableModel(instanceTypes),
		verboseModel: *initVerboseModel(instanceTypes),
	}
}

// Init is used by bubble tea to initialize a bubble tea table
func (m BubbleTeaModel) Init() tea.Cmd {
	return nil
}

// Update is used by bubble tea to update the state of the bubble
// tea model based on user input
func (m BubbleTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// check for quit or change in state
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "e":
			switch m.currentState {
			case stateTable:
				// don't change state if using text input
				if m.tableModel.filterTextInput.Focused() {
					break
				}

				// switch from table state to verbose state
				m.currentState = stateVerbose

				// get focused instance type
				rowIndex := m.tableModel.table.GetHighlightedRowIndex()
				focusedInstance := m.verboseModel.instanceTypes[rowIndex]

				// set content of view
				m.verboseModel.focusedInstanceName = focusedInstance.InstanceType
				m.verboseModel.viewport.SetContent(VerboseInstanceTypeOutput([]*instancetypes.Details{focusedInstance})[0])

				// move viewport to top of printout
				m.verboseModel.viewport.SetYOffset(0)
			case stateVerbose:
				// switch from verbose state to table state
				m.currentState = stateTable
			}
		}
	case tea.WindowSizeMsg:
		// This is needed to handle a bug with bubble tea
		// where resizing causes misprints (https://github.com/Evertras/bubble-table/issues/121)
		termenv.ClearScreen()

		// handle screen resizing
		m.tableModel = m.tableModel.resizeView(msg)
		m.verboseModel = m.verboseModel.resizeView(msg)
	}

	switch m.currentState {
	case stateTable:
		// update table
		var cmd tea.Cmd
		m.tableModel, cmd = m.tableModel.update(msg)

		return m, cmd
	case stateVerbose:
		// update viewport
		var cmd tea.Cmd
		m.verboseModel, cmd = m.verboseModel.update(msg)
		return m, cmd
	}

	return m, nil
}

// View is used by bubble tea to render the bubble tea model
func (m BubbleTeaModel) View() string {
	switch m.currentState {
	case stateTable:
		return m.tableModel.view()
	case stateVerbose:
		return m.verboseModel.view()
	}

	return ""
}
