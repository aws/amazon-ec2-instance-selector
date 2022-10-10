package main

import (
	"context"
	"fmt"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
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
	instanceSelector := selector.New(ctx, cfg)

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
