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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	commandline "github.com/aws/amazon-ec2-instance-selector/pkg/cli"
	"github.com/aws/amazon-ec2-instance-selector/pkg/selector"
	"github.com/aws/amazon-ec2-instance-selector/pkg/selector/outputs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
)

const (
	binName = "ec2-instance-selector"
	// cfnJSON is an output type
	cfnJSON = "cfn-json"
	// cfnYAML is an output type
	cfnYAML         = "cfn-yaml"
	terraformHCL    = "terraform-hcl"
	tableOutput     = "table"
	tableWideOutput = "table-wide"
)

// Filter Flag Constants
const (
	vcpus                  = "vcpus"
	memory                 = "memory"
	vcpusToMemoryRatio     = "vcpus-to-memory-ratio"
	cpuArchitecture        = "cpu-architecture"
	gpus                   = "gpus"
	gpuMemoryTotal         = "gpu-memory-total"
	placementGroupStrategy = "placement-group-strategy"
	usageClass             = "usage-class"
	rootDeviceType         = "root-device-type"
	enaSupport             = "ena-support"
	hibernationSupport     = "hibernation-support"
	baremetal              = "baremetal"
	fpgaSupport            = "fpga-support"
	burstSupport           = "burst-support"
	hypervisor             = "hypervisor"
	availabilityZone       = "availability-zone"
	currentGeneration      = "current-generation"
	networkInterfaces      = "network-interfaces"
	networkPerformance     = "network-performance"
)

// Configuration Flag Constants
const (
	maxResults = "max-results"
	profile    = "profile"
	help       = "help"
	verbose    = "verbose"
	version    = "version"
	region     = "region"
	output     = "output"
)

var (
	// versionID is overridden at compilation with the version based on the git tag
	versionID = "dev"
)

func main() {

	log.SetOutput(os.Stderr)

	shortUsage := "A tool to filter EC2 Instance Types based on various resource criteria"
	longUsage := binName + ` is a CLI tool to filter EC2 instance types based on resource criteria. 
Filtering allows you to select all the instance types that match your application requirements.
Full docs can be found at github.com/aws/amazon-` + binName
	examples := fmt.Sprintf(`%s --vcpus 4 --region us-east-2 --availability-zone us-east-2b
%s --memory-min 4096 --memory-max 8192 --vcpus-min 4 --vcpus-max 8 --region us-east-2`, binName, binName)

	runFunc := func(cmd *cobra.Command, args []string) {}
	cli := commandline.New(binName, shortUsage, longUsage, examples, runFunc)

	cliOutputTypes := []string{
		tableOutput,
		tableWideOutput,
	}
	resultsOutputFn := outputs.SimpleInstanceTypeOutput

	// Registers flags with specific input types from the cli pkg
	// Filter Flags - These will be grouped at the top of the help flags

	cli.IntMinMaxRangeFlags(vcpus, cli.StringMe("c"), nil, "Number of vcpus available to the instance type.")
	cli.IntMinMaxRangeFlags(memory, cli.StringMe("m"), nil, "Amount of Memory available in MiB (Example: 4096)")
	cli.RatioFlag(vcpusToMemoryRatio, nil, nil, "The ratio of vcpus to memory in MiB. (Example: 1:2)")
	cli.StringFlag(cpuArchitecture, cli.StringMe("a"), nil, "CPU architecture [x86_64, i386, or arm64]", nil)
	cli.IntMinMaxRangeFlags(gpus, cli.StringMe("g"), nil, "Total Number of GPUs (Example: 4)")
	cli.IntMinMaxRangeFlags(gpuMemoryTotal, nil, nil, "Number of GPUs' total memory in MiB (Example: 4096)")
	cli.StringFlag(placementGroupStrategy, nil, nil, "Placement group strategy: [cluster, partition, spread]", nil)
	cli.StringFlag(usageClass, cli.StringMe("u"), nil, "Usage class: [spot or on-demand]", nil)
	cli.StringFlag(rootDeviceType, nil, nil, "Supported root device types: [ebs or instance-store]", nil)
	cli.BoolFlag(enaSupport, cli.StringMe("e"), nil, "Instance types where ENA is supported or required")
	cli.BoolFlag(hibernationSupport, nil, nil, "Hibernation supported")
	cli.BoolFlag(baremetal, nil, nil, "Bare Metal instance types (.metal instances)")
	cli.BoolFlag(fpgaSupport, cli.StringMe("f"), nil, "FPGA instance types")
	cli.BoolFlag(burstSupport, cli.StringMe("b"), nil, "Burstable instance types")
	cli.StringFlag(hypervisor, nil, nil, "Hypervisor: [xen or nitro]", nil)
	cli.StringFlag(availabilityZone, cli.StringMe("z"), nil, "Availability zone or zone id to check only EC2 capacity offered in a specific AZ", nil)
	cli.BoolFlag(currentGeneration, nil, nil, "Current generation instance types (explicitly set this to false to not return current generation instance types)")
	cli.IntMinMaxRangeFlags(networkInterfaces, nil, nil, "Number of network interfaces (ENIs) that can be attached to the instance")
	cli.IntMinMaxRangeFlags(networkPerformance, nil, nil, "Bandwidth in Gib/s of network performance (Example: 100)")

	// Configuration Flags - These will be grouped at the bottom of the help flags

	cli.ConfigIntFlag(maxResults, nil, cli.IntMe(20), "The maximum number of instance types that match your criteria to return")
	cli.ConfigStringFlag(profile, nil, nil, "AWS CLI profile to use for credentials and config", nil)
	cli.ConfigStringFlag(region, cli.StringMe("r"), nil, "AWS Region to use for API requests (NOTE: if not passed in, uses AWS SDK default precedence)", nil)
	cli.ConfigStringFlag(output, cli.StringMe("o"), nil, fmt.Sprintf("Specify the output format (%s)", strings.Join(cliOutputTypes, ", ")), nil)
	cli.ConfigBoolFlag(verbose, cli.StringMe("v"), nil, "Verbose - will print out full instance specs")
	cli.ConfigBoolFlag(help, cli.StringMe("h"), nil, "Help")
	cli.ConfigBoolFlag(version, nil, nil, "Prints CLI version")

	// Parses the user input with the registered flags and runs type specific validation on the user input
	flags, err := cli.ParseAndValidateFlags()
	if err != nil {
		log.Printf("There was an error while parsing the commandline flags: %v", err)
		os.Exit(1)
	}

	if flags[help] != nil {
		os.Exit(0)
	}

	if flags[version] != nil {
		fmt.Printf("%s", versionID)
		os.Exit(0)
	}

	sessOpts := session.Options{}

	if flags[region] != nil {
		sessOpts.Config.Region = cli.StringMe(flags[region])
	}
	if flags[profile] != nil {
		sessOpts.Profile = *cli.StringMe(flags[profile])
	}

	sess := session.Must(session.NewSessionWithOptions(sessOpts))

	instanceSelector := selector.New(sess)

	filters := selector.Filters{
		VCpusRange:             cli.IntRangeMe(flags[vcpus]),
		MemoryRange:            cli.IntRangeMe(flags[memory]),
		VCpusToMemoryRatio:     cli.Float64Me(flags[vcpusToMemoryRatio]),
		CPUArchitecture:        cli.StringMe(flags[cpuArchitecture]),
		GpusRange:              cli.IntRangeMe(flags[gpus]),
		GpuMemoryRange:         cli.IntRangeMe(flags[gpuMemoryTotal]),
		PlacementGroupStrategy: cli.StringMe(flags[placementGroupStrategy]),
		UsageClass:             cli.StringMe(flags[usageClass]),
		RootDeviceType:         cli.StringMe(flags[rootDeviceType]),
		EnaSupport:             cli.BoolMe(flags[enaSupport]),
		HibernationSupported:   cli.BoolMe(flags[hibernationSupport]),
		Hypervisor:             cli.StringMe(flags[hypervisor]),
		BareMetal:              cli.BoolMe(flags[baremetal]),
		Fpga:                   cli.BoolMe(flags[fpgaSupport]),
		Burstable:              cli.BoolMe(flags[burstSupport]),
		Region:                 cli.StringMe(flags[region]),
		AvailabilityZone:       cli.StringMe(flags[availabilityZone]),
		CurrentGeneration:      cli.BoolMe(flags[currentGeneration]),
		MaxResults:             cli.IntMe(flags[maxResults]),
		NetworkInterfaces:      cli.IntRangeMe(flags[networkInterfaces]),
		NetworkPerformance:     cli.IntRangeMe(flags[networkPerformance]),
	}

	if flags[verbose] != nil {
		resultsOutputFn = outputs.VerboseInstanceTypeOutput
		filtersJSON, err := json.MarshalIndent(filters, "", "    ")
		if err != nil {
			fmt.Printf("An error occurred when printing filters due to --verbose being specified: %v", err)
			os.Exit(1)
		}
		log.Println("\n\n\"Filters\":", string(filtersJSON))
	}

	outputFlag := cli.StringMe(flags[output])
	outputFn := getOutputFn(outputFlag, selector.InstanceTypesOutputFn(resultsOutputFn))

	instanceTypes, err := instanceSelector.FilterWithOutput(filters, outputFn)
	if err != nil {
		fmt.Printf("An error occurred when filtering instance types: %v", err)
		os.Exit(1)
	}
	if len(instanceTypes) == 0 {
		log.Println("The criteria was too narrow and returned no valid instance types. Consider broadening your criteria so that more instance types are returned.")
		os.Exit(1)
	}

	for _, instanceType := range instanceTypes {
		fmt.Println(instanceType)
	}
}

func getOutputFn(outputFlag *string, currentFn selector.InstanceTypesOutputFn) selector.InstanceTypesOutputFn {
	outputFn := selector.InstanceTypesOutputFn(currentFn)
	if outputFlag != nil {
		switch *outputFlag {
		case cfnJSON:
			return selector.InstanceTypesOutputFn(outputs.CloudFormationSpotMixedInstancesPolicyJSONOutput)
		case cfnYAML:
			return selector.InstanceTypesOutputFn(outputs.CloudFormationSpotMixedInstancesPolicyYAMLOutput)
		case terraformHCL:
			return selector.InstanceTypesOutputFn(outputs.TerraformSpotMixedInstancesPolicyHCLOutput)
		case tableWideOutput:
			return selector.InstanceTypesOutputFn(outputs.TableOutputWide)
		case tableOutput:
			return selector.InstanceTypesOutputFn(outputs.TableOutputShort)
		}
	}
	return outputFn
}
