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

// NewSorter creates a new Sorter object to be used to sort the given instance types
// based on the sorting field and direction
// TODO: explain how sorting field is a JSON path to the appropriate Details struct property
// TODO: explain valid sort directions
// TODO: maybe make the strings pointers instead so that the CLI can pass the flags
// in directly (also can have nil checks in here and return appropriate errors)
// TODO: maybe instead of "sortField" call it something to do with "JSON path"/"path"
func NewSorter(instanceTypes []*instancetypes.Details, sortField string, sortDirection string) (*Sorter, error) {
	// TODO: determine if sortField is valid. Maybe do this in newSorterNode because
	// the json path library we are using already validates the sortField

	// validate direction flag and determine sorting direction
	// if sortDirection == nil {
	// 	return nil, fmt.Errorf("sort direction is nil")
	// }
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
	// TODO: figure out if there is a better way to get correct format than to
	// marshal and then unmarshal instance types

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

	// TODO:
	// Bug where one value might be a "float" but then another might be nil
	// and we need some way to compare those two. Can't just set the value
	// of float to nil and can't make nil be a pointer because then it
	// will try to compare float with a nil

	// maybe just do a "if valj != reflect.float" when vali is, then treat
	// case where valj is nil
	// This does not work.........

	// maybe turn everything into pointers?

	// Maybe if we knew the type of the field we could use that to determine
	// which of the types is correct? AKA if we have a type that is a float
	// switch on that type instead of the type of valI and then in
	// the case we can have an if where if the val does not equal the type,
	// then it is not less (AKA it bubbles up)

	// if result == nil {
	// 	var ptr *interface{} = nil
	// 	result = ptr
	// }

	return &sorterNode{
		instanceType: instanceType,
		fieldValue:   reflect.ValueOf(result),
	}, nil
}

// Sort the instance types in the Sorter based on the Sorter's sort field and
// direction
func (s *Sorter) Sort() {
	sort.Sort(s)
}

func (s *Sorter) Len() int {
	return len(s.sorters)
}

func (s *Sorter) Swap(i, j int) {
	// originalI := s.sorters[i]

	// s.sorters[i] = s.sorters[j]
	// s.sorters[j] = originalI

	s.sorters[i], s.sorters[j] = s.sorters[j], s.sorters[i]
}

func (s *Sorter) Less(i, j int) bool {
	valI := s.sorters[i].fieldValue
	valJ := s.sorters[j].fieldValue

	fmt.Println("CALLING IS LESS")
	less, _ := isLess(valI, valJ, s.isDescending)

	return less
}

// isLess determines whether the first value (valI) is less than the
// second value (valJ) or not
func isLess(valI, valJ reflect.Value, isDescending bool) (bool, error) {
	// TODO: add more types
	switch valI.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Println("=====found int type!======")

		if isDescending {
			return valI.Int() > valJ.Int(), nil
		} else {
			return valI.Int() <= valJ.Int(), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Println("=====found uint type!======")

		if isDescending {
			return valI.Uint() > valJ.Uint(), nil
		} else {
			return valI.Uint() <= valJ.Uint(), nil
		}
	case reflect.Float32, reflect.Float64:
		fmt.Printf("=====found float type!====== values: %f, %f\n")
		//fmt.Printf("        Result: %b\n", (valI.Float() < valJ.Float()))

		if valJ.Kind() != reflect.Float32 && valJ.Kind() != reflect.Float64 {
			fmt.Println("valJ is not a float")
			return true, nil
		}

		if isDescending {
			return valI.Float() > valJ.Float(), nil
		} else {
			return valI.Float() <= valJ.Float(), nil
		}
	case reflect.String:
		fmt.Println("=====found string type!======")

		if isDescending {
			return strings.Compare(valI.String(), valJ.String()) > 0, nil
		} else {
			return strings.Compare(valI.String(), valJ.String()) <= 0, nil
		}
	case reflect.Pointer:
		fmt.Println("=====found ptr type!======")

		// Handle nil values by making non nil values always less than the nil values. That way the
		// nil values can be bubbled up to the end of the list.
		if valI.IsNil() {
			return false, nil
		} else if valJ.Kind() == reflect.Pointer && valJ.IsNil() {
			return true, nil
		}

		return isLess(valI.Elem(), valJ.Elem(), isDescending)
	case reflect.Invalid:
		// handle invalid values (like nil values) by making valid values
		// always less than the invalid values. That way the invalid values
		// always bubble up to the end of the list
		return false, nil
	default:
		fmt.Printf("====unsortable type!==== %v\n", valI.Kind())

		// TODO: log that an unsortable type has been passed

		// handle unsortable values (like nil values) by making sortable values
		// always less than the unsortable values. That way the unsortable values
		// always bubble up to the end of the list
		return false, nil
	}

	// TODO: add bool types
}

// InstanceTypes returns the list of instance types held in the Sorter
func (s *Sorter) InstanceTypes() []*instancetypes.Details {
	instanceTypes := []*instancetypes.Details{}

	for _, node := range s.sorters {
		instanceTypes = append(instanceTypes, node.instanceType)
	}

	return instanceTypes
}
