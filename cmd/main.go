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
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	commandline "github.com/aws/amazon-ec2-instance-selector/v2/pkg/cli"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/env"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector/outputs"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/sorter"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

const (
	binName             = "ec2-instance-selector"
	awsRegionEnvVar     = "AWS_REGION"
	defaultRegionEnvVar = "AWS_DEFAULT_REGION"
	defaultProfile      = "default"
	awsConfigFile       = "~/.aws/config"
	spotPricingDaysBack = 30

	tableOutput     = "table"
	tableWideOutput = "table-wide"
	oneLine         = "one-line"
	bubbleTeaOutput = "interactive"

	// Sort filter default
	instanceNamePath = ".InstanceType"
)

// Filter Flag Constants
const (
	vcpus                            = "vcpus"
	memory                           = "memory"
	vcpusToMemoryRatio               = "vcpus-to-memory-ratio"
	defaultCores                     = "default-cores"
	defaultThreadsPerCore            = "default-threads-per-core"
	cpuArchitecture                  = "cpu-architecture"
	cpuManufacturer                  = "cpu-manufacturer"
	gpus                             = "gpus"
	gpuMemoryTotal                   = "gpu-memory-total"
	gpuManufacturer                  = "gpu-manufacturer"
	gpuModel                         = "gpu-model"
	inferenceAccelerators            = "inference-accelerators"
	inferenceAcceleratorManufacturer = "inference-accelerator-manufacturer"
	inferenceAcceleratorModel        = "inference-accelerator-model"
	placementGroupStrategy           = "placement-group-strategy"
	usageClass                       = "usage-class"
	rootDeviceType                   = "root-device-type"
	enaSupport                       = "ena-support"
	efaSupport                       = "efa-support"
	hibernationSupport               = "hibernation-support"
	baremetal                        = "baremetal"
	fpgaSupport                      = "fpga-support"
	burstSupport                     = "burst-support"
	hypervisor                       = "hypervisor"
	availabilityZones                = "availability-zones"
	currentGeneration                = "current-generation"
	networkInterfaces                = "network-interfaces"
	networkPerformance               = "network-performance"
	networkEncryption                = "network-encryption"
	ipv6                             = "ipv6"
	allowList                        = "allow-list"
	denyList                         = "deny-list"
	virtualizationType               = "virtualization-type"
	pricePerHour                     = "price-per-hour"
	instanceStorage                  = "instance-storage"
	diskType                         = "disk-type"
	diskEncryption                   = "disk-encryption"
	nvme                             = "nvme"
	ebsOptimized                     = "ebs-optimized"
	ebsOptimizedBaselineBandwidth    = "ebs-optimized-baseline-bandwidth"
	ebsOptimizedBaselineThroughput   = "ebs-optimized-baseline-throughput"
	ebsOptimizedBaselineIOPS         = "ebs-optimized-baseline-iops"
	freeTier                         = "free-tier"
	autoRecovery                     = "auto-recovery"
	dedicatedHosts                   = "dedicated-hosts"
)

// Aggregate Filter Flags
const (
	instanceTypeBase = "base-instance-type"
	flexible         = "flexible"
	service          = "service"
)

// Configuration Flag Constants
const (
	maxResults    = "max-results"
	profile       = "profile"
	help          = "help"
	verbose       = "verbose"
	version       = "version"
	region        = "region"
	output        = "output"
	cacheTTL      = "cache-ttl"
	cacheDir      = "cache-dir"
	sortDirection = "sort-direction"
	sortBy        = "sort-by"
)

var (
	// versionID is overridden at compilation with the version based on the git tag
	versionID = "dev"
)

func main() {

	log.SetOutput(os.Stderr)
	log.SetPrefix("NOTE: ")
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	shortUsage := "A tool to filter EC2 Instance Types based on various resource criteria"
	longUsage := binName + ` is a CLI tool to filter EC2 instance types based on resource criteria. 
Filtering allows you to select all the instance types that match your application requirements.
Full docs can be found at github.com/aws/amazon-` + binName
	examples := fmt.Sprintf(`%s --vcpus 4 --region us-east-2 --availability-zones us-east-2b
%s --memory-min 4 --memory-max 8 --vcpus-min 4 --vcpus-max 8 --region us-east-2`, binName, binName)

	runFunc := func(cmd *cobra.Command, args []string) {}
	cli := commandline.New(binName, shortUsage, longUsage, examples, runFunc)

	cliOutputTypes := []string{
		tableOutput,
		tableWideOutput,
		oneLine,
		bubbleTeaOutput,
	}
	resultsOutputFn := outputs.SimpleInstanceTypeOutput

	cliSortDirections := []string{
		sorter.SortAscending,
		sorter.SortAsc,
		sorter.SortDescending,
		sorter.SortDesc,
	}

	// Registers flags with specific input types from the cli pkg
	// Filter Flags - These will be grouped at the top of the help flags

	cli.Int32MinMaxRangeFlags(vcpus, cli.StringMe("c"), nil, "Number of vcpus available to the instance type.")
	cli.Int32MinMaxRangeFlags(defaultCores, cli.StringMe("p"), nil, "Number of real cores available to the instance type.")
	cli.Int32MinMaxRangeFlags(defaultThreadsPerCore, nil, nil, "Default threads per core (i.e., hyperthreading).")
	cli.ByteQuantityMinMaxRangeFlags(memory, cli.StringMe("m"), nil, "Amount of Memory available (Example: 4 GiB)")
	cli.RatioFlag(vcpusToMemoryRatio, nil, nil, "The ratio of vcpus to GiBs of memory. (Example: 1:2)")
	cli.StringOptionsFlag(cpuArchitecture, cli.StringMe("a"), nil, "CPU architecture [x86_64/amd64, x86_64_mac, i386, or arm64]", []string{"x86_64", "x86_64_mac", "amd64", "i386", "arm64"})
	cli.StringOptionsFlag(cpuManufacturer, nil, nil, "CPU manufacturer [amd, intel, aws]", []string{"amd", "intel", "aws"})
	cli.Int32MinMaxRangeFlags(gpus, cli.StringMe("g"), nil, "Total Number of GPUs (Example: 4)")
	cli.ByteQuantityMinMaxRangeFlags(gpuMemoryTotal, nil, nil, "Number of GPUs' total memory (Example: 4 GiB)")
	cli.StringFlag(gpuManufacturer, nil, nil, "GPU Manufacturer name (Example: NVIDIA)", nil)
	cli.StringFlag(gpuModel, nil, nil, "GPU Model name (Example: K520)", nil)
	cli.IntMinMaxRangeFlags(inferenceAccelerators, nil, nil, "Total Number of inference accelerators (Example: 4)")
	cli.StringFlag(inferenceAcceleratorManufacturer, nil, nil, "Inference Accelerator Manufacturer name (Example: AWS)", nil)
	cli.StringFlag(inferenceAcceleratorModel, nil, nil, "Inference Accelerator Model name (Example: Inferentia)", nil)
	cli.StringOptionsFlag(placementGroupStrategy, nil, nil, "Placement group strategy: [cluster, partition, spread]", []string{"cluster", "partition", "spread"})
	cli.StringOptionsFlag(usageClass, cli.StringMe("u"), nil, "Usage class: [spot or on-demand]", []string{"spot", "on-demand"})
	cli.StringOptionsFlag(rootDeviceType, nil, nil, "Supported root device types: [ebs or instance-store]", []string{"ebs", "instance-store"})
	cli.BoolFlag(enaSupport, cli.StringMe("e"), nil, "Instance types where ENA is supported or required")
	cli.BoolFlag(efaSupport, nil, nil, "Instance types that support Elastic Fabric Adapters (EFA)")
	cli.BoolFlag(hibernationSupport, nil, nil, "Hibernation supported")
	cli.BoolFlag(baremetal, nil, nil, "Bare Metal instance types (.metal instances)")
	cli.BoolFlag(fpgaSupport, cli.StringMe("f"), nil, "FPGA instance types")
	cli.BoolFlag(burstSupport, cli.StringMe("b"), nil, "Burstable instance types")
	cli.StringOptionsFlag(hypervisor, nil, nil, "Hypervisor: [xen or nitro]", []string{"xen", "nitro"})
	cli.StringSliceFlag(availabilityZones, cli.StringMe("z"), nil, "Availability zones or zone ids to check EC2 capacity offered in specific AZs")
	cli.BoolFlag(currentGeneration, nil, nil, "Current generation instance types (explicitly set this to false to not return current generation instance types)")
	cli.Int32MinMaxRangeFlags(networkInterfaces, nil, nil, "Number of network interfaces (ENIs) that can be attached to the instance")
	cli.IntMinMaxRangeFlags(networkPerformance, nil, nil, "Bandwidth in Gib/s of network performance (Example: 100)")
	cli.BoolFlag(networkEncryption, nil, nil, "Instance Types that support automatic network encryption in-transit")
	cli.BoolFlag(ipv6, nil, nil, "Instance Types that support IPv6")
	cli.RegexFlag(allowList, nil, nil, "List of allowed instance types to select from w/ regex syntax (Example: m[3-5]\\.*)")
	cli.RegexFlag(denyList, nil, nil, "List of instance types which should be excluded w/ regex syntax (Example: m[1-2]\\.*)")
	cli.StringOptionsFlag(virtualizationType, nil, nil, "Virtualization Type supported: [hvm or pv]", []string{"hvm", "paravirtual", "pv"})
	cli.Float64MinMaxRangeFlags(pricePerHour, nil, nil, "Price/hour in USD (Example: 0.09)")
	cli.ByteQuantityMinMaxRangeFlags(instanceStorage, nil, nil, "Amount of local instance storage (Example: 4 GiB)")
	cli.StringOptionsFlag(diskType, nil, nil, "Disk Type: [hdd or ssd]", []string{"hdd", "ssd"})
	cli.BoolFlag(nvme, nil, nil, "EBS or local instance storage where NVME is supported or required")
	cli.BoolFlag(diskEncryption, nil, nil, "EBS or local instance storage where encryption is supported or required")
	cli.BoolFlag(ebsOptimized, nil, nil, "EBS Optimized is supported or default")
	cli.ByteQuantityMinMaxRangeFlags(ebsOptimizedBaselineBandwidth, nil, nil, "EBS Optimized baseline bandwidth (Example: 4 GiB)")
	cli.ByteQuantityMinMaxRangeFlags(ebsOptimizedBaselineThroughput, nil, nil, "EBS Optimized baseline throughput per second (Example: 4 GiB)")
	cli.IntMinMaxRangeFlags(ebsOptimizedBaselineIOPS, nil, nil, "EBS Optimized baseline IOPS per second (Example: 10000)")
	cli.BoolFlag(freeTier, nil, nil, "Free Tier supported")
	cli.BoolFlag(autoRecovery, nil, nil, "EC2 Auto-Recovery supported")
	cli.BoolFlag(dedicatedHosts, nil, nil, "Dedicated Hosts supported")

	// Suite Flags - higher level aggregate filters that return opinionated result

	cli.SuiteStringFlag(instanceTypeBase, nil, nil, "Instance Type used to retrieve similarly spec'd instance types", nil)
	cli.SuiteBoolFlag(flexible, nil, nil, "Retrieves a group of instance types spanning multiple generations based on opinionated defaults and user overridden resource filters")
	cli.SuiteStringFlag(service, nil, nil, "Filter instance types based on service support (Example: emr-5.20.0)", nil)

	// Configuration Flags - These will be grouped at the bottom of the help flags

	cli.ConfigIntFlag(maxResults, nil, env.WithDefaultInt("EC2_INSTANCE_SELECTOR_MAX_RESULTS", 20), "The maximum number of instance types that match your criteria to return")
	cli.ConfigStringFlag(profile, nil, nil, "AWS CLI profile to use for credentials and config", nil)
	cli.ConfigStringFlag(region, cli.StringMe("r"), nil, "AWS Region to use for API requests (NOTE: if not passed in, uses AWS SDK default precedence)", nil)
	cli.ConfigStringFlag(output, cli.StringMe("o"), nil, fmt.Sprintf("Specify the output format (%s)", strings.Join(cliOutputTypes, ", ")), nil)
	cli.ConfigIntFlag(cacheTTL, nil, env.WithDefaultInt("EC2_INSTANCE_SELECTOR_CACHE_TTL", 168), "Cache TTLs in hours for pricing and instance type caches. Setting the cache to 0 will turn off caching and cleanup any on-disk caches.")
	cli.ConfigPathFlag(cacheDir, nil, env.WithDefaultString("EC2_INSTANCE_SELECTOR_CACHE_DIR", "~/.ec2-instance-selector/"), "Directory to save the pricing and instance type caches")
	cli.ConfigBoolFlag(verbose, cli.StringMe("v"), nil, "Verbose - will print out full instance specs")
	cli.ConfigBoolFlag(help, cli.StringMe("h"), nil, "Help")
	cli.ConfigBoolFlag(version, nil, nil, "Prints CLI version")
	cli.ConfigStringOptionsFlag(sortDirection, nil, cli.StringMe(sorter.SortAscending), fmt.Sprintf("Specify the direction to sort in (%s)", strings.Join(cliSortDirections, ", ")), cliSortDirections)
	cli.ConfigStringFlag(sortBy, nil, cli.StringMe(instanceNamePath), "Specify the field to sort by. Quantity flags present in this CLI (memory, gpus, etc.) or a JSON path to the appropriate instance type field (Ex: \".MemoryInfo.SizeInMiB\") is acceptable.", nil)

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

	if flags[service] != nil {
		log.Println("--service eks is deprecated. EKS generally supports all instance types")
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(
			aws.ToString(
				cli.StringMe(flags[profile]),
			),
		),
		config.WithRegion(
			aws.ToString(
				cli.StringMe(flags[region]),
			),
		),
	)
	if err != nil {
		fmt.Printf("Failed to load default AWS configuration: %s\n", err.Error())
		os.Exit(1)
	}

	flags[region] = cfg.Region

	cacheTTLDuration := time.Hour * time.Duration(*cli.IntMe(flags[cacheTTL]))
	instanceSelector, err := selector.NewWithCache(ctx, cfg, cacheTTLDuration, *cli.StringMe(flags[cacheDir]))
	if err != nil {
		fmt.Printf("An error occurred when initialising the ec2 selector: %v", err)
		os.Exit(1)
	}
	shutdown := func() {
		if err := instanceSelector.Save(); err != nil {
			log.Printf("There was an error saving pricing caches: %v", err)
		}
	}
	registerShutdown(shutdown)

	sortField := cli.StringMe(flags[sortBy])
	lowercaseSortField := strings.ToLower(*sortField)
	outputFlag := cli.StringMe(flags[output])
	if outputFlag != nil && (*outputFlag == tableWideOutput || *outputFlag == bubbleTeaOutput) {
		// If output type is `table-wide`, simply print both prices for better comparison,
		//   even if the actual filter is applied on any one of those based on usage class
		// Save time by hydrating all caches in parallel
		if err := hydrateCaches(ctx, *instanceSelector); err != nil {
			log.Printf("%v", err)
		}
	} else {
		// Else, if price filters are applied, only hydrate the respective cache as we don't have to print the prices
		if flags[pricePerHour] != nil {
			if flags[usageClass] == nil || *cli.StringMe(flags[usageClass]) == "on-demand" {
				if instanceSelector.EC2Pricing.OnDemandCacheCount() == 0 {
					if err := instanceSelector.EC2Pricing.RefreshOnDemandCache(ctx); err != nil {
						log.Printf("There was a problem refreshing the on-demand pricing cache: %v", err)
					}
				}
			} else {
				if instanceSelector.EC2Pricing.SpotCacheCount() == 0 {
					if err := instanceSelector.EC2Pricing.RefreshSpotCache(ctx, spotPricingDaysBack); err != nil {
						log.Printf("There was a problem refreshing the spot pricing cache: %v", err)
					}
				}
			}
		}

		// refresh appropriate caches if sorting by either spot or on demand pricing
		if strings.Contains(lowercaseSortField, "price") {
			if strings.Contains(lowercaseSortField, "spot") {
				if instanceSelector.EC2Pricing.SpotCacheCount() == 0 {
					if err := instanceSelector.EC2Pricing.RefreshSpotCache(ctx, spotPricingDaysBack); err != nil {
						log.Printf("There was a problem refreshing the spot pricing cache: %v", err)
					}
				}
			} else {
				if instanceSelector.EC2Pricing.OnDemandCacheCount() == 0 {
					if err := instanceSelector.EC2Pricing.RefreshOnDemandCache(ctx); err != nil {
						log.Printf("There was a problem refreshing the on-demand pricing cache: %v", err)
					}
				}
			}
		}
	}

	var cpuArchitectureFilterValue *ec2types.ArchitectureType

	if arch, ok := flags[cpuArchitecture].(*string); ok && arch != nil {
		value := ec2types.ArchitectureType(*arch)
		cpuArchitectureFilterValue = &value
	}

	var cpuManufacturerFilterValue *selector.CPUManufacturer

	if cpuMan, ok := flags[cpuManufacturer].(*string); ok && cpuMan != nil {
		value := selector.CPUManufacturer(*cpuMan)
		cpuManufacturerFilterValue = &value
	}

	var virtualizationTypeFilterValue *ec2types.VirtualizationType

	if virtType, ok := flags[virtualizationType].(*string); ok && virtType != nil {
		value := ec2types.VirtualizationType(*virtType)
		virtualizationTypeFilterValue = &value
	}

	var deviceTypeFilterValue *ec2types.RootDeviceType

	if rootDev, ok := flags[rootDeviceType].(*string); ok && rootDev != nil {
		value := ec2types.RootDeviceType(*rootDev)
		deviceTypeFilterValue = &value
	}

	var usageClassFilterValue *ec2types.UsageClassType

	if useClass, ok := flags[usageClass].(*string); ok && useClass != nil {
		value := ec2types.UsageClassType(*useClass)
		usageClassFilterValue = &value
	}

	var hypervisorFilterValue *ec2types.InstanceTypeHypervisor

	if hype, ok := flags[hypervisor].(*string); ok && hype != nil {
		value := ec2types.InstanceTypeHypervisor(*hype)
		hypervisorFilterValue = &value
	}

	filters := selector.Filters{
		VCpusRange:                       cli.Int32RangeMe(flags[vcpus]),
		DefaultCores:                     cli.Int32RangeMe(flags[defaultCores]),
		DefaultThreadsPerCore:            cli.Int32RangeMe(flags[defaultThreadsPerCore]),
		MemoryRange:                      cli.ByteQuantityRangeMe(flags[memory]),
		VCpusToMemoryRatio:               cli.Float64Me(flags[vcpusToMemoryRatio]),
		CPUArchitecture:                  cpuArchitectureFilterValue,
		CPUManufacturer:                  cpuManufacturerFilterValue,
		GpusRange:                        cli.Int32RangeMe(flags[gpus]),
		GpuMemoryRange:                   cli.ByteQuantityRangeMe(flags[gpuMemoryTotal]),
		GPUManufacturer:                  cli.StringMe(flags[gpuManufacturer]),
		GPUModel:                         cli.StringMe(flags[gpuModel]),
		InferenceAcceleratorsRange:       cli.IntRangeMe(flags[inferenceAccelerators]),
		InferenceAcceleratorManufacturer: cli.StringMe(flags[inferenceAcceleratorManufacturer]),
		InferenceAcceleratorModel:        cli.StringMe(flags[inferenceAcceleratorModel]),
		PlacementGroupStrategy:           cli.StringMe(flags[placementGroupStrategy]),
		UsageClass:                       usageClassFilterValue,
		RootDeviceType:                   deviceTypeFilterValue,
		EnaSupport:                       cli.BoolMe(flags[enaSupport]),
		EfaSupport:                       cli.BoolMe(flags[efaSupport]),
		HibernationSupported:             cli.BoolMe(flags[hibernationSupport]),
		Hypervisor:                       hypervisorFilterValue,
		BareMetal:                        cli.BoolMe(flags[baremetal]),
		Fpga:                             cli.BoolMe(flags[fpgaSupport]),
		Burstable:                        cli.BoolMe(flags[burstSupport]),
		Region:                           cli.StringMe(flags[region]),
		AvailabilityZones:                cli.StringSliceMe(flags[availabilityZones]),
		CurrentGeneration:                cli.BoolMe(flags[currentGeneration]),
		MaxResults:                       cli.IntMe(flags[maxResults]),
		NetworkInterfaces:                cli.Int32RangeMe(flags[networkInterfaces]),
		NetworkPerformance:               cli.IntRangeMe(flags[networkPerformance]),
		NetworkEncryption:                cli.BoolMe(flags[networkEncryption]),
		IPv6:                             cli.BoolMe(flags[ipv6]),
		AllowList:                        cli.RegexMe(flags[allowList]),
		DenyList:                         cli.RegexMe(flags[denyList]),
		InstanceTypeBase:                 cli.StringMe(flags[instanceTypeBase]),
		Flexible:                         cli.BoolMe(flags[flexible]),
		Service:                          cli.StringMe(flags[service]),
		VirtualizationType:               virtualizationTypeFilterValue,
		PricePerHour:                     cli.Float64RangeMe(flags[pricePerHour]),
		InstanceStorageRange:             cli.ByteQuantityRangeMe(flags[instanceStorage]),
		DiskType:                         cli.StringMe(flags[diskType]),
		DiskEncryption:                   cli.BoolMe(flags[diskEncryption]),
		NVME:                             cli.BoolMe(flags[nvme]),
		EBSOptimized:                     cli.BoolMe(flags[ebsOptimized]),
		EBSOptimizedBaselineBandwidth:    cli.ByteQuantityRangeMe(flags[ebsOptimizedBaselineBandwidth]),
		EBSOptimizedBaselineThroughput:   cli.ByteQuantityRangeMe(flags[ebsOptimizedBaselineThroughput]),
		EBSOptimizedBaselineIOPS:         cli.IntRangeMe(flags[ebsOptimizedBaselineIOPS]),
		FreeTier:                         cli.BoolMe(flags[freeTier]),
		AutoRecovery:                     cli.BoolMe(flags[autoRecovery]),
		DedicatedHosts:                   cli.BoolMe(flags[dedicatedHosts]),
	}

	if flags[verbose] != nil {
		resultsOutputFn = outputs.VerboseInstanceTypeOutput
		transformedFilters, err := instanceSelector.AggregateFilterTransform(ctx, filters)
		if err != nil {
			fmt.Printf("An error occurred while transforming the aggregate filters")
			os.Exit(1)
		}
		filtersJSON, err := filters.MarshalIndent("", "    ")
		if err != nil {
			fmt.Printf("An error occurred when printing filters due to --verbose being specified: %v", err)
			os.Exit(1)
		}
		transformedFiltersJSON, err := transformedFilters.MarshalIndent("", "    ")
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

	// fetch instance types without truncating results
	prevMaxResults := filters.MaxResults
	filters.MaxResults = nil
	instanceTypesDetails, err := instanceSelector.FilterVerbose(ctx, filters)
	if err != nil {
		fmt.Printf("An error occurred when filtering instance types: %v", err)
		os.Exit(1)
	}

	// sort instance types
	sortDirection := cli.StringMe(flags[sortDirection])
	instanceTypesDetails, err = sorter.Sort(instanceTypesDetails, *sortField, *sortDirection)
	if err != nil {
		fmt.Printf("Sorting error: %v", err)
		os.Exit(1)
	}

	// handle output format
	var itemsTruncated int
	var instanceTypes []string
	if outputFlag != nil && *outputFlag == bubbleTeaOutput {
		p := tea.NewProgram(outputs.NewBubbleTeaModel(instanceTypesDetails), tea.WithMouseCellMotion())
		if err := p.Start(); err != nil {
			fmt.Printf("An error occurred when starting bubble tea: %v", err)
			os.Exit(1)
		}

		shutdown()
		return
	} else {
		// handle regular output modes

		// truncate instance types based on user passed in maxResults
		instanceTypesDetails, itemsTruncated = truncateResults(prevMaxResults, instanceTypesDetails)
		if len(instanceTypesDetails) == 0 {
			log.Println("The criteria was too narrow and returned no valid instance types. Consider broadening your criteria so that more instance types are returned.")
			os.Exit(1)
		}

		// format instance types for output
		outputFn := getOutputFn(outputFlag, selector.InstanceTypesOutputFn(resultsOutputFn))
		instanceTypes = outputFn(instanceTypesDetails)
	}

	for _, instanceType := range instanceTypes {
		fmt.Println(instanceType)
	}

	if itemsTruncated > 0 {
		log.Printf("%d entries were truncated, increase --%s to see more", itemsTruncated, maxResults)
	}
	shutdown()
}

func hydrateCaches(ctx context.Context, instanceSelector selector.Selector) (errs error) {
	wg := &sync.WaitGroup{}
	hydrateTasks := []func(*sync.WaitGroup) error{
		func(waitGroup *sync.WaitGroup) error {
			defer waitGroup.Done()
			if instanceSelector.EC2Pricing.OnDemandCacheCount() == 0 {
				if err := instanceSelector.EC2Pricing.RefreshOnDemandCache(ctx); err != nil {
					return multierr.Append(errs, fmt.Errorf("There was a problem refreshing the on-demand pricing cache: %w", err))
				}
			}
			return nil
		},
		func(waitGroup *sync.WaitGroup) error {
			defer waitGroup.Done()
			if instanceSelector.EC2Pricing.SpotCacheCount() == 0 {
				if err := instanceSelector.EC2Pricing.RefreshSpotCache(ctx, spotPricingDaysBack); err != nil {
					return multierr.Append(errs, fmt.Errorf("There was a problem refreshing the spot pricing cache: %w", err))
				}
			}
			return nil
		},
		func(waitGroup *sync.WaitGroup) error {
			defer waitGroup.Done()
			if instanceSelector.InstanceTypesProvider.CacheCount() == 0 {
				if _, err := instanceSelector.InstanceTypesProvider.Get(ctx, nil); err != nil {
					return multierr.Append(errs, fmt.Errorf("There was a problem refreshing the instance types cache: %w", err))
				}
			}
			return nil
		},
	}
	wg.Add(len(hydrateTasks))
	for _, task := range hydrateTasks {
		go task(wg)
	}
	wg.Wait()
	return errs
}

func getOutputFn(outputFlag *string, currentFn selector.InstanceTypesOutputFn) selector.InstanceTypesOutputFn {
	outputFn := selector.InstanceTypesOutputFn(currentFn)
	if outputFlag != nil {
		switch *outputFlag {
		case tableWideOutput:
			return selector.InstanceTypesOutputFn(outputs.TableOutputWide)
		case tableOutput:
			return selector.InstanceTypesOutputFn(outputs.TableOutputShort)
		case oneLine:
			return selector.InstanceTypesOutputFn(outputs.OneLineOutput)
		}
	}
	return outputFn
}

func registerShutdown(shutdown func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		shutdown()
	}()
}

func truncateResults(maxResults *int, instanceTypeInfoSlice []*instancetypes.Details) ([]*instancetypes.Details, int) {
	if maxResults == nil {
		return instanceTypeInfoSlice, 0
	}
	upperIndex := *maxResults
	if *maxResults > len(instanceTypeInfoSlice) {
		upperIndex = len(instanceTypeInfoSlice)
	}
	return instanceTypeInfoSlice[0:upperIndex], len(instanceTypeInfoSlice) - upperIndex
}
