<h1>Amazon EC2 Instance Selector</h1>

<h4>A CLI tool and go library which recommends instance types based on resource criteria like vcpus and memory.</h4>

<p>
  <a href="https://golang.org/doc/go1.23">
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

There are over 800 different instance types available on EC2 which can make the process of selecting appropriate instance types difficult. Instance Selector helps you select compatible instance types for your application to run on. The command line interface can be passed resource criteria like vcpus, memory, network performance, and much more and then return the available, matching instance types. 

If you are using spot instances to save on costs, it is a best practice to use multiple instances types within your auto-scaling group (ASG) to ensure your application doesn't experience downtime due to one instance type being interrupted. Instance Selector will help to find a set of instance types that your application can run on.

Instance Selector can also be consumed as a go library for direct integration into your go code.

## Major Features

- Filter AWS Instance Types using declarative resource criteria like vcpus, memory, network performance, and much more!
- Aggregate filters allow for more opinionated instance selections like `--base-instance-type` and `--flexible`
- Consumable as a go library or CLI
- Interactive TUI w/ `--output interactive`

## Installation and Configuration

#### Install w/ Homebrew

```
brew tap aws/tap
brew install ec2-instance-selector
```

#### Install w/ Curl for Linux/Mac

```
curl -Lo ec2-instance-selector https://github.com/aws/amazon-ec2-instance-selector/releases/download/v2.4.1/ec2-instance-selector-`uname | tr '[:upper:]' '[:lower:]'`-amd64 && chmod +x ec2-instance-selector
sudo mv ec2-instance-selector /usr/local/bin/
ec2-instance-selector --version
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
c6id.large
c6in.large
c7a.large
c7i-flex.large
c7i.large
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
c6in.16xlarge
c7gn.8xlarge
dl1.24xlarge
f2.48xlarge
g4dn.metal
g5.48xlarge
g6.48xlarge
g6e.12xlarge
i3en.24xlarge
i3en.metal
i7ie.24xlarge
i7ie.48xlarge
im4gn.16xlarge
inf1.24xlarge
inf2.48xlarge
m5dn.24xlarge
m5dn.metal
NOTE: 22 entries were truncated, increase --max-results to see more
```

**Short Table Output**
```
$ ec2-instance-selector --memory 4 --vcpus 2 --cpu-architecture x86_64 -r us-east-1 -o table
Instance Type         VCPUs        Mem (GiB)
-------------         -----        ---------
c5.large              2            4
c5a.large             2            4
c5ad.large            2            4
c5d.large             2            4
c6a.large             2            4
c6i.large             2            4
c6id.large            2            4
c6in.large            2            4
c7a.large             2            4
c7i-flex.large        2            4
c7i.large             2            4
t2.medium             2            4
t3.medium             2            4
t3a.medium            2            4
```

**Wide Table Output**
```
$ ec2-instance-selector --memory 4 --vcpus 2 --cpu-architecture x86_64 -r us-east-1 -o table-wide
Instance Type   VCPUs   Mem (GiB)  Hypervisor  Current Gen  Hibernation Support  CPU Arch      Network Performance  ENIs    GPUs    GPU Mem (GiB)  GPU Info  On-Demand Price/Hr  Spot Price/Hr
-------------   -----   ---------  ----------  -----------  -------------------  --------      -------------------  ----    ----    -------------  --------  ------------------  -------------
c5.large        2       4          nitro       true         true                 x86_64        Up to 10 Gigabit     3       0       0              none      $0.085              $0.0264
c5a.large       2       4          nitro       true         false                x86_64        Up to 10 Gigabit     3       0       0              none      $0.077              $0.025
c5ad.large      2       4          nitro       true         false                x86_64        Up to 10 Gigabit     3       0       0              none      $0.086              $0.0358
c5d.large       2       4          nitro       true         true                 x86_64        Up to 10 Gigabit     3       0       0              none      $0.096              $0.036
c6a.large       2       4          nitro       true         true                 x86_64        Up to 12.5 Gigabit   3       0       0              none      $0.0765             $0.0252
c6i.large       2       4          nitro       true         true                 x86_64        Up to 12.5 Gigabit   3       0       0              none      $0.085              $0.0252
c6id.large      2       4          nitro       true         true                 x86_64        Up to 12.5 Gigabit   3       0       0              none      $0.1008             $0.0359
c6in.large      2       4          nitro       true         true                 x86_64        Up to 25 Gigabit     3       0       0              none      $0.1134             $0.043
c7a.large       2       4          nitro       true         true                 x86_64        Up to 12.5 Gigabit   3       0       0              none      $0.10264            $0.0429
c7i-flex.large  2       4          nitro       true         true                 x86_64        Up to 12.5 Gigabit   3       0       0              none      $0.08479            $0.0245
c7i.large       2       4          nitro       true         true                 x86_64        Up to 12.5 Gigabit   3       0       0              none      $0.08925            $0.0367
t2.medium       2       4          xen         true         true                 i386, x86_64  Low to Moderate      3       0       0              none      $0.0464             $0.0135
t3.medium       2       4          nitro       true         true                 x86_64        Up to 5 Gigabit      3       0       0              none      $0.0416             $0.016
t3a.medium      2       4          nitro       true         true                 x86_64        Up to 5 Gigabit      3       0       0              none      $0.0376             $0.0108
```

**Interactive Output**
```
$ ec2-instance-selector -o interactive
```
https://user-images.githubusercontent.com/68402662/184218343-6b236d4a-3fe6-42ae-9fe3-3fd3ee92a4b5.mov

**Sort by memory in ascending order using shorthand**
```
$ ec2-instance-selector -r us-east-1 -o table-wide --max-results 10 --sort-by memory --sort-direction asc
Instance Type  VCPUs   Mem (GiB)  Hypervisor  Current Gen  Hibernation Support  CPU Arch      Network Performance  ENIs    GPUs    GPU Mem (GiB)  GPU Info  On-Demand Price/Hr  Spot Price/Hr
-------------  -----   ---------  ----------  -----------  -------------------  --------      -------------------  ----    ----    -------------  --------  ------------------  -------------
t2.nano        1       0.5        xen         true         true                 i386, x86_64  Low to Moderate      2       0       0              none      $0.0058             -Not Fetched-
t4g.nano       2       0.5        nitro       true         true                 arm64         Up to 5 Gigabit      2       0       0              none      $0.0042             $0.0011
t3.nano        2       0.5        nitro       true         true                 x86_64        Up to 5 Gigabit      2       0       0              none      $0.0052             $0.0011
t3a.nano       2       0.5        nitro       true         true                 x86_64        Up to 5 Gigabit      2       0       0              none      $0.0047             $0.0022
t1.micro       1       0.6123     xen         false        false                i386, x86_64  Very Low             2       0       0              none      $0.02               $0.0024
t2.micro       1       1          xen         true         true                 i386, x86_64  Low to Moderate      2       0       0              none      $0.0116             $0.0029
t3a.micro      2       1          nitro       true         true                 x86_64        Up to 5 Gigabit      2       0       0              none      $0.0094             $0.0033
t4g.micro      2       1          nitro       true         true                 arm64         Up to 5 Gigabit      2       0       0              none      $0.0084             $0.0046
t3.micro       2       1          nitro       true         true                 x86_64        Up to 5 Gigabit      2       0       0              none      $0.0104             $0.0028
m1.small       1       1.69922    xen         false        false                i386, x86_64  Low                  2       0       0              none      $0.044              $0.0115
NOTE: 855 entries were truncated, increase --max-results to see more
```
Available shorthand flags: vcpus, memory, gpu-memory-total, network-interfaces, spot-price, on-demand-price, instance-storage, ebs-optimized-baseline-bandwidth, ebs-optimized-baseline-throughput, ebs-optimized-baseline-iops, gpus, inference-accelerators

**Sort by memory in descending order using JSON path**
```
$ ec2-instance-selector -r us-east-1 -o table-wide --max-results 10 --sort-by .MemoryInfo.SizeInMiB --sort-direction desc
Instance Type        VCPUs   Mem (GiB)  Hypervisor  Current Gen  Hibernation Support  CPU Arch  Network Performance  ENIs    GPUs    GPU Mem (GiB)  GPU Info  On-Demand Price/Hr  Spot Price/Hr
-------------        -----   ---------  ----------  -----------  -------------------  --------  -------------------  ----    ----    -------------  --------  ------------------  -------------
u7in-32tb.224xlarge  896     32,768     nitro       true         false                x86_64    200 Gigabit          16      0       0              none      $407.68             -Not Fetched-
u7in-24tb.224xlarge  896     24,576     nitro       true         false                x86_64    200 Gigabit          16      0       0              none      $305.76             -Not Fetched-
u-24tb1.112xlarge    448     24,576     nitro       true         false                x86_64    100 Gigabit          15      0       0              none      $218.4              -Not Fetched-
u-18tb1.112xlarge    448     18,432     nitro       true         false                x86_64    100 Gigabit          15      0       0              none      $163.8              -Not Fetched-
u7in-16tb.224xlarge  896     16,384     nitro       true         false                x86_64    200 Gigabit          16      0       0              none      $203.84             -Not Fetched-
u7i-12tb.224xlarge   896     12,288     nitro       true         false                x86_64    100 Gigabit          15      0       0              none      $152.88             -Not Fetched-
u-12tb1.112xlarge    448     12,288     nitro       true         false                x86_64    100 Gigabit          15      0       0              none      $109.2              -Not Fetched-
u-9tb1.112xlarge     448     9,216      nitro       true         false                x86_64    100 Gigabit          15      0       0              none      $81.9               -Not Fetched-
u7i-8tb.112xlarge    448     8,192      nitro       true         false                x86_64    100 Gigabit          15      0       0              none      $83.72              -Not Fetched-
u-6tb1.56xlarge      224     6,144      nitro       true         false                x86_64    100 Gigabit          15      0       0              none      $46.40391           -Not Fetched-
NOTE: 855 entries were truncated, increase --max-results to see more
```
JSON path must point to a field in the [instancetype.Details struct](https://github.com/aws/amazon-ec2-instance-selector/blob/5bffbf2750ee09f5f1308bdc8d4b635a2c6e2721/pkg/instancetypes/instancetypes.go#L37).

**Example output of instance type object using Verbose output**
```
$ ec2-instance-selector --max-results 1 -v
NOTE:

"Filters": {
    "AllowList": null,
    "DenyList": null,
    "AvailabilityZones": [],
    "BareMetal": null,
    "Burstable": null,
    "AutoRecovery": null,
    "FreeTier": null,
    "CPUArchitecture": null,
    "CPUManufacturer": null,
    "CurrentGeneration": null,
    "EnaSupport": null,
    "EfaSupport": null,
    "Fpga": null,
    "GpusRange": null,
    "GpuMemoryRange": null,
    "GPUManufacturer": null,
    "GPUModel": null,
    "InferenceAcceleratorsRange": null,
    "InferenceAcceleratorManufacturer": null,
    "InferenceAcceleratorModel": null,
    "HibernationSupported": null,
    "Hypervisor": null,
    "MaxResults": 1,
    "MemoryRange": null,
    "NetworkInterfaces": null,
    "NetworkPerformance": null,
    "NetworkEncryption": null,
    "IPv6": null,
    "PlacementGroupStrategy": null,
    "Region": "us-east-1",
    "RootDeviceType": null,
    "UsageClass": null,
    "VCpusRange": null,
    "VCpusToMemoryRatio": null,
    "InstanceTypeBase": null,
    "Flexible": null,
    "Service": null,
    "InstanceTypes": null,
    "VirtualizationType": null,
    "PricePerHour": null,
    "InstanceStorageRange": null,
    "DiskType": null,
    "NVME": null,
    "EBSOptimized": null,
    "DiskEncryption": null,
    "EBSOptimizedBaselineBandwidth": null,
    "EBSOptimizedBaselineThroughput": null,
    "EBSOptimizedBaselineIOPS": null,
    "DedicatedHosts": null,
    "Generation": null
}
NOTE: There were no transformations on the filters to display
[
    {
        "AutoRecoverySupported": true,
        "BareMetal": false,
        "BurstablePerformanceSupported": false,
        "CurrentGeneration": false,
        "DedicatedHostsSupported": true,
        "EbsInfo": {
            "EbsOptimizedInfo": {
                "BaselineBandwidthInMbps": 1750,
                "BaselineIops": 10000,
                "BaselineThroughputInMBps": 218.75,
                "MaximumBandwidthInMbps": 3500,
                "MaximumIops": 20000,
                "MaximumThroughputInMBps": 437.5
            },
            "EbsOptimizedSupport": "default",
            "EncryptionSupport": "supported",
            "NvmeSupport": "required"
        },
        "FpgaInfo": null,
        "FreeTierEligible": false,
        "GpuInfo": null,
        "HibernationSupported": false,
        "Hypervisor": "nitro",
        "InferenceAcceleratorInfo": null,
        "InstanceStorageInfo": null,
        "InstanceStorageSupported": false,
        "InstanceType": "a1.2xlarge",
        "MediaAcceleratorInfo": null,
        "MemoryInfo": {
            "SizeInMiB": 16384
        },
        "NetworkInfo": {
            "DefaultNetworkCardIndex": 0,
            "EfaInfo": null,
            "EfaSupported": false,
            "EnaSrdSupported": false,
            "EnaSupport": "required",
            "EncryptionInTransitSupported": false,
            "Ipv4AddressesPerInterface": 15,
            "Ipv6AddressesPerInterface": 15,
            "Ipv6Supported": true,
            "MaximumNetworkCards": 1,
            "MaximumNetworkInterfaces": 4,
            "NetworkCards": [
                {
                    "BaselineBandwidthInGbps": 2.5,
                    "MaximumNetworkInterfaces": 4,
                    "NetworkCardIndex": 0,
                    "NetworkPerformance": "Up to 10 Gigabit",
                    "PeakBandwidthInGbps": 10
                }
            ],
            "NetworkPerformance": "Up to 10 Gigabit"
        },
        "NeuronInfo": null,
        "NitroEnclavesSupport": "unsupported",
        "NitroTpmInfo": null,
        "NitroTpmSupport": "unsupported",
        "PhcSupport": "unsupported",
        "PlacementGroupInfo": {
            "SupportedStrategies": [
                "cluster",
                "partition",
                "spread"
            ]
        },
        "ProcessorInfo": {
            "Manufacturer": "AWS",
            "SupportedArchitectures": [
                "arm64"
            ],
            "SupportedFeatures": null,
            "SustainedClockSpeedInGhz": 2.3
        },
        "SupportedBootModes": [
            "uefi"
        ],
        "SupportedRootDeviceTypes": [
            "ebs"
        ],
        "SupportedUsageClasses": [
            "on-demand",
            "spot"
        ],
        "SupportedVirtualizationTypes": [
            "hvm"
        ],
        "VCpuInfo": {
            "DefaultCores": 8,
            "DefaultThreadsPerCore": 1,
            "DefaultVCpus": 8,
            "ValidCores": null,
            "ValidThreadsPerCore": null
        },
        "OndemandPricePerHour": null,
        "SpotPrice": null
    }
]
NOTE: 864 entries were truncated, increase --max-results to see more
```
NOTE: Use this JSON format as reference when finding JSON paths for sorting

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
  -a, --cpu-architecture string                        CPU architecture [x86_64, amd64, x86_64_mac, i386, arm64, or arm64_mac]
      --cpu-manufacturer string                        CPU manufacturer [amd, intel, aws, apple]
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
      --generation int                                 Generation of the instance type (i.e. c7i.xlarge is 7) (sets --generation-min and -max to the same value)
      --generation-max int                             Maximum Generation of the instance type (i.e. c7i.xlarge is 7) If --generation-min is not specified, the lower bound will be 0
      --generation-min int                             Minimum Generation of the instance type (i.e. c7i.xlarge is 7) If --generation-max is not specified, the upper bound will be infinity
      --gpu-manufacturer string                        GPU Manufacturer name (Example: NVIDIA)
      --gpu-memory-total string                        Number of GPUs' total memory (Example: 4 GiB) (sets --gpu-memory-total-min and -max to the same value)
      --gpu-memory-total-max string                    Maximum Number of GPUs' total memory (Example: 4 GiB) If --gpu-memory-total-min is not specified, the lower bound will be 0
      --gpu-memory-total-min string                    Minimum Number of GPUs' total memory (Example: 4 GiB) If --gpu-memory-total-max is not specified, the upper bound will be infinity
      --gpu-model string                               GPU Model name (Example: K520)
  -g, --gpus int32                                     Total Number of GPUs (Example: 4) (sets --gpus-min and -max to the same value)
      --gpus-max int32                                 Maximum Total Number of GPUs (Example: 4) If --gpus-min is not specified, the lower bound will be 0
      --gpus-min int32                                 Minimum Total Number of GPUs (Example: 4) If --gpus-max is not specified, the upper bound will be infinity
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
      --instance-types strings                         Instance Type names (must be exact, use allow-list for regex)
      --ipv6                                           Instance Types that support IPv6
  -m, --memory string                                  Amount of Memory available (Example: 4 GiB) (sets --memory-min and -max to the same value)
      --memory-max string                              Maximum Amount of Memory available (Example: 4 GiB) If --memory-min is not specified, the lower bound will be 0
      --memory-min string                              Minimum Amount of Memory available (Example: 4 GiB) If --memory-max is not specified, the upper bound will be infinity
      --network-encryption                             Instance Types that support automatic network encryption in-transit
      --network-interfaces int32                       Number of network interfaces (ENIs) that can be attached to the instance (sets --network-interfaces-min and -max to the same value)
      --network-interfaces-max int32                   Maximum Number of network interfaces (ENIs) that can be attached to the instance If --network-interfaces-min is not specified, the lower bound will be 0
      --network-interfaces-min int32                   Minimum Number of network interfaces (ENIs) that can be attached to the instance If --network-interfaces-max is not specified, the upper bound will be infinity
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
  -c, --vcpus int32                                    Number of vcpus available to the instance type. (sets --vcpus-min and -max to the same value)
      --vcpus-max int32                                Maximum Number of vcpus available to the instance type. If --vcpus-min is not specified, the lower bound will be 0
      --vcpus-min int32                                Minimum Number of vcpus available to the instance type. If --vcpus-max is not specified, the upper bound will be infinity
      --vcpus-to-memory-ratio string                   The ratio of vcpus to GiBs of memory. (Example: 1:2)
      --virtualization-type string                     Virtualization Type supported: [hvm or pv]


Suite Flags:
      --base-instance-type string   Instance Type used to retrieve similarly spec'd instance types
      --flexible                    Retrieves a group of instance types spanning multiple generations based on opinionated defaults and user overridden resource filters
      --service string              Filter instance types based on service support (Example: emr-5.20.0)


Global Flags:
      --cache-dir string        Directory to save the pricing and instance type caches (default "~/.ec2-instance-selector/")
      --cache-ttl int           Cache TTLs in hours for pricing and instance type caches. Setting the cache to 0 will turn off caching and cleanup any on-disk caches.
      --debug                   Debug - prints debug log messages
  -h, --help                    Help
      --max-results int         The maximum number of instance types that match your criteria to return (default 20)
  -o, --output string           Specify the output format (table, table-wide, one-line, interactive)
      --profile string          AWS CLI profile to use for credentials and config
  -r, --region string           AWS Region to use for API requests (NOTE: if not passed in, uses AWS SDK default precedence)
      --sort-by string          Specify the field to sort by. Quantity flags present in this CLI (memory, gpus, etc.) or a JSON path to the appropriate instance type field (Ex: ".MemoryInfo.SizeInMiB") is acceptable. (default ".InstanceType")
      --sort-direction string   Specify the direction to sort in (ascending, asc, descending, desc) (default "ascending")
  -v, --verbose                 Verbose - will print out full instance specs
      --version                 Prints CLI version
```


### Go Library

This is a minimal example of using the instance selector go package directly:

**cmd/examples/example1.go**
```go#cmd/examples/example1.go
package main

import (
	"context"
	"fmt"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector"
	"github.com/aws/aws-sdk-go-v2/config"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	// Initialize a context for the application
	ctx := context.Background()

	// Load an AWS session by looking at shared credentials or environment variables
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-2"))
	if err != nil {
		fmt.Printf("Oh no, AWS session credentials cannot be found: %v", err)
		return
	}

	// Instantiate a new instance of a selector with the AWS session
	instanceSelector, err := selector.New(ctx, cfg)
	if err != nil {
		fmt.Printf("Oh no, there was an error :( %v", err)
		return
	}

	// Instantiate an int range filter to specify min and max vcpus
	vcpusRange := selector.Int32RangeFilter{
		LowerBound: 2,
		UpperBound: 4,
	}
	// Instantiate a byte quantity range filter to specify min and max memory in GiB
	memoryRange := selector.ByteQuantityRangeFilter{
		LowerBound: bytequantity.FromGiB(2),
		UpperBound: bytequantity.FromGiB(4),
	}
	// Create a variable for the CPU Architecture so that it can be passed as a pointer
	// when creating the Filter struct
	cpuArch := ec2types.ArchitectureTypeX8664

	// Create a Filter struct with criteria you would like to filter
	// The full struct definition can be found here for all of the supported filters:
	// https://github.com/aws/amazon-ec2-instance-selector/blob/main/pkg/selector/types.go
	filters := selector.Filters{
		VCpusRange:      &vcpusRange,
		MemoryRange:     &memoryRange,
		CPUArchitecture: &cpuArch,
	}

	// Pass the Filter struct to the Filter function of your selector instance
	instanceTypesSlice, err := instanceSelector.Filter(ctx, filters)
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
[c4.large c5.large c5a.large c5ad.large c5d.large c6a.large c6i.large c6id.large c6in.large c7a.large c7i-flex.large c7i.large t2.medium t3.medium t3.small t3a.medium t3a.small]
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
