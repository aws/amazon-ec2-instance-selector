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
	"os"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/pkg/cli"
	h "github.com/aws/amazon-ec2-instance-selector/pkg/test"
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
	h.Assert(t, *flagMinOutput == 10 && *flagMaxOutput == 500, "Flag %s max should have been parsed from cmdline and min set to 0", flagArg)
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
	flagName := "test-flag"
	ratioName := flagName + "-ratio"
	configName := flagName + "-config"
	suiteName := flagName + "-suite"
	flagArg := fmt.Sprintf("--%s", flagName)
	ratioArg := fmt.Sprintf("--%s", ratioName)
	configArg := fmt.Sprintf("--%s", configName)
	suiteArg := fmt.Sprintf("--%s", suiteName)

	cli.IntFlag(flagName, nil, nil, "Test Filter Flag")
	cli.RatioFlag(ratioName, nil, nil, "Test Ratio Flag")
	cli.ConfigStringFlag(configName, nil, nil, "Test Config Flag", nil)
	cli.SuiteBoolFlag(suiteName, nil, nil, "Test Suite Flag")

	os.Args = []string{"ec2-instance-selector"}
	flags, err := cli.ParseFlags()
	h.Ok(t, err)
	val, ok := flags[flagName]
	ratioVal, ratioOk := flags[ratioName]
	configVal, configOk := flags[configName]
	suiteVal, suiteOk := flags[suiteName]
	h.Assert(t, ok && ratioOk && configOk && suiteOk, "Flags %s, %s, %s should exist for all types in flags map", flagArg, ratioArg, configArg, suiteArg)
	h.Assert(t, val == nil && ratioVal == nil && configVal == nil && suiteVal == nil,
		"Flag %s, %s, %s should be set to nil when not explicitly set", flagArg, ratioArg, configArg, suiteArg)
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
