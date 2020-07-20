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
	"fmt"
	"testing"

	h "github.com/aws/amazon-ec2-instance-selector/pkg/test"
)

// Tests

func TestBoolFlag(t *testing.T) {
	cli := getTestCLI()
	for _, flagFn := range []func(string, *string, *bool, string){cli.BoolFlag, cli.ConfigBoolFlag, cli.SuiteBoolFlag} {
		flagName := "test-int"
		flagFn(flagName, cli.StringMe("t"), nil, "Test Bool")
		_, ok := cli.Flags[flagName]
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag", flagName)

		cli = getTestCLI()
		cli.BoolFlag(flagName, nil, nil, "Test Bool")
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag w/ no shorthand")
		h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
	}
}

func TestIntFlag(t *testing.T) {
	cli := getTestCLI()
	for _, flagFn := range []func(string, *string, *int, string){cli.IntFlag, cli.ConfigIntFlag} {
		flagName := "test-int"
		flagFn(flagName, cli.StringMe("t"), nil, "Test Int")
		_, ok := cli.Flags[flagName]
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag", flagName)

		cli = getTestCLI()
		cli.IntFlag(flagName, nil, nil, "Test Int")
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag w/ no shorthand")
		h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
	}
}

func TestStringFlag(t *testing.T) {
	cli := getTestCLI()
	for _, flagFn := range []func(string, *string, *string, string, func(interface{}) error){cli.StringFlag, cli.ConfigStringFlag, cli.SuiteStringFlag} {
		flagName := "test-string"
		flagFn(flagName, cli.StringMe("t"), nil, "Test String", nil)
		_, ok := cli.Flags[flagName]
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag", flagName)

		cli = getTestCLI()
		flagFn(flagName, cli.StringMe("t"), nil, "Test String w/ validation", func(val interface{}) error {
			return fmt.Errorf("validation failed")
		})
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag with validation", flagName)

		cli = getTestCLI()
		flagFn(flagName, nil, nil, "Test String", nil)
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag w/ no shorthand")
		h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
	}
}

func TestStringOptionsFlag(t *testing.T) {
	cli := getTestCLI()
	for _, flagFn := range []func(string, *string, *string, string, []string){cli.StringOptionsFlag, cli.ConfigStringOptionsFlag, cli.SuiteStringOptionsFlag} {
		flagName := "test-string"
		flagFn(flagName, cli.StringMe("t"), nil, "Test String", nil)
		_, ok := cli.Flags[flagName]
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag", flagName)

		cli = getTestCLI()
		flagFn(flagName, cli.StringMe("t"), nil, "Test String w/ options", []string{"opt1", "opt2"})
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag with options", flagName)

		cli = getTestCLI()
		flagFn(flagName, nil, nil, "Test String", []string{"opt1", "opt2"})
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag w/ no shorthand")
		h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
	}
}

func TestStringSliceFlag(t *testing.T) {
	cli := getTestCLI()
	for _, flagFn := range []func(string, *string, []string, string){cli.StringSliceFlag, cli.ConfigStringSliceFlag, cli.SuiteStringSliceFlag} {
		flagName := "test-string-slice"
		flagFn(flagName, cli.StringMe("t"), nil, "Test String Slice")
		_, ok := cli.Flags[flagName]
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag", flagName)

		cli = getTestCLI()
		flagFn(flagName, cli.StringMe("t"), []string{"def1", "def2"}, "Test String w/ slice default")
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag with default slice", flagName)

		cli = getTestCLI()
		flagFn(flagName, nil, []string{"def1", "def2"}, "Test String Slice")
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag w/ no shorthand")
		h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
	}
}

func TestRatioFlag(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-ratio"
	cli.RatioFlag(flagName, cli.StringMe("t"), nil, "Test Ratio")
	_, ok := cli.Flags[flagName]
	h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
	h.Assert(t, ok, "Should contain %s flag", flagName)

	cli = getTestCLI()
	cli.RatioFlag(flagName, nil, nil, "Test Ratio")
	h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag w/ no shorthand")
	h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
}

func TestIntMinMaxRangeFlags(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-int-min-max-range"
	cli.IntMinMaxRangeFlags(flagName, cli.StringMe("t"), nil, "Test Min Max Range")
	_, ok := cli.Flags[flagName]
	_, minOk := cli.Flags[flagName+"-min"]
	_, maxOk := cli.Flags[flagName+"-max"]
	h.Assert(t, len(cli.Flags) == 3, "Should contain 3 flags")
	h.Assert(t, ok, "Should contain %s flag", flagName)
	h.Assert(t, minOk, "Should contain %s flag", flagName)
	h.Assert(t, maxOk, "Should contain %s flag", flagName)

	cli = getTestCLI()
	cli.IntMinMaxRangeFlags(flagName, nil, nil, "Test Min Max Range")
	h.Assert(t, len(cli.Flags) == 3, "Should contain 3 flags w/ no shorthand")
	h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
}

func TestFloat64MinMaxRangeFlags(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-float-min-max-range"
	cli.Float64MinMaxRangeFlags(flagName, cli.StringMe("t"), nil, "Test Min Max Range")
	_, ok := cli.Flags[flagName]
	_, minOk := cli.Flags[flagName+"-min"]
	_, maxOk := cli.Flags[flagName+"-max"]
	h.Assert(t, len(cli.Flags) == 3, "Should contain 3 flags")
	h.Assert(t, ok, "Should contain %s flag", flagName)
	h.Assert(t, minOk, "Should contain %s flag", flagName)
	h.Assert(t, maxOk, "Should contain %s flag", flagName)

	cli = getTestCLI()
	cli.Float64MinMaxRangeFlags(flagName, nil, nil, "Test Min Max Range")
	h.Assert(t, len(cli.Flags) == 3, "Should contain 3 flags w/ no shorthand")
	h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
}

func TestRegexFlag(t *testing.T) {
	cli := getTestCLI()
	for _, flagFn := range []func(string, *string, *string, string){cli.RegexFlag} {
		flagName := "test-regex"
		flagFn(flagName, cli.StringMe("t"), nil, "Test Regex")
		_, ok := cli.Flags[flagName]
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag")
		h.Assert(t, ok, "Should contain %s flag", flagName)

		cli = getTestCLI()
		flagFn(flagName, nil, nil, "Test Regex")
		h.Assert(t, len(cli.Flags) == 1, "Should contain 1 flag w/ no shorthand")
		h.Assert(t, ok, "Should contain %s flag w/ no shorthand", flagName)
	}
}
