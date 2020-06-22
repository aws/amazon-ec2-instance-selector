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
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

const (
	binName             = "ec2-instance-selector"
	awsRegionEnvVar     = "AWS_REGION"
	defaultRegionEnvVar = "AWS_DEFAULT_REGION"
	defaultProfile      = "default"
	awsConfigFile       = "~/.aws/config"

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
	availabilityZones      = "availability-zones"
	currentGeneration      = "current-generation"
	networkInterfaces      = "network-interfaces"
	networkPerformance     = "network-performance"
	allowList              = "allow-list"
	denyList               = "deny-list"
)

// Aggregate Filter Flags
const (
	instanceTypeBase = "base-instance-type"
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
	examples := fmt.Sprintf(`%s --vcpus 4 --region us-east-2 --availability-zones us-east-2b
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
	cli.StringFlag(availabilityZone, nil, nil, "[DEPRECATED] use --availability-zones instead", nil)
	cli.StringSliceFlag(availabilityZones, cli.StringMe("z"), nil, "Availability zones or zone ids to check EC2 capacity offered in specific AZs")
	cli.BoolFlag(currentGeneration, nil, nil, "Current generation instance types (explicitly set this to false to not return current generation instance types)")
	cli.IntMinMaxRangeFlags(networkInterfaces, nil, nil, "Number of network interfaces (ENIs) that can be attached to the instance")
	cli.IntMinMaxRangeFlags(networkPerformance, nil, nil, "Bandwidth in Gib/s of network performance (Example: 100)")
	cli.RegexFlag(allowList, nil, nil, "List of allowed instance types to select from w/ regex syntax (Example: m[3-5]\\.*)")
	cli.RegexFlag(denyList, nil, nil, "List of instance types which should be excluded w/ regex syntax (Example: m[1-2]\\.*)")

	// Suite Flags - higher level aggregate filters that return opinionated result

	cli.SuiteStringFlag(instanceTypeBase, nil, nil, "Instance Type used to retrieve similarly spec'd instance types", nil)

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

	sess, err := getRegionAndProfileAWSSession(cli.StringMe(flags[region]), cli.StringMe(flags[profile]))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	flags[region] = sess.Config.Region

	if flags[availabilityZone] != nil {
		log.Printf("You are using a deprecated flag --%s which will be removed in future versions, switch to --%s to avoid issues.\n", availabilityZone, availabilityZones)
		if flags[availabilityZones] != nil {
			flags[availabilityZones] = append(*cli.StringSliceMe(flags[availabilityZones]), *cli.StringMe(flags[availabilityZone]))
		} else {
			flags[availabilityZones] = []string{*cli.StringMe(flags[availabilityZone])}
		}
	}

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
		AvailabilityZones:      cli.StringSliceMe(flags[availabilityZones]),
		CurrentGeneration:      cli.BoolMe(flags[currentGeneration]),
		MaxResults:             cli.IntMe(flags[maxResults]),
		NetworkInterfaces:      cli.IntRangeMe(flags[networkInterfaces]),
		NetworkPerformance:     cli.IntRangeMe(flags[networkPerformance]),
		AllowList:              cli.RegexMe(flags[allowList]),
		DenyList:               cli.RegexMe(flags[denyList]),
		InstanceTypeBase:       cli.StringMe(flags[instanceTypeBase]),
	}

	if flags[verbose] != nil {
		resultsOutputFn = outputs.VerboseInstanceTypeOutput
		transformedFilters, err := instanceSelector.AggregateFilterTransform(filters, selector.AggregateLowPercentile, selector.AggregateHighPercentile)
		if err != nil {
			fmt.Printf("An error occurred while transforming the aggregate filters")
			os.Exit(1)
		}
		filtersJSON, err := json.MarshalIndent(filters, "", "    ")
		if err != nil {
			fmt.Printf("An error occurred when printing filters due to --verbose being specified: %v", err)
			os.Exit(1)
		}
		transformedFiltersJSON, err := json.MarshalIndent(transformedFilters, "", "    ")
		if err != nil {
			fmt.Printf("An error occurred when printing aggregate filters due to --verbose being specified: %v", err)
			os.Exit(1)
		}
		log.Println("\n\n\"Filters\":", string(filtersJSON))
		if string(transformedFiltersJSON) != string(filtersJSON) {
			log.Println("\n\n\"Transformed Filters\":", string(transformedFiltersJSON))
		} else {
			log.Println("There were no transformations on the filters to display")
		}
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

func getRegionAndProfileAWSSession(regionName *string, profileName *string) (*session.Session, error) {
	sessOpts := session.Options{SharedConfigState: session.SharedConfigEnable}
	if regionName != nil {
		sessOpts.Config.Region = regionName
	}

	if profileName != nil {
		sessOpts.Profile = *profileName
		if sessOpts.Config.Region == nil {
			if profileRegion, err := getProfileRegion(*profileName); err != nil {
				log.Println(err)
			} else {
				sessOpts.Config.Region = &profileRegion
			}
		}
	}

	sess := session.Must(session.NewSessionWithOptions(sessOpts))
	if sess.Config.Region != nil && *sess.Config.Region != "" {
		return sess, nil
	}
	if defaultProfileRegion, err := getProfileRegion(defaultProfile); err == nil {
		sess.Config.Region = &defaultProfileRegion
		return sess, nil
	}

	if defaultRegion, ok := os.LookupEnv(defaultRegionEnvVar); ok && defaultRegion != "" {
		sess.Config.Region = &defaultRegion
		return sess, nil
	}

	errorMsg := "Unable to find a region in the usual places: \n"
	errorMsg = errorMsg + "\t - --region flag\n"
	errorMsg = errorMsg + fmt.Sprintf("\t - %s environment variable\n", awsRegionEnvVar)
	if profileName != nil {
		errorMsg = errorMsg + fmt.Sprintf("\t - profile region in %s\n", awsConfigFile)
	}
	errorMsg = errorMsg + fmt.Sprintf("\t - default profile region in %s\n", awsConfigFile)
	errorMsg = errorMsg + fmt.Sprintf("\t - %s environment variable\n", defaultRegionEnvVar)
	return sess, fmt.Errorf(errorMsg)
}

func getProfileRegion(profileName string) (string, error) {
	if profileName != defaultProfile {
		profileName = fmt.Sprintf("profile %s", profileName)
	}
	awsConfigPath, err := homedir.Expand(awsConfigFile)
	if err != nil {
		return "", fmt.Errorf("Warning: unable to find home directory to parse aws config file")
	}
	awsConfigIni, err := ini.Load(awsConfigPath)
	if err != nil {
		return "", fmt.Errorf("Warning: unable to load aws config file for profile at path: %s", awsConfigPath)
	}
	section, err := awsConfigIni.GetSection(profileName)
	if err != nil {
		return "", fmt.Errorf("Warning: there is no configuration for the specified aws profile %s at %s", profileName, awsConfigPath)
	}
	regionConfig, err := section.GetKey("region")
	if err != nil || regionConfig.String() == "" {
		return "", fmt.Errorf("Warning: there is no region configured for the specified aws profile %s at %s", profileName, awsConfigPath)
	}
	return regionConfig.String(), nil
}
