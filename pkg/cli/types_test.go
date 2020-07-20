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

package cli_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/pkg/test"
)

// Tests

func TestBoolMe(t *testing.T) {
	cli := getTestCLI()
	boolTrue := true
	val := cli.BoolMe(boolTrue)
	h.Assert(t, *val == true, "Should return true from passed in value bool")
	val = cli.BoolMe(&boolTrue)
	h.Assert(t, *val == true, "Should return true from passed in pointer bool")
	val = cli.BoolMe(7)
	h.Assert(t, val == nil, "Should return nil from other data type passed in")
	val = cli.BoolMe(nil)
	h.Assert(t, val == nil, "Should return nil if nil is passed in")
}

func TestStringMe(t *testing.T) {
	cli := getTestCLI()
	stringVal := "test"
	val := cli.StringMe(stringVal)
	h.Assert(t, *val == stringVal, "Should return %s from passed in string value", stringVal)
	val = cli.StringMe(&stringVal)
	h.Assert(t, *val == stringVal, "Should return %s from passed in string pointer", stringVal)
	val = cli.StringMe(7)
	h.Assert(t, val == nil, "Should return nil from other data type passed in")
	val = cli.StringMe(nil)
	h.Assert(t, val == nil, "Should return nil if nil is passed in")
}

func TestStringSliceMe(t *testing.T) {
	cli := getTestCLI()
	stringSliceVal := []string{"test"}
	val := cli.StringSliceMe(stringSliceVal)
	h.Assert(t, reflect.DeepEqual(*val, stringSliceVal), "Should return %s from passed in string slice value", stringSliceVal)
	val = cli.StringSliceMe(&stringSliceVal)
	h.Assert(t, reflect.DeepEqual(*val, stringSliceVal), "Should return %s from passed in string slicepointer", stringSliceVal)
	val = cli.StringSliceMe(7)
	h.Assert(t, val == nil, "Should return nil from other data type passed in")
	val = cli.StringSliceMe(nil)
	h.Assert(t, val == nil, "Should return nil if nil is passed in")
}

func TestIntMe(t *testing.T) {
	cli := getTestCLI()
	intVal := 10
	int32Val := int32(intVal)
	val := cli.IntMe(intVal)
	h.Assert(t, *val == intVal, "Should return %s from passed in int value", intVal)
	val = cli.IntMe(&intVal)
	h.Assert(t, *val == intVal, "Should return %s from passed in int pointer", intVal)
	val = cli.IntMe(int32Val)
	h.Assert(t, *val == intVal, "Should return %s from passed in int32 value", intVal)
	val = cli.IntMe(&int32Val)
	h.Assert(t, *val == intVal, "Should return %s from passed in int32 pointer", intVal)
	val = cli.IntMe(true)
	h.Assert(t, val == nil, "Should return nil from other data type passed in")
	val = cli.IntMe(nil)
	h.Assert(t, val == nil, "Should return nil if nil is passed in")
}

func TestFloat64Me(t *testing.T) {
	cli := getTestCLI()
	fVal := 10.01
	val := cli.Float64Me(fVal)
	h.Assert(t, *val == fVal, "Should return %s from passed in float64 value", fVal)
	val = cli.Float64Me(&fVal)
	h.Assert(t, *val == fVal, "Should return %s from passed in float64 pointer", fVal)
	val = cli.Float64Me(true)
	h.Assert(t, val == nil, "Should return nil from other data type passed in")
	val = cli.Float64Me(nil)
	h.Assert(t, val == nil, "Should return nil if nil is passed in")
}

func TestIntRangeMe(t *testing.T) {
	cli := getTestCLI()
	intRangeVal := selector.IntRangeFilter{LowerBound: 1, UpperBound: 2}
	val := cli.IntRangeMe(intRangeVal)
	h.Assert(t, *val == intRangeVal, "Should return %s from passed in int range value", intRangeVal)
	val = cli.IntRangeMe(&intRangeVal)
	h.Assert(t, *val == intRangeVal, "Should return %s from passed in range pointer", intRangeVal)
	val = cli.IntRangeMe(true)
	h.Assert(t, val == nil, "Should return nil from other data type passed in")
	val = cli.IntRangeMe(nil)
	h.Assert(t, val == nil, "Should return nil if nil is passed in")
}

func TestFloat64RangeMe(t *testing.T) {
	cli := getTestCLI()
	float64RangeVal := selector.Float64RangeFilter{LowerBound: 1.0, UpperBound: 2.0}
	val := cli.Float64RangeMe(float64RangeVal)
	h.Assert(t, *val == float64RangeVal, "Should return %s from passed in float64 range value", float64RangeVal)
	val = cli.Float64RangeMe(&float64RangeVal)
	h.Assert(t, *val == float64RangeVal, "Should return %s from passed in range pointer", float64RangeVal)
	val = cli.Float64RangeMe(true)
	h.Assert(t, val == nil, "Should return nil from other data type passed in")
	val = cli.Float64RangeMe(nil)
	h.Assert(t, val == nil, "Should return nil if nil is passed in")
}
func TestRegexMe(t *testing.T) {
	cli := getTestCLI()
	regexVal, err := regexp.Compile("c4.*")
	h.Ok(t, err)
	val := cli.RegexMe(*regexVal)
	h.Assert(t, val.String() == regexVal.String(), "Should return %s from passed in regex value", regexVal)
	val = cli.RegexMe(regexVal)
	h.Assert(t, val.String() == regexVal.String(), "Should return %s from passed in regex pointer", regexVal)
	val = cli.RegexMe(true)
	h.Assert(t, val == nil, "Should return nil from other data type passed in")
	val = cli.RegexMe(nil)
	h.Assert(t, val == nil, "Should return nil if nil is passed in")
}
