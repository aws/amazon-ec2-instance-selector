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
	"io"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/sorter"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	// formatting
	sortingTitlePadding  = 3
	sortingFooterPadding = 2

	// shorthand flags
	gpus                           = "gpus"
	inferenceAccelerators          = "inference-accelerators"
	vcpus                          = "vcpus"
	memory                         = "memory"
	gpuMemoryTotal                 = "gpu-memory-total"
	networkInterfaces              = "network-interfaces"
	spotPrice                      = "spot-price"
	odPrice                        = "on-demand-price"
	instanceStorage                = "instance-storage"
	ebsOptimizedBaselineBandwidth  = "ebs-optimized-baseline-bandwidth"
	ebsOptimizedBaselineThroughput = "ebs-optimized-baseline-throughput"
	ebsOptimizedBaselineIOPS       = "ebs-optimized-baseline-iops"

	// controls
	sortingListControls = "Controls: ↑/↓ - up/down • esc - return to table • enter - select filter • q - quit"
	sortingTextControls = "Controls: ↑/↓ - up/down • enter - enter json path"
)

// sortingModel holds the state for the sorting view
type sortingModel struct {
	// list which holds the available shorting shorthands
	shorthandList list.Model

	// text input for json paths
	sortTextInput textinput.Model

	instanceTypes []*instancetypes.Details
}

// list format styles
var (
	listTitleStyle    = lipgloss.NewStyle().MarginLeft(2)
	listItemStyle     = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

// implement Item interface for list
type item string

func (i item) FilterValue() string { return "" }
func (i item) Title() string       { return string(i) }
func (i item) Description() string { return "" }

// implement ItemDelegate for list
type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := listItemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprintf(w, fn(str))
}

// initSortingView initializes and returns a new tableModel based on the given
// instance type details
func initSortingView(instanceTypes []*instancetypes.Details) *sortingModel {
	shorthandList := list.New(*createListItems(), itemDelegate{}, initialDimensionVal, initialDimensionVal)
	shorthandList.Title = "Select sorting filter:"
	shorthandList.Styles.Title = listTitleStyle
	shorthandList.SetFilteringEnabled(false)
	shorthandList.SetShowStatusBar(false)
	shorthandList.SetShowHelp(false)
	shorthandList.SetShowPagination(false)
	shorthandList.KeyMap = createListKeyMap()

	sortTextInput := textinput.New()
	sortTextInput.Prompt = "JSON Path: "
	sortTextInput.PromptStyle = lipgloss.NewStyle().Bold(true)

	return &sortingModel{
		shorthandList: shorthandList,
		sortTextInput: sortTextInput,
		instanceTypes: instanceTypes,
	}
}

// createListKeyMap creates a KeyMap with the controls for the shorthand list
func createListKeyMap() list.KeyMap {
	return list.KeyMap{
		CursorDown: key.NewBinding(
			key.WithKeys("down"),
		),
		CursorUp: key.NewBinding(
			key.WithKeys("up"),
		),
	}
}

// createListItems creates a list item for shorthand sorting flag
func createListItems() *[]list.Item {
	shorthandFlags := []string{
		gpus,
		inferenceAccelerators,
		vcpus,
		memory,
		gpuMemoryTotal,
		networkInterfaces,
		spotPrice,
		odPrice,
		instanceStorage,
		ebsOptimizedBaselineBandwidth,
		ebsOptimizedBaselineThroughput,
		ebsOptimizedBaselineIOPS,
	}

	items := []list.Item{}

	for _, flag := range shorthandFlags {
		items = append(items, item(flag))
	}

	return &items
}

// resizeSortingView will change the dimensions of the sorting view
// in order to accommodate the new window dimensions represented by
// the given tea.WindowSizeMsg
func (m sortingModel) resizeView(msg tea.WindowSizeMsg) sortingModel {
	shorthandList := &m.shorthandList
	shorthandList.SetWidth(msg.Width)
	// ensure that text input is right below last option
	if msg.Height >= len(shorthandList.Items())+sortingTitlePadding+sortingFooterPadding {
		shorthandList.SetHeight(len(shorthandList.Items()) + sortingTitlePadding)
	} else if msg.Height-sortingFooterPadding > 0 {
		shorthandList.SetHeight(msg.Height - sortingFooterPadding)
	} else {
		shorthandList.SetHeight(1)
	}

	// ensure cursor of list is still hidden after resize
	if m.sortTextInput.Focused() {
		shorthandList.Select(len(m.shorthandList.Items()))
	}

	m.shorthandList = *shorthandList

	return m
}

// sortTable sorts the table based on the sorting direction and sorting filter
// TODO: maybe move this to tableView and then call it in bubble tea when change of state occurs
func (m sortingModel) sortTable(model BubbleTeaModel, sortFilter string, sortDirection string) (table.Model, error) {
	instanceTypes, err := sorter.Sort(m.instanceTypes, sortFilter, sortDirection)
	if err != nil {
		return model.tableModel.table, err
	}

	// TODO: maybe ensure dimensions are the same as old table before returning
	// ensure truncation still occurs by maybe storing selected items and then doing the truncation process here?
	// ensure filtering is reapplied

	// TODO: To ensure that selected items remain after sorting, perhaps map instancetype to rows. That way
	// we can quickly sort rows based on the returned sorted instance types. This way we can maybe also avoid
	// storing instance types in the sortingModel struct because we can have a loop that not only does the mapping
	// but also gives us a list of instance types from the visible rows

	return createTable(instanceTypes), nil
}

// update updates the state of the sortingModel
func (m sortingModel) update(msg tea.Msg) (sortingModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			if m.shorthandList.Index() == len(m.shorthandList.Items())-1 {
				// focus text input and hide cursor in shorthand list
				m.shorthandList.Select(len(m.shorthandList.Items()))
				m.sortTextInput.Focus()
			}
		case "up":
			if m.sortTextInput.Focused() {
				// go back to list from text input
				m.shorthandList.Select(len(m.shorthandList.Items()))
				m.sortTextInput.Blur()
			}
		}

		if m.sortTextInput.Focused() {
			m.sortTextInput, cmd = m.sortTextInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if !m.sortTextInput.Focused() {
		m.shorthandList, cmd = m.shorthandList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// view returns a string representing the sorting view
func (m sortingModel) view() string {
	outputStr := strings.Builder{}

	outputStr.WriteString(m.shorthandList.View())
	outputStr.WriteString("\n")

	outputStr.WriteString(m.sortTextInput.View())
	outputStr.WriteString("\n")

	if m.sortTextInput.Focused() {
		outputStr.WriteString(controlsStyle.Render(sortingTextControls))
	} else {
		outputStr.WriteString(controlsStyle.Render(sortingListControls))
	}

	return outputStr.String()
}
