package instancetypes

import "github.com/aws/aws-sdk-go/service/ec2"

// Details hold all the information on an ec2 instance type
type Details struct {
	ec2.InstanceTypeInfo
	OndemandPricePerHour *float64
	SpotPrice            *float64
}
