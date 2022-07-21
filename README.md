<h1>Amazon EC2 Instance Selector</h1>

<h4>A CLI tool and go library which recommends instance types based on resource criteria like vcpus and memory.</h4>

<p>
  <a href="https://golang.org/doc/go1.17">
    <img src="https://img.shields.io/github/go-mod/go-version/aws/amazon-ec2-instance-selector?color=blueviolet" alt="go-version">
  </a>
  <a href="https://opensource.org/licenses/Apache-2.0">
    <img src="https://img.shields.io/badge/License-Apache%202.0-ff69b4.svg" alt="license">
  </a>
  <a href="https://goreportcard.com/report/github.com/aws/amazon-ec2-instance-selector">
    <img src="https://goreportcard.com/badge/github.com/aws/amazon-ec2-instance-selector" alt="go-report-card">
  </a>
  <a href="https://hub.docker.com/r/amazon/amazon-ec2-instance-selector">
    <img src="https://img.shields.io/docker/pulls/amazon/amazon-ec2-instance-selector" alt="docker-pulls">
  </a>
</p>

![EC2 Instance Selector CI and Release](https://github.com/aws/amazon-ec2-instance-selector/workflows/EC2%20Instance%20Selector%20CI%20and%20Release/badge.svg)

<div>
<hr>
</div>

## Summary

There are over 270 different instance types available on EC2 which can make the process of selecting appropriate instance types difficult. Instance Selector helps you select compatible instance types for your application to run on. The command line interface can be passed resource criteria like vcpus, memory, network performance, and much more and then return the available, matching instance types. 

If you are using spot instances to save on costs, it is a best practice to use multiple instances types within your auto-scaling group (ASG) to ensure your application doesn't experience downtime due to one instance type being interrupted. Instance Selector will help to find a set of instance types that your application can run on.

Instance Selector can also be consumed as a go library for direct integration into your go code.

## Major Features

- Filter AWS Instance Types using declarative resource criteria like vcpus, memory, network performance, and much more!
- Aggregate filters allow for more opinionated instance selections like `--base-instance-type` and `--flexible`
- Consumable as a go library

## Installation and Configuration

#### Install w/ Homebrew

```
brew tap aws/tap
brew install ec2-instance-selector
```

#### Install w/ Curl for Linux/Mac

```
curl -Lo ec2-instance-selector https://github.com/aws/amazon-ec2-instance-selector/releases/download/v2.3.3/ec2-instance-selector-`uname | tr '[:upper:]' '[:lower:]'`-amd64 && chmod +x ec2-instance-selector
```

To execute the CLI, you will need AWS credentials configured. Take a look at the [AWS CLI configuration documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#config-settings-and-precedence) for details on the various ways to configure credentials. An easy way to try out the ec2-instance-selector CLI is to populate the following environment variables with your AWS API credentials.

```
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
```

If you already have an AWS CLI profile setup, you can pass that directly into ec2-instance-selector:

```
$ ec2-instance-selector --profile my-aws-cli-profile --vcpus 2 --region us-east-1
```

You can set the AWS_REGION environment variable if you don't want to pass in `--region` on each run.

```
$ export AWS_REGION="us-east-1"
```

## Examples

### CLI

**Find Instance Types with 4 GiB of memory, 2 vcpus, and runs on the x86_64 CPU architecture**
```
$ ec2-instance-selector --memory 4 --vcpus 2 --cpu-architecture x86_64 -r us-east-1
c5.large
c5a.large
c5ad.large
c5d.large
c6a.large
c6i.large
t2.medium
t3.medium
t3a.medium
```

**Find instance types that support 100GB/s networking that can be purchased as spot instances**
```
$ ec2-instance-selector --network-performance 100 --usage-class spot -r us-east-1
c5n.18xlarge
c5n.metal
c6gn.16xlarge
dl1.24xlarge
g4dn.metal
g5.48xlarge
i3en.24xlarge
i3en.metal
im4gn.16xlarge
inf1.24xlarge
m5dn.24xlarge
m5dn.metal
m5n.24xlarge
m5n.metal
m5zn.12xlarge
m5zn.metal
p3dn.24xlarge
p4d.24xlarge
r5dn.24xlarge
r5dn.metal
```

**Short Table Output**
```
$ ec2-instance-selector --memory 4 --vcpus 2 --cpu-architecture x86_64 -r us-east-1 -o table
Instance Type        VCPUs        Mem (GiB)
-------------        -----        ---------
c5.large             2            4
c5a.large            2            4
c5ad.large           2            4
c5d.large            2            4
c6a.large            2            4
c6i.large            2            4
t2.medium            2            4
t3.medium            2            4
t3a.medium           2            4
```

**Wide Table Output**
```
$ ec2-instance-selector --memory 4 --vcpus 2 --cpu-architecture x86_64 -r us-east-1 -o table-wide
Instance Type  VCPUs   Mem (GiB)  Hypervisor  Current Gen  Hibernation Support  CPU Arch      Network Performance  ENIs    GPUs    GPU Mem (GiB)  GPU Info  On-Demand Price/Hr  Spot Price/Hr (30d avg)
-------------  -----   ---------  ----------  -----------  -------------------  --------      -------------------  ----    ----    -------------  --------  ------------------  -----------------------
c5.large       2       4          nitro       true         true                 x86_64        Up to 10 Gigabit     3       0       0                        $0.085              $0.04708
c5a.large      2       4          nitro       true         false                x86_64        Up to 10 Gigabit     3       0       0                        $0.077              $0.03249
c5ad.large     2       4          nitro       true         false                x86_64        Up to 10 Gigabit     3       0       0                        $0.086              $0.0324
c5d.large      2       4          nitro       true         true                 x86_64        Up to 10 Gigabit     3       0       0                        $0.096              $0.03525
c6a.large      2       4          nitro       true         false                x86_64        Up to 12.5 Gigabit   3       0       0                        $0.0765             $0.034
c6i.large      2       4          nitro       true         false                x86_64        Up to 12.5 Gigabit   3       0       0                        $0.085              $0.03416
t2.medium      2       4          xen         true         true                 i386, x86_64  Low to Moderate      3       0       0                        $0.0464             $0.01407
t3.medium      2       4          nitro       true         true                 x86_64        Up to 5 Gigabit      3       0       0                        $0.0416             $0.0125
t3a.medium     2       4          nitro       true         true                 x86_64        Up to 5 Gigabit      3       0       0                        $0.0376             $0.01431
```

**All CLI Options**

```
$ ec2-instance-selector --help
```

```bash#help
ec2-instance-selector is a CLI tool to filter EC2 instance types based on resource criteria.
Filtering allows you to select all the instance types that match your application requirements.
Full docs can be found at github.com/aws/amazon-ec2-instance-selector

Usage:
  ec2-instance-selector [flags]

Examples:
ec2-instance-selector --vcpus 4 --region us-east-2 --availability-zones us-east-2b
ec2-instance-selector --memory-min 4 --memory-max 8 --vcpus-min 4 --vcpus-max 8 --region us-east-2

Filter Flags:
      --allow-list string                              List of allowed instance types to select from w/ regex syntax (Example: m[3-5]\.*)
      --auto-recovery                                  EC2 Auto-Recovery supported
  -z, --availability-zones strings                     Availability zones or zone ids to check EC2 capacity offered in specific AZs
      --baremetal                                      Bare Metal instance types (.metal instances)
  -b, --burst-support                                  Burstable instance types
  -a, --cpu-architecture string                        CPU architecture [x86_64/amd64, x86_64_mac, i386, or arm64]
      --cpu-manufacturer string                        CPU manufacturer [amd, intel, aws]
      --current-generation                             Current generation instance types (explicitly set this to false to not return current generation instance types)
      --dedicated-hosts                                Dedicated Hosts supported
      --deny-list string                               List of instance types which should be excluded w/ regex syntax (Example: m[1-2]\.*)
      --disk-encryption                                EBS or local instance storage where encryption is supported or required
      --disk-type string                               Disk Type: [hdd or ssd]
      --ebs-optimized                                  EBS Optimized is supported or default
      --ebs-optimized-baseline-bandwidth string        EBS Optimized baseline bandwidth (Example: 4 GiB) (sets --ebs-optimized-baseline-bandwidth-min and -max to the same value)
      --ebs-optimized-baseline-bandwidth-max string    Maximum EBS Optimized baseline bandwidth (Example: 4 GiB) If --ebs-optimized-baseline-bandwidth-min is not specified, the lower bound will be 0
      --ebs-optimized-baseline-bandwidth-min string    Minimum EBS Optimized baseline bandwidth (Example: 4 GiB) If --ebs-optimized-baseline-bandwidth-max is not specified, the upper bound will be infinity
      --ebs-optimized-baseline-iops int                EBS Optimized baseline IOPS per second (Example: 10000) (sets --ebs-optimized-baseline-iops-min and -max to the same value)
      --ebs-optimized-baseline-iops-max int            Maximum EBS Optimized baseline IOPS per second (Example: 10000) If --ebs-optimized-baseline-iops-min is not specified, the lower bound will be 0
      --ebs-optimized-baseline-iops-min int            Minimum EBS Optimized baseline IOPS per second (Example: 10000) If --ebs-optimized-baseline-iops-max is not specified, the upper bound will be infinity
      --ebs-optimized-baseline-throughput string       EBS Optimized baseline throughput per second (Example: 4 GiB) (sets --ebs-optimized-baseline-throughput-min and -max to the same value)
      --ebs-optimized-baseline-throughput-max string   Maximum EBS Optimized baseline throughput per second (Example: 4 GiB) If --ebs-optimized-baseline-throughput-min is not specified, the lower bound will be 0
      --ebs-optimized-baseline-throughput-min string   Minimum EBS Optimized baseline throughput per second (Example: 4 GiB) If --ebs-optimized-baseline-throughput-max is not specified, the upper bound will be infinity
      --efa-support                                    Instance types that support Elastic Fabric Adapters (EFA)
  -e, --ena-support                                    Instance types where ENA is supported or required
  -f, --fpga-support                                   FPGA instance types
      --free-tier                                      Free Tier supported
      --gpu-manufacturer string                        GPU Manufacturer name (Example: NVIDIA)
      --gpu-memory-total string                        Number of GPUs' total memory (Example: 4 GiB) (sets --gpu-memory-total-min and -max to the same value)
      --gpu-memory-total-max string                    Maximum Number of GPUs' total memory (Example: 4 GiB) If --gpu-memory-total-min is not specified, the lower bound will be 0
      --gpu-memory-total-min string                    Minimum Number of GPUs' total memory (Example: 4 GiB) If --gpu-memory-total-max is not specified, the upper bound will be infinity
      --gpu-model string                               GPU Model name (Example: K520)
  -g, --gpus int                                       Total Number of GPUs (Example: 4) (sets --gpus-min and -max to the same value)
      --gpus-max int                                   Maximum Total Number of GPUs (Example: 4) If --gpus-min is not specified, the lower bound will be 0
      --gpus-min int                                   Minimum Total Number of GPUs (Example: 4) If --gpus-max is not specified, the upper bound will be infinity
      --hibernation-support                            Hibernation supported
      --hypervisor string                              Hypervisor: [xen or nitro]
      --inference-accelerator-manufacturer string      Inference Accelerator Manufacturer name (Example: AWS)
      --inference-accelerator-model string             Inference Accelerator Model name (Example: Inferentia)
      --inference-accelerators int                     Total Number of inference accelerators (Example: 4) (sets --inference-accelerators-min and -max to the same value)
      --inference-accelerators-max int                 Maximum Total Number of inference accelerators (Example: 4) If --inference-accelerators-min is not specified, the lower bound will be 0
      --inference-accelerators-min int                 Minimum Total Number of inference accelerators (Example: 4) If --inference-accelerators-max is not specified, the upper bound will be infinity
      --instance-storage string                        Amount of local instance storage (Example: 4 GiB) (sets --instance-storage-min and -max to the same value)
      --instance-storage-max string                    Maximum Amount of local instance storage (Example: 4 GiB) If --instance-storage-min is not specified, the lower bound will be 0
      --instance-storage-min string                    Minimum Amount of local instance storage (Example: 4 GiB) If --instance-storage-max is not specified, the upper bound will be infinity
      --ipv6                                           Instance Types that support IPv6
  -m, --memory string                                  Amount of Memory available (Example: 4 GiB) (sets --memory-min and -max to the same value)
      --memory-max string                              Maximum Amount of Memory available (Example: 4 GiB) If --memory-min is not specified, the lower bound will be 0
      --memory-min string                              Minimum Amount of Memory available (Example: 4 GiB) If --memory-max is not specified, the upper bound will be infinity
      --network-encryption                             Instance Types that support automatic network encryption in-transit
      --network-interfaces int                         Number of network interfaces (ENIs) that can be attached to the instance (sets --network-interfaces-min and -max to the same value)
      --network-interfaces-max int                     Maximum Number of network interfaces (ENIs) that can be attached to the instance If --network-interfaces-min is not specified, the lower bound will be 0
      --network-interfaces-min int                     Minimum Number of network interfaces (ENIs) that can be attached to the instance If --network-interfaces-max is not specified, the upper bound will be infinity
      --network-performance int                        Bandwidth in Gib/s of network performance (Example: 100) (sets --network-performance-min and -max to the same value)
      --network-performance-max int                    Maximum Bandwidth in Gib/s of network performance (Example: 100) If --network-performance-min is not specified, the lower bound will be 0
      --network-performance-min int                    Minimum Bandwidth in Gib/s of network performance (Example: 100) If --network-performance-max is not specified, the upper bound will be infinity
      --nvme                                           EBS or local instance storage where NVME is supported or required
      --placement-group-strategy string                Placement group strategy: [cluster, partition, spread]
      --price-per-hour float                           Price/hour in USD (Example: 0.09) (sets --price-per-hour-min and -max to the same value)
      --price-per-hour-max float                       Maximum Price/hour in USD (Example: 0.09) If --price-per-hour-min is not specified, the lower bound will be 0
      --price-per-hour-min float                       Minimum Price/hour in USD (Example: 0.09) If --price-per-hour-max is not specified, the upper bound will be infinity
      --root-device-type string                        Supported root device types: [ebs or instance-store]
  -u, --usage-class string                             Usage class: [spot or on-demand]
  -c, --vcpus int                                      Number of vcpus available to the instance type. (sets --vcpus-min and -max to the same value)
      --vcpus-max int                                  Maximum Number of vcpus available to the instance type. If --vcpus-min is not specified, the lower bound will be 0
      --vcpus-min int                                  Minimum Number of vcpus available to the instance type. If --vcpus-max is not specified, the upper bound will be infinity
      --vcpus-to-memory-ratio string                   The ratio of vcpus to GiBs of memory. (Example: 1:2)
      --virtualization-type string                     Virtualization Type supported: [hvm or pv]


Suite Flags:
      --base-instance-type string   Instance Type used to retrieve similarly spec'd instance types
      --flexible                    Retrieves a group of instance types spanning multiple generations based on opinionated defaults and user overridden resource filters
      --service string              Filter instance types based on service support (Example: eks, eks-20201211, or emr-5.20.0)


Global Flags:
      --cache-dir string   Directory to save the pricing and instance type caches (default "~/.ec2-instance-selector/")
      --cache-ttl int      Cache TTLs in hours for pricing and instance type caches. Setting the cache to 0 will turn off caching and cleanup any on-disk caches. (default 168)
  -h, --help               Help
      --max-results int    The maximum number of instance types that match your criteria to return (default 20)
  -o, --output string      Specify the output format (table, table-wide, one-line)
      --profile string     AWS CLI profile to use for credentials and config
  -r, --region string      AWS Region to use for API requests (NOTE: if not passed in, uses AWS SDK default precedence)
  -v, --verbose            Verbose - will print out full instance specs
      --version            Prints CLI version
```


### Go Library

This is a minimal example of using the instance selector go package directly:

**cmd/examples/example1.go**
```go#cmd/examples/example1.go
package main

import (
	"fmt"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	// Load an AWS session by looking at shared credentials or environment variables
	// https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	})
	if err != nil {
		fmt.Printf("Oh no, AWS session credentials cannot be found: %v", err)
		return
	}

	// Instantiate a new instance of a selector with the AWS session
	instanceSelector := selector.New(sess)

	// Instantiate an int range filter to specify min and max vcpus
	vcpusRange := selector.IntRangeFilter{
		LowerBound: 2,
		UpperBound: 4,
	}
	// Instantiate a byte quantity range filter to specify min and max memory in GiB
	memoryRange := selector.ByteQuantityRangeFilter{
		LowerBound: bytequantity.FromGiB(2),
		UpperBound: bytequantity.FromGiB(4),
	}
	// Create a string for the CPU Architecture so that it can be passed as a pointer
	// when creating the Filter struct
	cpuArch := "x86_64"

	// Create a Filter struct with criteria you would like to filter
	// The full struct definition can be found here for all of the supported filters:
	// https://github.com/aws/amazon-ec2-instance-selector/blob/main/pkg/selector/types.go
	filters := selector.Filters{
		VCpusRange:      &vcpusRange,
		MemoryRange:     &memoryRange,
		CPUArchitecture: &cpuArch,
	}

	// Pass the Filter struct to the Filter function of your selector instance
	instanceTypesSlice, err := instanceSelector.Filter(filters)
	if err != nil {
		fmt.Printf("Oh no, there was an error :( %v", err)
		return
	}
	// Print the returned instance types slice
	fmt.Println(instanceTypesSlice)
}
```

**Execute the example:**

*NOTE: Make sure you have [AWS credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html#cli-configure-files-settings) setup*
```bash#cmd/examples/example1.go
$ git clone https://github.com/aws/amazon-ec2-instance-selector.git
$ cd amazon-ec2-instance-selector/
$ go run cmd/examples/example1.go
[c4.large c5.large c5a.large c5ad.large c5d.large c6i.large t2.medium t3.medium t3.small t3a.medium t3a.small]
```

## Building
For build instructions please consult [BUILD.md](./BUILD.md).

## Communication
If you've run into a bug or have a new feature request, please open an [issue](https://github.com/aws/amazon-ec2-instance-selector/issues/new).

Check out the open source [Amazon EC2 Spot Instances Integrations Roadmap](https://github.com/aws/ec2-spot-instances-integrations-roadmap) to see what we're working on and give us feedback! 

##  Contributing
Contributions are welcome! Please read our [guidelines](https://github.com/aws/amazon-ec2-instance-selector/blob/main/CONTRIBUTING.md) and our [Code of Conduct](https://github.com/aws/amazon-ec2-instance-selector/blob/main/CODE_OF_CONDUCT.md).

## License
This project is licensed under the [Apache-2.0](LICENSE) License.
