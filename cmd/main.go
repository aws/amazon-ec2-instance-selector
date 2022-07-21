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
	"github.com/aws/aws-sdk-go/aws/session"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"gopkg.in/ini.v1"
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
)

// Filter Flag Constants
const (
	vcpus                            = "vcpus"
	memory                           = "memory"
	vcpusToMemoryRatio               = "vcpus-to-memory-ratio"
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

// Sorting Constants
const (
	// Direction

	sortAscending  = "ascending"
	sortAsc        = "asc"
	sortDescending = "descending"
	sortDesc       = "desc"

	// Sorting Fields
	spotPrice = "spot-price"
	odPrice   = "on-demand-price"

	// JSON field paths
	instanceNamePath                   = ".InstanceType"
	vcpuPath                           = ".VCpuInfo.DefaultVCpus"
	memoryPath                         = ".MemoryInfo.SizeInMiB"
	gpuMemoryTotalPath                 = ".GpuInfo.TotalGpuMemoryInMiB"
	networkInterfacesPath              = ".NetworkInfo.MaximumNetworkInterfaces"
	spotPricePath                      = ".SpotPrice"
	odPricePath                        = ".OndemandPricePerHour"
	instanceStoragePath                = ".InstanceStorageInfo.TotalSizeInGB"
	ebsOptimizedBaselineBandwidthPath  = ".EbsInfo.EbsOptimizedInfo.BaselineBandwidthInMbps"
	ebsOptimizedBaselineThroughputPath = ".EbsInfo.EbsOptimizedInfo.BaselineThroughputInMBps"
	ebsOptimizedBaselineIOPSPath       = ".EbsInfo.EbsOptimizedInfo.BaselineIops"
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
	}
	resultsOutputFn := outputs.SimpleInstanceTypeOutput

	cliSortDirections := []string{
		sortAscending,
		sortAsc,
		sortDescending,
		sortDesc,
	}

	sortingShorthandFlags := []string{
		vcpus,
		memory,
		gpuMemoryTotal,
		networkInterfaces,
		spotPrice,
		odPrice,
		instanceStorage,
		ebsOptimizedBaselineBandwidth,
		ebsOptimizedBaselineThroughput,
		ebsOptimizedBaselineIOPS,
		gpus,
		inferenceAccelerators,
	}

	sortingShorthandPaths := []string{
		vcpuPath,
		memoryPath,
		gpuMemoryTotalPath,
		networkInterfacesPath,
		spotPricePath,
		odPricePath,
		instanceStoragePath,
		ebsOptimizedBaselineBandwidthPath,
		ebsOptimizedBaselineThroughputPath,
		ebsOptimizedBaselineIOPS,
		gpus,
		inferenceAccelerators,
	}

	// map quantity cli flags to json paths for easier cli sorting
	sortingKeysMap := mapQuantityFlagsToPath(&sortingShorthandFlags, &sortingShorthandPaths)

	// Registers flags with specific input types from the cli pkg
	// Filter Flags - These will be grouped at the top of the help flags

	cli.IntMinMaxRangeFlags(vcpus, cli.StringMe("c"), nil, "Number of vcpus available to the instance type.")
	cli.ByteQuantityMinMaxRangeFlags(memory, cli.StringMe("m"), nil, "Amount of Memory available (Example: 4 GiB)")
	cli.RatioFlag(vcpusToMemoryRatio, nil, nil, "The ratio of vcpus to GiBs of memory. (Example: 1:2)")
	cli.StringOptionsFlag(cpuArchitecture, cli.StringMe("a"), nil, "CPU architecture [x86_64/amd64, x86_64_mac, i386, or arm64]", []string{"x86_64", "x86_64_mac", "amd64", "i386", "arm64"})
	cli.StringOptionsFlag(cpuManufacturer, nil, nil, "CPU manufacturer [amd, intel, aws]", []string{"amd", "intel", "aws"})
	cli.IntMinMaxRangeFlags(gpus, cli.StringMe("g"), nil, "Total Number of GPUs (Example: 4)")
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
	cli.IntMinMaxRangeFlags(networkInterfaces, nil, nil, "Number of network interfaces (ENIs) that can be attached to the instance")
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
	cli.SuiteStringFlag(service, nil, nil, "Filter instance types based on service support (Example: eks, eks-20201211, or emr-5.20.0)", nil)

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
	cli.ConfigStringOptionsFlag(sortDirection, nil, cli.StringMe(sortAscending), fmt.Sprintf("Specify the direction to sort in (%s)", strings.Join(cliSortDirections, ", ")), cliSortDirections)
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

	sess, err := getRegionAndProfileAWSSession(cli.StringMe(flags[region]), cli.StringMe(flags[profile]))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	flags[region] = sess.Config.Region
	cacheTTLDuration := time.Hour * time.Duration(*cli.IntMe(flags[cacheTTL]))
	instanceSelector := selector.NewWithCache(sess, cacheTTLDuration, *cli.StringMe(flags[cacheDir]))
	shutdown := func() {
		if err := instanceSelector.Save(); err != nil {
			log.Printf("There was an error saving pricing caches: %v", err)
		}
	}
	registerShutdown(shutdown)

	sortField := cli.StringMe(flags[sortBy])
	lowercaseSortField := strings.ToLower(*sortField)
	outputFlag := cli.StringMe(flags[output])
	if outputFlag != nil && *outputFlag == tableWideOutput {
		// If output type is `table-wide`, simply print both prices for better comparison,
		//   even if the actual filter is applied on any one of those based on usage class
		// Save time by hydrating all caches in parallel
		if err := hydrateCaches(*instanceSelector); err != nil {
			log.Printf("%v", err)
		}
	} else {
		// Else, if price filters are applied, only hydrate the respective cache as we don't have to print the prices
		if flags[pricePerHour] != nil {
			if flags[usageClass] == nil || *cli.StringMe(flags[usageClass]) == "on-demand" {
				if instanceSelector.EC2Pricing.OnDemandCacheCount() == 0 {
					if err := instanceSelector.EC2Pricing.RefreshOnDemandCache(); err != nil {
						log.Printf("There was a problem refreshing the on-demand pricing cache: %v", err)
					}
				}
			} else {
				if instanceSelector.EC2Pricing.SpotCacheCount() == 0 {
					if err := instanceSelector.EC2Pricing.RefreshSpotCache(spotPricingDaysBack); err != nil {
						log.Printf("There was a problem refreshing the spot pricing cache: %v", err)
					}
				}
			}
		}

		// refresh appropriate caches if sorting by either spot or on demand pricing
		if strings.Contains(lowercaseSortField, "price") {
			if strings.Contains(lowercaseSortField, "spot") {
				if instanceSelector.EC2Pricing.SpotCacheCount() == 0 {
					if err := instanceSelector.EC2Pricing.RefreshSpotCache(spotPricingDaysBack); err != nil {
						log.Printf("There was a problem refreshing the spot pricing cache: %v", err)
					}
				}
			} else {
				if instanceSelector.EC2Pricing.OnDemandCacheCount() == 0 {
					if err := instanceSelector.EC2Pricing.RefreshOnDemandCache(); err != nil {
						log.Printf("There was a problem refreshing the on-demand pricing cache: %v", err)
					}
				}
			}
		}
	}

	filters := selector.Filters{
		VCpusRange:                       cli.IntRangeMe(flags[vcpus]),
		MemoryRange:                      cli.ByteQuantityRangeMe(flags[memory]),
		VCpusToMemoryRatio:               cli.Float64Me(flags[vcpusToMemoryRatio]),
		CPUArchitecture:                  cli.StringMe(flags[cpuArchitecture]),
		CPUManufacturer:                  cli.StringMe(flags[cpuManufacturer]),
		GpusRange:                        cli.IntRangeMe(flags[gpus]),
		GpuMemoryRange:                   cli.ByteQuantityRangeMe(flags[gpuMemoryTotal]),
		GPUManufacturer:                  cli.StringMe(flags[gpuManufacturer]),
		GPUModel:                         cli.StringMe(flags[gpuModel]),
		InferenceAcceleratorsRange:       cli.IntRangeMe(flags[inferenceAccelerators]),
		InferenceAcceleratorManufacturer: cli.StringMe(flags[inferenceAcceleratorManufacturer]),
		InferenceAcceleratorModel:        cli.StringMe(flags[inferenceAcceleratorModel]),
		PlacementGroupStrategy:           cli.StringMe(flags[placementGroupStrategy]),
		UsageClass:                       cli.StringMe(flags[usageClass]),
		RootDeviceType:                   cli.StringMe(flags[rootDeviceType]),
		EnaSupport:                       cli.BoolMe(flags[enaSupport]),
		EfaSupport:                       cli.BoolMe(flags[efaSupport]),
		HibernationSupported:             cli.BoolMe(flags[hibernationSupport]),
		Hypervisor:                       cli.StringMe(flags[hypervisor]),
		BareMetal:                        cli.BoolMe(flags[baremetal]),
		Fpga:                             cli.BoolMe(flags[fpgaSupport]),
		Burstable:                        cli.BoolMe(flags[burstSupport]),
		Region:                           cli.StringMe(flags[region]),
		AvailabilityZones:                cli.StringSliceMe(flags[availabilityZones]),
		CurrentGeneration:                cli.BoolMe(flags[currentGeneration]),
		MaxResults:                       cli.IntMe(flags[maxResults]),
		NetworkInterfaces:                cli.IntRangeMe(flags[networkInterfaces]),
		NetworkPerformance:               cli.IntRangeMe(flags[networkPerformance]),
		NetworkEncryption:                cli.BoolMe(flags[networkEncryption]),
		IPv6:                             cli.BoolMe(flags[ipv6]),
		AllowList:                        cli.RegexMe(flags[allowList]),
		DenyList:                         cli.RegexMe(flags[denyList]),
		InstanceTypeBase:                 cli.StringMe(flags[instanceTypeBase]),
		Flexible:                         cli.BoolMe(flags[flexible]),
		Service:                          cli.StringMe(flags[service]),
		VirtualizationType:               cli.StringMe(flags[virtualizationType]),
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
		transformedFilters, err := instanceSelector.AggregateFilterTransform(filters)
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

	// determine if user used a shorthand for sorting flag
	sortFieldShorthandPath, ok := sortingKeysMap[*sortField]
	if ok {
		sortField = &sortFieldShorthandPath
	}

	outputFn := getOutputFn(outputFlag, selector.InstanceTypesOutputFn(resultsOutputFn))
	var instanceTypes []string
	var itemsTruncated int

	sortDirection := cli.StringMe(flags[sortDirection])
	if *sortField == instanceNamePath && (*sortDirection == sortAscending || *sortDirection == sortAsc) {
		// filter already sorts in ascending order by name
		instanceTypes, itemsTruncated, err = instanceSelector.FilterWithOutput(filters, outputFn)
		if err != nil {
			fmt.Printf("An error occurred when filtering instance types: %v", err)
			os.Exit(1)
		}
		if len(instanceTypes) == 0 {
			log.Println("The criteria was too narrow and returned no valid instance types. Consider broadening your criteria so that more instance types are returned.")
			os.Exit(1)
		}
	} else {
		// fetch instance types without truncating results
		prevMaxResults := filters.MaxResults
		filters.MaxResults = nil
		instanceTypeDetails, err := instanceSelector.FilterVerbose(filters)
		if err != nil {
			fmt.Printf("An error occurred when filtering instance types: %v", err)
			os.Exit(1)
		}

		// sort instance types
		sorter, err := sorter.NewSorter(instanceTypeDetails, *sortField, *sortDirection)
		if err != nil {
			fmt.Printf("An error occurred when preparing to sort instance types: %v", err)
			os.Exit(1)
		}
		err = sorter.Sort()
		if err != nil {
			fmt.Printf("An error occurred when sorting instance types: %v", err)
			os.Exit(1)
		}
		instanceTypeDetails = sorter.InstanceTypes()

		// truncate instance types based on user passed in maxResults
		instanceTypeDetails, itemsTruncated = truncateResults(prevMaxResults, instanceTypeDetails)
		if len(instanceTypeDetails) == 0 {
			log.Println("The criteria was too narrow and returned no valid instance types. Consider broadening your criteria so that more instance types are returned.")
			os.Exit(1)
		}

		// format instance types for output
		instanceTypes = outputFn(instanceTypeDetails)
	}

	for _, instanceType := range instanceTypes {
		fmt.Println(instanceType)
	}

	if itemsTruncated > 0 {
		log.Printf("%d entries were truncated, increase --%s to see more", itemsTruncated, maxResults)
	}
	shutdown()
}

func hydrateCaches(instanceSelector selector.Selector) (errs error) {
	wg := &sync.WaitGroup{}
	hydrateTasks := []func(*sync.WaitGroup) error{
		func(waitGroup *sync.WaitGroup) error {
			defer waitGroup.Done()
			if instanceSelector.EC2Pricing.OnDemandCacheCount() == 0 {
				if err := instanceSelector.EC2Pricing.RefreshOnDemandCache(); err != nil {
					return multierr.Append(errs, fmt.Errorf("There was a problem refreshing the on-demand pricing cache: %w", err))
				}
			}
			return nil
		},
		func(waitGroup *sync.WaitGroup) error {
			defer waitGroup.Done()
			if instanceSelector.EC2Pricing.SpotCacheCount() == 0 {
				if err := instanceSelector.EC2Pricing.RefreshSpotCache(spotPricingDaysBack); err != nil {
					return multierr.Append(errs, fmt.Errorf("There was a problem refreshing the spot pricing cache: %w", err))
				}
			}
			return nil
		},
		func(waitGroup *sync.WaitGroup) error {
			defer waitGroup.Done()
			if instanceSelector.InstanceTypesProvider.CacheCount() == 0 {
				if _, err := instanceSelector.InstanceTypesProvider.Get(nil); err != nil {
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

func mapQuantityFlagsToPath(flags *[]string, paths *[]string) map[string]string {
	sortingFlagKeys := make(map[string]string)
	for i := range *flags {
		sortingFlagKeys[(*flags)[i]] = (*paths)[i]
	}

	return sortingFlagKeys
}
