package main

import (
	"fmt"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector/outputs"
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

	// Pass the Filter struct to the FilteredInstanceTypes function of your
	// selector instance to get a list of filtered instance types and their details.
	instanceTypesSlice, err := instanceSelector.FilterInstanceTypes(filters)
	if err != nil {
		fmt.Printf("Oh no, there was an error getting instance types: %v", err)
		return
	}

	// Pass in the list of instance type details to the SortInstanceTypes function
	// if you wish to sort the instances based on set filters.
	sortFilter := "instance-type-name"
	sortDirection := "ascending"
	instanceTypesSlice, err = instanceSelector.SortInstanceTypes(instanceTypesSlice, &sortFilter, &sortDirection)
	if err != nil {
		fmt.Printf("Oh no, there was an error sorting instance types: %v", err)
		return
	}

	// Truncate results and format them for output with your desired formatting function.
	// All formatting functions can be found here:
	// https://github.com/aws/amazon-ec2-instance-selector/blob/main/pkg/selector/outputs/outputs.go
	// Examples of formatted outputs can be found here:
	// https://github.com/aws/amazon-ec2-instance-selector#examples
	maxResults := 10
	instanceTypesSlice, _, err = outputs.TruncateResults(&maxResults, instanceTypesSlice)
	if err != nil {
		fmt.Printf("Oh no, there was an error truncating instance types: %v", err)
		return
	}
	instanceTypes := outputs.SimpleInstanceTypeOutput(instanceTypesSlice)

	// Print the returned instance types slice
	fmt.Println(instanceTypes)
}
