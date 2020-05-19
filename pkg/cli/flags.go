package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

const (
	maxInt = int(^uint(0) >> 1)
)

// RatioFlag creates and registers a flag accepting a Ratio
func (cl *CommandLineInterface) RatioFlag(name string, shorthand *string, defaultValue *string, description string) error {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = cl.StringMe("")
	}
	if shorthand != nil {
		cl.Flags[name] = cl.rootCmd.Flags().StringP(name, string(*shorthand), *defaultValue, description)
		return nil
	}
	cl.Flags[name] = cl.rootCmd.Flags().String(name, *defaultValue, description)
	cl.validators[name] = func(val interface{}) error {
		if val == nil {
			return nil
		}
		vcpuToMemRatioVal := *val.(*string)
		valid, err := regexp.Match(`^[0-9]+:[0-9]+$`, []byte(vcpuToMemRatioVal))
		if err != nil || !valid {
			return fmt.Errorf("Invalid input for --%s. A valid example is 1:2", name)
		}
		vals := strings.Split(vcpuToMemRatioVal, ":")
		vcpusRatioVal, err1 := strconv.Atoi(vals[0])
		memRatioVal, err2 := strconv.Atoi(vals[1])
		if err1 != nil || err2 != nil {
			return fmt.Errorf("Invalid input for --%s. Ratio values must be integers. A valid example is 1:2", name)
		}
		cl.Flags[name] = cl.Float64Me(float64(memRatioVal) / float64(vcpusRatioVal))
		return nil
	}
	return nil
}

// IntMinMaxRangeFlags creates and registers a min, max, and helper flag each accepting an Integer
func (cl *CommandLineInterface) IntMinMaxRangeFlags(name string, shorthand *string, defaultValue *int, description string) {
	cl.IntMinMaxRangeFlagOnFlagSet(cl.rootCmd.Flags(), name, shorthand, defaultValue, description)
}

// IntFlag creates and registers a flag accepting an Integer
func (cl *CommandLineInterface) IntFlag(name string, shorthand *string, defaultValue *int, description string) {
	cl.IntFlagOnFlagSet(cl.rootCmd.Flags(), name, shorthand, defaultValue, description)
}

// StringFlag creates and registers a flag accepting a String and a validator function.
// The validator function is provided so that more complex flags can be created from a string input.
func (cl *CommandLineInterface) StringFlag(name string, shorthand *string, defaultValue *string, description string, validationFn validator) {
	cl.StringFlagOnFlagSet(cl.rootCmd.Flags(), name, shorthand, defaultValue, description, validationFn)
}

// BoolFlag creates and registers a flag accepting a boolean
func (cl *CommandLineInterface) BoolFlag(name string, shorthand *string, defaultValue *bool, description string) {
	cl.BoolFlagOnFlagSet(cl.rootCmd.Flags(), name, shorthand, defaultValue, description)
}

// ConfigStringFlag creates and registers a flag accepting a String for configuration purposes.
// Config flags will be grouped at the bottom in the output of --help
func (cl *CommandLineInterface) ConfigStringFlag(name string, shorthand *string, defaultValue *string, description string, validationFn validator) {
	cl.StringFlagOnFlagSet(cl.rootCmd.PersistentFlags(), name, shorthand, defaultValue, description, validationFn)
}

// ConfigIntFlag creates and registers a flag accepting an Integer for configuration purposes.
// Config flags will be grouped at the bottom in the output of --help
func (cl *CommandLineInterface) ConfigIntFlag(name string, shorthand *string, defaultValue *int, description string) {
	cl.IntFlagOnFlagSet(cl.rootCmd.PersistentFlags(), name, shorthand, defaultValue, description)
}

// ConfigBoolFlag creates and registers a flag accepting a boolean for configuration purposes.
// Config flags will be grouped at the bottom in the output of --help
func (cl *CommandLineInterface) ConfigBoolFlag(name string, shorthand *string, defaultValue *bool, description string) {
	cl.BoolFlagOnFlagSet(cl.rootCmd.PersistentFlags(), name, shorthand, defaultValue, description)
}

// SuiteBoolFlag creates and registers a flag accepting a boolean for configuration purposes.
// Suite flags will be grouped in the middle of the output --help
func (cl *CommandLineInterface) SuiteBoolFlag(name string, shorthand *string, defaultValue *bool, description string) {
	cl.BoolFlagOnFlagSet(cl.suiteFlags, name, shorthand, defaultValue, description)
}

// BoolFlagOnFlagSet creates and registers a flag accepting a boolean for configuration purposes.
func (cl *CommandLineInterface) BoolFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *bool, description string) {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = cl.BoolMe(false)
	}
	if shorthand != nil {
		cl.Flags[name] = flagSet.BoolP(name, string(*shorthand), *defaultValue, description)
		return
	}
	cl.Flags[name] = flagSet.Bool(name, *defaultValue, description)
}

// IntMinMaxRangeFlagOnFlagSet creates and registers a min, max, and helper flag each accepting an Integer
func (cl *CommandLineInterface) IntMinMaxRangeFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *int, description string) {
	cl.IntFlagOnFlagSet(flagSet, name, shorthand, defaultValue, fmt.Sprintf("%s (sets --%s-min and -max to the same value)", description, name))
	cl.IntFlagOnFlagSet(flagSet, name+"-min", nil, nil, fmt.Sprintf("Minimum %s If --%s-max is not specified, the upper bound will be infinity", description, name))
	cl.IntFlagOnFlagSet(flagSet, name+"-max", nil, nil, fmt.Sprintf("Maximum %s If --%s-min is not specified, the lower bound will be 0", description, name))
	cl.validators[name] = func(val interface{}) error {
		if cl.Flags[name+"-min"] == nil || cl.Flags[name+"-max"] == nil {
			return nil
		}
		minArg := name + "-min"
		maxArg := name + "-max"
		minVal := cl.Flags[minArg].(*int)
		maxVal := cl.Flags[maxArg].(*int)
		if *minVal > *maxVal {
			return fmt.Errorf("Invalid input for --%s and --%s. %s must be less than or equal to %s", minArg, maxArg, minArg, maxArg)
		}
		return nil
	}
	cl.intRangeFlags[name] = true
}

// IntFlagOnFlagSet creates and registers a flag accepting an Integer
func (cl *CommandLineInterface) IntFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *int, description string) {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = cl.IntMe(0)
	}
	if shorthand != nil {
		cl.Flags[name] = flagSet.IntP(name, string(*shorthand), *defaultValue, description)
		return
	}
	cl.Flags[name] = flagSet.Int(name, *defaultValue, description)
}

// StringFlagOnFlagSet creates and registers a flag accepting a String and a validator function.
// The validator function is provided so that more complex flags can be created from a string input.
func (cl *CommandLineInterface) StringFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *string, description string, validationFn validator) {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = cl.StringMe("")
	}
	if shorthand != nil {
		cl.Flags[name] = flagSet.StringP(name, string(*shorthand), *defaultValue, description)
		cl.validators[name] = validationFn
		return
	}
	cl.Flags[name] = flagSet.String(name, *defaultValue, description)
	cl.validators[name] = validationFn
}
