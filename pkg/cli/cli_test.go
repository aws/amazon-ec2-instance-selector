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
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/cli"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/v3/pkg/test"
	"github.com/spf13/cobra"
)

const (
	maxInt = int(^uint(0) >> 1)
)

// Helpers

func getTestCLI() cli.CommandLineInterface {
	runFunc := func(cmd *cobra.Command, args []string) {}
	return cli.New("test", "test short usage", "test long usage", "test examples", runFunc)
}

// Tests

func TestValidateFlags(t *testing.T) {
	// Nil validator should succeed validation
	cli := getTestCLI()
	flagName := "test-flag"
	cli.StringFlag(flagName, nil, nil, "Test String w/o validation", nil)
	err := cli.ValidateFlags()
	h.Ok(t, err)

	// Validator which returns nil error should succeed validation
	cli = getTestCLI()
	flagName = "test-flag"
	cli.StringFlag(flagName, nil, nil, "Test String w/ successful validation", func(val interface{}) error {
		return nil
	})
	err = cli.ValidateFlags()
	h.Ok(t, err)

	// Validator which returns error should fail validation
	cli = getTestCLI()
	cli.StringFlag(flagName, nil, nil, "Test String w/ validation failure", func(val interface{}) error {
		return fmt.Errorf("validation failed")
	})
	err = cli.ValidateFlags()
	h.Nok(t, err)
}

func TestParseAndValidateFlags_Ratio(t *testing.T) {
	// Nil validator should succeed validation
	cli := getTestCLI()
	flagName := "test-ratio-flag"
	cli.RatioFlag(flagName, nil, nil, "Test Ratio")
	os.Args = []string{"", "--" + flagName, "1:2"}
	flags, err := cli.ParseAndValidateFlags()
	h.Ok(t, err)
	h.Assert(t, len(flags) == 1, "1 Flag should have been parsed and validated")

	// Validator which returns error should fail validation
	cli = getTestCLI()
	cli.RatioFlag(flagName, nil, nil, "Test Ratio w/ validation failure")
	os.Args = []string{"", "--" + flagName, "1"}
	_, err = cli.ParseAndValidateFlags()
	h.Nok(t, err)
}

func TestParseAndValidateFlags_StringOptions(t *testing.T) {
	// Nil validator should succeed validation
	cli := getTestCLI()
	flagName := "test-string-opts-flag"
	opts := []string{"opt1", "opt2"}
	cli.StringOptionsFlag(flagName, nil, nil, "Test String Options", opts)
	os.Args = []string{"", "--" + flagName, "opt1"}
	flags, err := cli.ParseAndValidateFlags()
	h.Ok(t, err)
	h.Assert(t, len(flags) == 1, "1 Flag should have been parsed and validated")

	// Validator which returns error should fail validation
	cli = getTestCLI()
	cli.StringOptionsFlag(flagName, nil, nil, "Test String Options w/ validation failure", opts)
	os.Args = []string{"", "--" + flagName, "opt55"}
	_, err = cli.ParseAndValidateFlags()
	h.Nok(t, err)
}

func TestParseFlags(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	flagArg := fmt.Sprintf("--%s", flagName)
	cli.StringFlag(flagName, nil, nil, "Test String w/o validation", nil)
	os.Args = []string{"ec2-instance-selector", flagArg, "test"}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	flagOutput := flags[flagName].(*string)
	h.Assert(t, *flagOutput == "test", "Flag %s should have been parsed", flagArg)
}

func TestParseFlags_IntRange(t *testing.T) {
	flagName := "test-flag"
	flagMinArg := fmt.Sprintf("%s-%s", flagName, "min")
	flagMaxArg := fmt.Sprintf("%s-%s", flagName, "max")
	flagArg := fmt.Sprintf("--%s", flagName)

	// Root set Min and Max to the same val
	cli := getTestCLI()
	cli.IntMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", flagArg, "5"}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput := flags[flagMinArg].(*int)
	flagMaxOutput := flags[flagMaxArg].(*int)
	h.Assert(t, *flagMinOutput == 5 && *flagMaxOutput == 5, "Flag %s min and max should have been parsed to the same number", flagArg)

	// Min is set to a val and max is set to maxInt
	cli = getTestCLI()
	cli.IntMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMinArg, "5"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*int)
	flagMaxOutput = flags[flagMaxArg].(*int)
	h.Assert(t, *flagMinOutput == 5 && *flagMaxOutput == maxInt, "Flag %s min should have been parsed from cmdline and max set to maxInt", flagArg)

	// Max is set to a val and min is set to 0
	cli = getTestCLI()
	cli.IntMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMaxArg, "50"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*int)
	flagMaxOutput = flags[flagMaxArg].(*int)
	h.Assert(t, *flagMinOutput == 0 && *flagMaxOutput == 50, "Flag %s max should have been parsed from cmdline and min set to 0", flagArg)

	// Min and Max are set to separate values
	cli = getTestCLI()
	cli.IntMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMinArg, "10", "--" + flagMaxArg, "500"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*int)
	flagMaxOutput = flags[flagMaxArg].(*int)
	h.Assert(t, *flagMinOutput == 10 && *flagMaxOutput == 500, "Flag %s min and max should have been parsed from cmdline", flagArg)
}

func TestParseFlags_IntRangeErr(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	flagArg := fmt.Sprintf("--%s", flagName)
	cli.IntMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", flagArg, "1", flagArg + "-min", "1", flagArg + "-max", "2"}
	_, err := cli.ParseFlags()
	h.Nok(t, err)
}

func TestParseFlags_ByteQuantityRange(t *testing.T) {
	flagName := "test-flag"
	flagMinArg := fmt.Sprintf("%s-%s", flagName, "min")
	flagMaxArg := fmt.Sprintf("%s-%s", flagName, "max")
	flagArg := fmt.Sprintf("--%s", flagName)

	// Root set Min and Max to the same val
	cli := getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", flagArg, "5"}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput := flags[flagMinArg].(*bytequantity.ByteQuantity)
	flagMaxOutput := flags[flagMaxArg].(*bytequantity.ByteQuantity)
	h.Assert(t, flagMinOutput.GiB() == 5.0 && flagMaxOutput.GiB() == 5.0, "Flag %s min and max should have been parsed to the same number", flagArg)

	// Min is set to a val and max is set to maxInt
	cli = getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMinArg, "5"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*bytequantity.ByteQuantity)
	flagMaxOutput = flags[flagMaxArg].(*bytequantity.ByteQuantity)
	h.Assert(t, flagMinOutput.GiB() == 5.0 && flagMaxOutput.Quantity == math.MaxUint64, "Flag %s min should have been parsed from cmdline and max set to maxInt", flagArg)

	// Max is set to a val and min is set to 0
	cli = getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMaxArg, "50"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*bytequantity.ByteQuantity)
	flagMaxOutput = flags[flagMaxArg].(*bytequantity.ByteQuantity)
	h.Assert(t, flagMinOutput.Quantity == 0 && flagMaxOutput.GiB() == 50.0, "Flag %s max should have been parsed from cmdline and min set to 0", flagArg)

	// Min and Max are set to separate values
	cli = getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMinArg, "10", "--" + flagMaxArg, "500"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*bytequantity.ByteQuantity)
	flagMaxOutput = flags[flagMaxArg].(*bytequantity.ByteQuantity)
	h.Assert(t, flagMinOutput.GiB() == 10.0 && flagMaxOutput.GiB() == 500.0, "Flag %s max and min should have been parsed from cmdline", flagArg)
	flagType := reflect.TypeOf(flags[flagName])
	bqRangeFilterType := reflect.TypeOf(&selector.ByteQuantityRangeFilter{})
	h.Assert(t, flagType == bqRangeFilterType, "%s should be of type %v, instead got %v", flagArg, bqRangeFilterType, flagType)
}

func TestParseFlags_Float64Range(t *testing.T) {
	flagName := "test-flag"
	flagMinArg := fmt.Sprintf("%s-%s", flagName, "min")
	flagMaxArg := fmt.Sprintf("%s-%s", flagName, "max")
	flagArg := fmt.Sprintf("--%s", flagName)

	// Root set Min and Max to the same val
	cli := getTestCLI()
	cli.Float64MinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", flagArg, "5.1"}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput := flags[flagMinArg].(*float64)
	flagMaxOutput := flags[flagMaxArg].(*float64)
	h.Assert(t, *flagMinOutput == 5.1 && *flagMaxOutput == 5.1, "Flag %s min and max should have been parsed to the same number", flagArg)

	// Min is set to a val and max is set to maxInt
	cli = getTestCLI()
	cli.Float64MinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMinArg, "5.1"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*float64)
	flagMaxOutput = flags[flagMaxArg].(*float64)
	h.Assert(t, *flagMinOutput == 5.1 && *flagMaxOutput == math.MaxFloat64, "Flag %s min should have been parsed from cmdline and max set to math.MaxFloat64", flagArg)

	// Max is set to a val and min is set to 0
	cli = getTestCLI()
	cli.Float64MinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMaxArg, "5.1"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*float64)
	flagMaxOutput = flags[flagMaxArg].(*float64)
	h.Assert(t, *flagMinOutput == 0.0 && *flagMaxOutput == 5.1, "Flag %s max should have been parsed from cmdline and min set to 0.0", flagArg)

	// Min and Max are set to separate values
	cli = getTestCLI()
	cli.Float64MinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMinArg, "10.1", "--" + flagMaxArg, "500.1"}
	flags, err = cli.ParseFlags()
	h.Ok(t, err)
	flagMinOutput = flags[flagMinArg].(*float64)
	flagMaxOutput = flags[flagMaxArg].(*float64)
	h.Assert(t, *flagMinOutput == 10.1 && *flagMaxOutput == 500.1, "Flag %s min and max should have been parsed from cmdline", flagArg)
}

func TestParseAndValidateFlags_ByteQuantityRange(t *testing.T) {
	flagName := "test-flag"
	flagMinArg := fmt.Sprintf("%s-%s", flagName, "min")
	flagMaxArg := fmt.Sprintf("%s-%s", flagName, "max")
	flagArg := fmt.Sprintf("--%s", flagName)

	// Root set Min and Max to the same val
	cli := getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", flagArg, "5"}
	_, err := cli.ParseAndValidateFlags()
	h.Ok(t, err)

	// Min is set to a val and max is set to maxInt
	cli = getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMinArg, "5"}
	_, err = cli.ParseAndValidateFlags()
	h.Ok(t, err)

	// Max is set to a val and min is set to 0
	cli = getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMaxArg, "50"}
	_, err = cli.ParseAndValidateFlags()
	h.Ok(t, err)

	// Min and Max are set to separate values
	cli = getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector", "--" + flagMinArg, "10", "--" + flagMaxArg, "500"}
	_, err = cli.ParseFlags()
	h.Ok(t, err)

	// no args
	cli = getTestCLI()
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test")
	os.Args = []string{"ec2-instance-selector"}
	_, err = cli.ParseAndValidateFlags()
	h.Ok(t, err)
}

func TestParseFlags_RootErr(t *testing.T) {
	cli := getTestCLI()
	os.Args = []string{"ec2-instance-selector", "--test", "test"}
	_, err := cli.ParseFlags()
	h.Nok(t, err)
}

func TestParseFlags_SuiteErr(t *testing.T) {
	cli := getTestCLI()
	cli.SuiteBoolFlag("test", nil, nil, "")
	os.Args = []string{"ec2-instance-selector", "-----test"}
	_, err := cli.ParseFlags()
	h.Nok(t, err)
}

func TestParseFlags_SuiteFlags(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	flagArg := fmt.Sprintf("--%s", flagName)
	cli.SuiteBoolFlag(flagName, nil, nil, "Test Suite Flag")
	os.Args = []string{"ec2-instance-selector", flagArg}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	flagOutput := flags[flagName].(*bool)
	h.Assert(t, *flagOutput == true, "Suite Flag %s should have been parsed", flagArg)
}

func TestParseFlags_ConfigFlags(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	flagArg := fmt.Sprintf("--%s", flagName)
	cli.ConfigBoolFlag(flagName, nil, nil, "Test Config Flag")
	os.Args = []string{"ec2-instance-selector", flagArg}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	flagOutput := flags[flagName].(*bool)
	h.Assert(t, *flagOutput == true, "Config Flag %s should have been parsed", flagArg)
}

func TestParseFlags_AllTypes(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	configName := flagName + "-config"
	suiteName := flagName + "-suite"
	flagArg := fmt.Sprintf("--%s", flagName)
	configArg := fmt.Sprintf("--%s", configName)
	suiteArg := fmt.Sprintf("--%s", suiteName)

	cli.BoolFlag(flagName, nil, nil, "Test Filter Flag")
	cli.ConfigBoolFlag(configName, nil, nil, "Test Config Flag")
	cli.SuiteBoolFlag(suiteName, nil, nil, "Test Suite Flag")
	os.Args = []string{"ec2-instance-selector", flagArg, configArg, suiteArg}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	flagOutput := flags[flagName].(*bool)
	configOutput := flags[configName].(*bool)
	suiteOutput := flags[suiteName].(*bool)
	h.Assert(t, *flagOutput && *configOutput && *suiteOutput, "Filter, Config, and Sutie Flags %s should have been parsed", flagArg)
}

func TestParseFlags_UntouchedFlags(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	flagArg := fmt.Sprintf("--%s", flagName)

	cli.BoolFlag(flagName, nil, nil, "Test Filter Flag")
	os.Args = []string{"ec2-instance-selector"}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	val, ok := flags[flagName]
	h.Assert(t, ok, "Flag %s should exist in flags map", flagArg)
	h.Assert(t, val == nil, "Flag %s should be set to nil when not explicitly set", flagArg)
}

func TestParseFlags_UntouchedFlagsAllTypes(t *testing.T) {
	cli := getTestCLI()
	intName := "int"
	ratioName := "ratio"
	byteQName := "bq"
	configName := "config"
	suiteName := "suite"

	cli.IntFlag(intName, nil, nil, "Test Filter Flag")
	cli.RatioFlag(ratioName, nil, nil, "Test Ratio Flag")
	cli.ByteQuantityFlag(byteQName, nil, nil, "Test Byte Quantity Flag")
	cli.ConfigStringFlag(configName, nil, nil, "Test Config Flag", nil)
	cli.SuiteBoolFlag(suiteName, nil, nil, "Test Suite Flag")

	os.Args = []string{"ec2-instance-selector"}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	for _, name := range []string{intName, ratioName, byteQName, configName, suiteName} {
		val, ok := flags[name]
		h.Assert(t, ok, "Flag %s should exist in flags map", "--"+name)
		h.Assert(t, val == nil, "Flag %s should be set to nil when not explicitly set", "--"+name)
	}
}

func TestParseAndValidateFlags_Err(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	flagArg := fmt.Sprintf("--%s", flagName)
	flagMin := flagArg + "-min"
	flagMax := flagArg + "-max"
	cli.IntMinMaxRangeFlags(flagName, nil, nil, "Test with validation")
	os.Args = []string{"ec2-instance-selector", flagMin, "5", flagMax, "1"}
	_, err := cli.ParseAndValidateFlags()
	h.Nok(t, err)
}

func TestParseAndValidateFlags_ByteQuantityErr(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	flagArg := fmt.Sprintf("--%s", flagName)
	flagMin := flagArg + "-min"
	flagMax := flagArg + "-max"
	cli.ByteQuantityMinMaxRangeFlags(flagName, nil, nil, "Test with validation")
	os.Args = []string{"ec2-instance-selector", flagMin, "5", flagMax, "1"}
	_, err := cli.ParseAndValidateFlags()
	h.Nok(t, err)
}

func TestParseAndValidateFlags(t *testing.T) {
	cli := getTestCLI()
	flagName := "test-flag"
	flagArg := fmt.Sprintf("--%s", flagName)
	flagMin := flagArg + "-min"
	flagMax := flagArg + "-max"
	cli.IntMinMaxRangeFlags(flagName, nil, nil, "Test with validation")
	os.Args = []string{"ec2-instance-selector", flagMin, "1", flagMax, "5"}
	flags, err := cli.ParseAndValidateFlags()
	h.Ok(t, err)
	flagType := reflect.TypeOf(flags[flagName])
	intRangeFilterType := reflect.TypeOf(&selector.IntRangeFilter{})
	h.Assert(t, flagType == intRangeFilterType, "%s should be of type %v, instead got %v", flagArg, intRangeFilterType, flagType)
}

func TestParseAndValidateRegexFlag(t *testing.T) {
	flagName := "test-regex-flag"
	flagArg := fmt.Sprintf("--%s", flagName)

	cli := getTestCLI()
	cli.RegexFlag(flagName, nil, nil, "Test with validation")
	os.Args = []string{"ec2-instance-selector", flagArg, "c4.*"}
	flags, err := cli.ParseAndValidateFlags()
	h.Ok(t, err)
	h.Assert(t, len(flags) == 1, "1 flag should have been processed")
	_, err = cli.ParseAndValidateFlags()
	h.Ok(t, err)

	cli = getTestCLI()
	cli.RegexFlag(flagName, nil, nil, "Test with validation")
	os.Args = []string{"ec2-instance-selector", flagArg, "(("}
	_, err = cli.ParseAndValidateFlags()
	h.Nok(t, err)
}

func TestParseAndValidateByteQuantityFlag(t *testing.T) {
	flagName := "test-bq-flag"
	flagArg := fmt.Sprintf("--%s", flagName)

	cli := getTestCLI()
	cli.ByteQuantityFlag(flagName, nil, nil, "Test with validation")
	os.Args = []string{"ec2-instance-selector", flagArg, "450"}
	flags, err := cli.ParseAndValidateFlags()
	h.Ok(t, err)
	h.Assert(t, len(flags) == 1, "1 flag should have been processed")
	_, err = cli.ParseAndValidateFlags()
	h.Ok(t, err)

	cli = getTestCLI()
	cli.ByteQuantityFlag(flagName, nil, nil, "Test with validation")
	os.Args = []string{"ec2-instance-selector", flagArg, "(("}
	_, err = cli.ParseAndValidateFlags()
	h.Nok(t, err)
}
