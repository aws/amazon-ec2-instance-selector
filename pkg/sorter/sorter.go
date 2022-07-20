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

package sorter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/oliveagle/jsonpath"
)

const (
	sortAscending  = "ascending"
	sortAsc        = "asc"
	sortDescending = "descending"
	sortDesc       = "desc"
)

// sorterNode represents a sortable instance type which holds the value
// to sort by instance sort
type sorterNode struct {
	instanceType *instancetypes.Details
	fieldValue   reflect.Value
}

// Sorter is used to sort instance types based on a sorting field
// and direction
type Sorter struct {
	sorters      []*sorterNode
	sortField    string
	isDescending bool
}

// sorterWrapper is used to internally hide the implementation of the
// sorting interface from users
type sorterWrapper struct {
	sorter *Sorter
	err    error
}

// NewSorter creates a new Sorter object to be used to sort the given instance types
// based on the sorting field and direction
//
// sortField is a JSON path to a field in the instancetypes.Details struct which represents
// the field to sort instance types by. JSON path must start with "$" character (Ex: "$.MemoryInfo.SizeInMiB").
//
// sortDirection represents the direction to sort in. Valid options: "ascending", "asc", "descending", "desc".
func NewSorter(instanceTypes []*instancetypes.Details, sortField string, sortDirection string) (*Sorter, error) {
	var isDescending bool
	switch sortDirection {
	case sortDescending, sortDesc:
		isDescending = true
	case sortAscending, sortAsc:
		isDescending = false
	default:
		return nil, fmt.Errorf("invalid sort direction: %s (valid options: %s, %s, %s, %s)", sortDirection, sortAscending, sortAsc, sortDescending, sortDesc)
	}

	// Create sorterNode objects for each instance type
	sorters := []*sorterNode{}
	for _, instanceType := range instanceTypes {
		newSorter, err := newSorterNode(instanceType, sortField)
		if err != nil {
			return nil, fmt.Errorf("error creating sorting node: %v", err)
		}

		sorters = append(sorters, newSorter)
	}

	return &Sorter{
		sorters:      sorters,
		sortField:    sortField,
		isDescending: isDescending,
	}, nil
}

// newSorterNode creates a new sorterNode object which represents the given instance type
// and can be used in sorting of instance types based on the given sortField
func newSorterNode(instanceType *instancetypes.Details, sortField string) (*sorterNode, error) {
	// convert instance type into json
	jsonInstanceType, err := json.Marshal(instanceType)
	if err != nil {
		return nil, err
	}

	// unmarshal json instance types in order to get proper format
	// for json path parsing
	var jsonData interface{}
	err = json.Unmarshal(jsonInstanceType, &jsonData)
	if err != nil {
		return nil, err
	}

	// get the desired field from the json data based on the passed in
	// json path
	result, err := jsonpath.JsonPathLookup(jsonData, sortField)
	if err != nil {
		return nil, err
	}

	return &sorterNode{
		instanceType: instanceType,
		fieldValue:   reflect.ValueOf(result),
	}, nil
}

// Sort the instance types in the Sorter based on the Sorter's sort field and
// direction
func (s *Sorter) Sort() error {
	sortWrapper := sorterWrapper{
		sorter: s,
		err:    nil,
	}

	sort.Sort(&sortWrapper)

	return sortWrapper.err
}

// Len returns the number of sorter nodes in the Sorter
func (sw *sorterWrapper) Len() int {
	s := sw.sorter
	return len(s.sorters)
}

// Swap swaps the positions of the sorter nodes at indices i and j
func (sw *sorterWrapper) Swap(i, j int) {
	s := sw.sorter
	s.sorters[i], s.sorters[j] = s.sorters[j], s.sorters[i]
}

// Less determines whether the value of the sorter node at index i
// is less than the value of the sorter node at index j or not
func (sw *sorterWrapper) Less(i, j int) bool {
	s := sw.sorter
	valI := s.sorters[i].fieldValue
	valJ := s.sorters[j].fieldValue

	less, err := isLess(valI, valJ, s.isDescending)
	if err != nil {
		sw.err = err
	}

	return less
}

// isLess determines whether the first value (valI) is less than the
// second value (valJ) or not
func isLess(valI, valJ reflect.Value, isDescending bool) (bool, error) {
	switch valI.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// if valJ is not an int (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		vaJKind := valJ.Kind()
		if vaJKind != reflect.Int && vaJKind != reflect.Int8 && vaJKind != reflect.Int16 && vaJKind != reflect.Int32 && vaJKind != reflect.Int64 {
			return true, nil
		}

		if isDescending {
			return valI.Int() > valJ.Int(), nil
		} else {
			return valI.Int() <= valJ.Int(), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// if valJ is not a uint (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		vaJKind := valJ.Kind()
		if vaJKind != reflect.Uint && vaJKind != reflect.Uint8 && vaJKind != reflect.Uint16 && vaJKind != reflect.Uint32 && vaJKind != reflect.Uint64 {
			return true, nil
		}

		if isDescending {
			return valI.Uint() > valJ.Uint(), nil
		} else {
			return valI.Uint() <= valJ.Uint(), nil
		}
	case reflect.Float32, reflect.Float64:
		// if valJ is not a float (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		vaJKind := valJ.Kind()
		if vaJKind != reflect.Float32 && vaJKind != reflect.Float64 {
			return true, nil
		}

		if isDescending {
			return valI.Float() > valJ.Float(), nil
		} else {
			return valI.Float() <= valJ.Float(), nil
		}
	case reflect.String:
		// if valJ is not a string (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		if valJ.Kind() != reflect.String {
			return true, nil
		}

		if isDescending {
			return strings.Compare(valI.String(), valJ.String()) > 0, nil
		} else {
			return strings.Compare(valI.String(), valJ.String()) <= 0, nil
		}
	case reflect.Pointer:
		// Handle nil values by making non nil values always less than the nil values. That way the
		// nil values can be bubbled up to the end of the list.
		if valI.IsNil() {
			return false, nil
		} else if valJ.Kind() != reflect.Pointer || valJ.IsNil() {
			return true, nil
		}

		return isLess(valI.Elem(), valJ.Elem(), isDescending)
	case reflect.Bool:
		// if valJ is not a bool (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		if valJ.Kind() != reflect.Bool {
			return true, nil
		}

		if isDescending {
			return !valI.Bool(), nil
		} else {
			return valI.Bool(), nil
		}
	case reflect.Invalid:
		// handle invalid values (like nil values) by making valid values
		// always less than the invalid values. That way the invalid values
		// always bubble up to the end of the list
		return false, nil
	default:
		// unsortable value
		return false, fmt.Errorf("unsortable value")
	}
}

// InstanceTypes returns the list of instance types held in the Sorter
func (s *Sorter) InstanceTypes() []*instancetypes.Details {
	instanceTypes := []*instancetypes.Details{}

	for _, node := range s.sorters {
		instanceTypes = append(instanceTypes, node.instanceType)
	}

	return instanceTypes
}
