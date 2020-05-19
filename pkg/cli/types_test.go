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

func TestIntMe(t *testing.T) {
	cli := getTestCLI()
	intVal := 10
	val := cli.IntMe(intVal)
	h.Assert(t, *val == intVal, "Should return %s from passed in int value", intVal)
	val = cli.IntMe(&intVal)
	h.Assert(t, *val == intVal, "Should return %s from passed in int pointer", intVal)
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
