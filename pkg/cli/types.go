// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cli provides functions to build the selector command line interface
package cli

import (
	"log"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector"
)

const (
	// Usage Template to run on --help.
	usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Filter Flags:
{{.LocalNonPersistentFlags.FlagUsages | trimTrailingWhitespaces}}
%s
Global Flags:
{{.PersistentFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}`
)

// validator defines the function for providing validation on a flag.
type validator = func(val interface{}) error

// processor defines the function for providing mutating processing on a flag.
type processor = func(val interface{}) error

// CommandLineInterface is a type to group CLI funcs and state.
type CommandLineInterface struct {
	Command     *cobra.Command
	Flags       map[string]interface{}
	nilDefaults map[string]bool
	rangeFlags  map[string]bool
	validators  map[string]validator
	processors  map[string]processor
	suiteFlags  *pflag.FlagSet
}

// Float64Me takes an interface and returns a pointer to a float64 value
// If the underlying interface kind is not float64 or *float64 then nil is returned.
func (*CommandLineInterface) Float64Me(i interface{}) *float64 {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *float64:
		return v
	case float64:
		return &v
	default:
		log.Printf("%s cannot be converted to a float64", i)
		return nil
	}
}

// IntMe takes an interface and returns a pointer to an int value
// If the underlying interface kind is not int or *int then nil is returned.
func (*CommandLineInterface) IntMe(i interface{}) *int {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *int:
		return v
	case int:
		return &v
	case *int32:
		val := int(*v)
		return &val
	case int32:
		val := int(v)
		return &val
	default:
		log.Printf("%s cannot be converted to an int", i)
		return nil
	}
}

// Int32Me takes an interface and returns a pointer to an int value
// If the underlying interface kind is not int or *int then nil is returned.
func (*CommandLineInterface) Int32Me(i interface{}) *int32 {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *int:
		val := int32(*v)
		return &val
	case int:
		val := int32(v)
		return &val
	case *int32:
		return v
	case int32:
		return &v
	default:
		log.Printf("%s cannot be converted to an int32", i)
		return nil
	}
}

// IntRangeMe takes an interface and returns a pointer to an IntRangeFilter value
// If the underlying interface kind is not IntRangeFilter or *IntRangeFilter then nil is returned.
func (*CommandLineInterface) IntRangeMe(i interface{}) *selector.IntRangeFilter {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *selector.IntRangeFilter:
		return v
	case selector.IntRangeFilter:
		return &v
	default:
		log.Printf("%s cannot be converted to an IntRange", i)
		return nil
	}
}

// Int32RangeMe takes an interface and returns a pointer to an Int32RangeFilter value
// If the underlying interface kind is not Int32RangeFilter or *Int32RangeFilter then nil is returned.
func (*CommandLineInterface) Int32RangeMe(i interface{}) *selector.Int32RangeFilter {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *selector.Int32RangeFilter:
		return v
	case selector.Int32RangeFilter:
		return &v
	default:
		log.Printf("%s cannot be converted to an Int32Range", i)
		return nil
	}
}

// ByteQuantityRangeMe takes an interface and returns a pointer to a ByteQuantityRangeFilter value
// If the underlying interface kind is not ByteQuantityRangeFilter or *ByteQuantityRangeFilter then nil is returned.
func (*CommandLineInterface) ByteQuantityRangeMe(i interface{}) *selector.ByteQuantityRangeFilter {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *selector.ByteQuantityRangeFilter:
		return v
	case selector.ByteQuantityRangeFilter:
		return &v
	default:
		log.Printf("%s cannot be converted to a ByteQuantityRange", i)
		return nil
	}
}

// Float64RangeMe takes an interface and returns a pointer to a Float64RangeFilter value
// If the underlying interface kind is not Float64RangeFilter or *Float64RangeFilter then nil is returned.
func (*CommandLineInterface) Float64RangeMe(i interface{}) *selector.Float64RangeFilter {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *selector.Float64RangeFilter:
		return v
	case selector.Float64RangeFilter:
		return &v
	default:
		log.Printf("%s cannot be converted to a Float64Range", i)
		return nil
	}
}

// StringMe takes an interface and returns a pointer to a string value
// If the underlying interface kind is not string or *string then nil is returned.
func (*CommandLineInterface) StringMe(i interface{}) *string {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *string:
		return v
	case string:
		return &v
	default:
		log.Printf("%s cannot be converted to a string", i)
		return nil
	}
}

// BoolMe takes an interface and returns a pointer to a bool value
// If the underlying interface kind is not bool or *bool then nil is returned.
func (*CommandLineInterface) BoolMe(i interface{}) *bool {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *bool:
		return v
	case bool:
		return &v
	default:
		log.Printf("%s cannot be converted to a bool", i)
		return nil
	}
}

// StringSliceMe takes an interface and returns a pointer to a string slice
// If the underlying interface kind is not []string or *[]string then nil is returned.
func (*CommandLineInterface) StringSliceMe(i interface{}) *[]string {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *[]string:
		return v
	case []string:
		return &v
	default:
		log.Printf("%s cannot be converted to a string list", i)
		return nil
	}
}

// RegexMe takes an interface and returns a pointer to a regex
// If the underlying interface kind is not regexp.Regexp or *regexp.Regexp then nil is returned.
func (*CommandLineInterface) RegexMe(i interface{}) *regexp.Regexp {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *regexp.Regexp:
		return v
	case regexp.Regexp:
		return &v
	default:
		log.Printf("%s cannot be converted to a regexp", i)
		return nil
	}
}

// ByteQuantityMe takes an interface and returns a pointer to a ByteQuantity
// If the underlying interface kind is not bytequantity.ByteQuantity or *bytequantity.ByteQuantity then nil is returned.
func (*CommandLineInterface) ByteQuantityMe(i interface{}) *bytequantity.ByteQuantity {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *bytequantity.ByteQuantity:
		return v
	case bytequantity.ByteQuantity:
		return &v
	default:
		log.Printf("%s cannot be converted to a byte quantity", i)
		return nil
	}
}
