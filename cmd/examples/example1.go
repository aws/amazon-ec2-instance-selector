package main

import (
	"fmt"

	"github.com/aws/amazon-ec2-instance-selector/pkg/selector"
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
	// Instantiate a float64 range filter to specify min and max memory in GiB
	memoryRange := selector.Float64RangeFilter{
		LowerBound: 1.0,
		UpperBound: 4.0,
	}
	// Create a string for the CPU Architecture so that it can be passed as a pointer
	// when creating the Filter struct
	cpuArch := "x86_64"

	// Create a Filter struct with criteria you would like to filter
	// The full struct definition can be found here for all of the supported filters:
	// https://github.com/aws/amazon-ec2-instance-selector/blob/master/pkg/selector/types.go
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
