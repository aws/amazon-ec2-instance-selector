// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package selector

import (
	"math"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	h "github.com/aws/amazon-ec2-instance-selector/v3/pkg/test"
)

func TestIsSupportedFromStrings_Supported(t *testing.T) {
	arm64 := aws.String("arm64")
	instanceTypeArchitectures := []*string{arm64}
	isSupported := isSupportedFromStrings(instanceTypeArchitectures, arm64)
	h.Assert(t, isSupported == true, "arm64 should be a supported cpu architecture")
}

func TestIsSupportedFromStrings_Nil(t *testing.T) {
	arm64 := aws.String("arm64")
	isSupported := isSupportedFromStrings(nil, arm64)
	h.Assert(t, isSupported == false, "arm64 should NOT be a supported cpu architecture")
}

func TestIsSupportedFromStrings_NilTarget(t *testing.T) {
	instanceTypeArchitectures := []*string{aws.String("arm64")}
	isSupported := isSupportedFromStrings(instanceTypeArchitectures, nil)
	h.Assert(t, isSupported == true, "arm64 should be a supported cpu architecture")
}

func TestIsSupportedFromString_Supported(t *testing.T) {
	hypervisor := aws.String("nitro")
	nitro := aws.String("nitro")
	isSupported := isSupportedFromString(nitro, hypervisor)
	h.Assert(t, isSupported == true, "nitro should be the supported hypervisor")
}

func TestIsSupportedFromString_Nil(t *testing.T) {
	hypervisor := aws.String("nitro")
	isSupported := isSupportedFromString(nil, hypervisor)
	h.Assert(t, isSupported == false, "nil source should NOT be supported for specified target string")
}

func TestIsSupportedFromString_NilTarget(t *testing.T) {
	nitro := aws.String("nitro")
	isSupported := isSupportedFromString(nitro, nil)
	h.Assert(t, isSupported == true, "nil target should be supported for specified source string")
}

func TestIsSupportedWithBool(t *testing.T) {
	hibernationSupported := aws.Bool(true)
	userFilter := aws.Bool(true)
	isSupported := isSupportedWithBool(hibernationSupported, userFilter)
	h.Assert(t, isSupported == true, "Hibernation should be supported")
}

func TestIsSupportedWithBool_Nil(t *testing.T) {
	hibernationSupported := aws.Bool(false)
	isSupported := isSupportedWithBool(hibernationSupported, nil)
	h.Assert(t, isSupported == true, "Hibernation should be supported")
}

func TestIsSupportedWithBool_Unsupported(t *testing.T) {
	hibernationSupported := aws.Bool(false)
	userFilter := aws.Bool(true)
	isSupported := isSupportedWithBool(hibernationSupported, userFilter)
	h.Assert(t, isSupported == false, "Hibernation should NOT be supported")
}

func TestIsSupportedWithRangeInt_SupportedExact(t *testing.T) {
	target := IntRangeFilter{LowerBound: 4, UpperBound: 4}
	isSupported := isSupportedWithRangeInt(aws.Int(4), &target)
	h.Assert(t, isSupported == true, "IntRangeFilter should match exactly")
}

func TestIsSupportedWithRangeInt_SupportedAround(t *testing.T) {
	target := IntRangeFilter{LowerBound: 2, UpperBound: 6}
	isSupported := isSupportedWithRangeInt(aws.Int(4), &target)
	h.Assert(t, isSupported == true, "IntRangeFilter should match with lower and upper bound around the desired source")
}

func TestIsSupportedWithRangeInt_Nil(t *testing.T) {
	target := IntRangeFilter{LowerBound: 2, UpperBound: 6}
	isSupported := isSupportedWithRangeInt(nil, &target)
	h.Assert(t, isSupported == false, "IntRangeFilter should NOT match with nil source")
}

func TestIsSupportedWithRangeInt_NilTarget(t *testing.T) {
	isSupported := isSupportedWithRangeInt(aws.Int(4), nil)
	h.Assert(t, isSupported == true, "IntRangeFilter should match with nil target")
}

func TestIsSupportedWithRangeInt_BothNil(t *testing.T) {
	isSupported := isSupportedWithRangeInt(nil, nil)
	h.Assert(t, isSupported == true, "IntRangeFilter should match with nil target and nil source")
}

func TestIsSupportedWithRangeInt_SourceNilTarget0(t *testing.T) {
	target := IntRangeFilter{LowerBound: 0, UpperBound: 0}
	isSupported := isSupportedWithRangeInt(nil, &target)
	h.Assert(t, isSupported == true, "IntRangeFilter should match with 0 target and nil source")
}

// ==================

func TestIsSupportedWithRangeInt64_SupportedExact(t *testing.T) {
	target := IntRangeFilter{LowerBound: 4, UpperBound: 4}
	isSupported := isSupportedWithRangeInt64(aws.Int64(4), &target)
	h.Assert(t, isSupported == true, "IntRangeFilter should match exactly")
}

func TestIsSupportedWithRangeInt64_SupportedAround(t *testing.T) {
	target := IntRangeFilter{LowerBound: 2, UpperBound: 6}
	isSupported := isSupportedWithRangeInt64(aws.Int64(4), &target)
	h.Assert(t, isSupported == true, "IntRangeFilter should match with lower and upper bound around the desired source")
}

func TestIsSupportedWithRangeInt64_Nil(t *testing.T) {
	target := IntRangeFilter{LowerBound: 2, UpperBound: 6}
	isSupported := isSupportedWithRangeInt64(nil, &target)
	h.Assert(t, isSupported == false, "IntRangeFilter should NOT match with nil source")
}

func TestIsSupportedWithRangeInt64_NilTarget(t *testing.T) {
	isSupported := isSupportedWithRangeInt64(aws.Int64(4), nil)
	h.Assert(t, isSupported == true, "IntRangeFilter should match with nil target")
}

func TestIsSupportedWithRangeInt64_BothNil(t *testing.T) {
	isSupported := isSupportedWithRangeInt64(nil, nil)
	h.Assert(t, isSupported == true, "IntRangeFilter should match with nil target and nil source")
}

func TestIsSupportedWithRangeInt64_SourceNilTarget0(t *testing.T) {
	target := IntRangeFilter{LowerBound: 0, UpperBound: 0}
	isSupported := isSupportedWithRangeInt64(nil, &target)
	h.Assert(t, isSupported == true, "IntRangeFilter should match with 0 target and nil source")
}

// uint64

func TestIsSupportedWithRangeUint64_SupportedExact(t *testing.T) {
	target := IntRangeFilter{LowerBound: 4, UpperBound: 4}
	isSupported := isSupportedWithRangeInt64(aws.Int64(4), &target)
	h.Assert(t, isSupported == true, "IntRangeFilter should match exactly")
}

func TestIsSupportedWithRangeUint64_SupportedAround(t *testing.T) {
	target := Uint64RangeFilter{LowerBound: 2, UpperBound: 6}
	isSupported := isSupportedWithRangeUint64(aws.Int64(4), &target)
	h.Assert(t, isSupported == true, "UintRangeFilter should match with lower and upper bound around the desired source")
}

func TestIsSupportedWithRangeUint64_Nil(t *testing.T) {
	target := Uint64RangeFilter{LowerBound: 2, UpperBound: 6}
	isSupported := isSupportedWithRangeUint64(nil, &target)
	h.Assert(t, isSupported == false, "Uint64RangeFilter should NOT match with nil source")
}

func TestIsSupportedWithRangeUint64_NilTarget(t *testing.T) {
	isSupported := isSupportedWithRangeUint64(aws.Int64(4), nil)
	h.Assert(t, isSupported == true, "Uint64RangeFilter should match with nil target")
}

func TestIsSupportedWithRangeUint64_BothNil(t *testing.T) {
	isSupported := isSupportedWithRangeUint64(nil, nil)
	h.Assert(t, isSupported == true, "Uint64RangeFilter should match with nil target and nil source")
}

func TestIsSupportedWithRangeUint64_SourceNilTarget0(t *testing.T) {
	target := Uint64RangeFilter{LowerBound: 0, UpperBound: 0}
	isSupported := isSupportedWithRangeUint64(nil, &target)
	h.Assert(t, isSupported == true, "Uint64RangeFilter should match with 0 target and nil source")
}

func TestIsSupportedWithRangeUint64_Overflow(t *testing.T) {
	target := Uint64RangeFilter{LowerBound: 0, UpperBound: math.MaxUint64}
	isSupported := isSupportedWithRangeUint64(aws.Int64(4), &target)
	h.Assert(t, isSupported == true, "Uint64RangeFilter should match with 0 - MAX target and source 4")
}

// float64

func TestIsSupportedWithFloat64_Supported(t *testing.T) {
	isSupported := isSupportedWithFloat64(aws.Float64(0.33), aws.Float64(0.33))
	h.Assert(t, isSupported == true, "Float64 comparison should match exactly with 2 decimal places")
}

func TestIsSupportedWithFloat64_SupportedTruncatedDecPlacesExact(t *testing.T) {
	isSupported := isSupportedWithFloat64(aws.Float64(0.3322), aws.Float64(0.3322))
	h.Assert(t, isSupported == true, "Float64 comparison should match exactly with 4 decimal places")
}

func TestIsSupportedWithFloat64_SupportedTruncatedDecPlaces(t *testing.T) {
	isSupported := isSupportedWithFloat64(aws.Float64(0.3399), aws.Float64(0.3311))
	h.Assert(t, isSupported == true, "Float64 comparison should match when truncating to 2 decimal places")
}

func TestIsSupportedWithFloat64_Unsupported(t *testing.T) {
	isSupported := isSupportedWithFloat64(aws.Float64(0.4), aws.Float64(0.3399))
	h.Assert(t, isSupported == false, "Float64 comparison should NOT match")
}

func TestIsSupportedWithFloat64_SourceNil(t *testing.T) {
	isSupported := isSupportedWithFloat64(nil, aws.Float64(0.3399))
	h.Assert(t, isSupported == false, "Float64 comparison should NOT match with nil source")
}

func TestIsSupportedWithFloat64_TargetNil(t *testing.T) {
	isSupported := isSupportedWithFloat64(aws.Float64(0.3399), nil)
	h.Assert(t, isSupported == true, "Float64 comparison should match with nil target")
}

func TestIsSupportedWithFloat64_BothNil(t *testing.T) {
	isSupported := isSupportedWithFloat64(nil, nil)
	h.Assert(t, isSupported == true, "Float64 comparison should match with nil target and source")
}

// bools

func TestSupportSyntaxToBool_Supported(t *testing.T) {
	isSupported := supportSyntaxToBool(aws.String("supported"))
	h.Assert(t, *isSupported == true, "Supported should evaluate to true")
}

func TestSupportSyntaxToBool_Required(t *testing.T) {
	isSupported := supportSyntaxToBool(aws.String("required"))
	h.Assert(t, *isSupported == true, "Required should evaluate to true")
}

func TestSupportSyntaxToBool_Unsupported(t *testing.T) {
	isSupported := supportSyntaxToBool(aws.String("unsupported"))
	h.Assert(t, *isSupported == false, "Unsupported should evaluate to false")
}

func TestSupportSyntaxToBool_UnsupportedCaps(t *testing.T) {
	isSupported := supportSyntaxToBool(aws.String("SuPpOrTeD"))
	h.Assert(t, *isSupported == true, "Supported with weird casing should evaluate to true")
}

func TestSupportSyntaxToBool_ArbitraryString(t *testing.T) {
	isSupported := supportSyntaxToBool(aws.String("blah"))
	h.Assert(t, *isSupported == false, "Arbitrary string should evaluate to false")
}

func TestSupportSyntaxToBool_Nil(t *testing.T) {
	isSupported := supportSyntaxToBool(nil)
	h.Assert(t, isSupported == nil, "nil should evaluate to nil")
}

func TestCalculateVCpusToMemoryRatio(t *testing.T) {
	vcpus := aws.Int32(4)
	memory := aws.Int64(4096)
	ratio := calculateVCpusToMemoryRatio(vcpus, memory)
	h.Assert(t, *ratio == 1.00, "ratio should equal 1:1")

	vcpus = aws.Int32(2)
	memory = aws.Int64(4096)
	ratio = calculateVCpusToMemoryRatio(vcpus, memory)
	h.Assert(t, *ratio == 2.00, "ratio should equal 1:2")

	vcpus = aws.Int32(1)
	memory = aws.Int64(512)
	ratio = calculateVCpusToMemoryRatio(vcpus, memory)
	h.Assert(t, *ratio == 1.0, "ratio should take the ceiling which equals 1:1")

	vcpus = aws.Int32(0)
	memory = aws.Int64(512)
	ratio = calculateVCpusToMemoryRatio(vcpus, memory)
	h.Assert(t, ratio == nil, "ratio should be nil when vcpus is 0")
}

func TestCalculateVCpusToMemoryRatio_Nil(t *testing.T) {
	memory := aws.Int64(4096)
	ratio := calculateVCpusToMemoryRatio(nil, memory)
	h.Assert(t, ratio == nil, "nil vcpus should evaluate to nil")

	vcpus := aws.Int32(2)
	ratio = calculateVCpusToMemoryRatio(vcpus, nil)
	h.Assert(t, ratio == nil, "nil memory should evaluate to nil")

	ratio = calculateVCpusToMemoryRatio(nil, nil)
	h.Assert(t, ratio == nil, "nil vcpus and memory should evaluate to nil")
}

func TestGetNetworkPerformance(t *testing.T) {
	netPerformance := getNetworkPerformance(aws.String("10 Gigabit"))
	h.Assert(t, *netPerformance == 10, "Networking performance should parse properly")

	netPerformance = getNetworkPerformance(aws.String("Up to 10 Gigabit"))
	h.Assert(t, *netPerformance == 10, "Networking performance should parse properly")

	netPerformance = getNetworkPerformance(aws.String("Up to 10 Gigabit"))
	h.Assert(t, *netPerformance == 10, "Networking performance should parse properly")

	netPerformance = getNetworkPerformance(aws.String("100 Gigabit"))
	h.Assert(t, *netPerformance == 100, "Networking performance should parse properly")

	netPerformance = getNetworkPerformance(aws.String("10 Gigabit abcd"))
	h.Assert(t, *netPerformance == 10, "Networking performance should parse properly when an arbitrary string is passed after quantity-unit syntax")

	netPerformance = getNetworkPerformance(aws.String("High"))
	h.Assert(t, *netPerformance == -1, "Networking performance should return -1 when not a number")

	netPerformance = getNetworkPerformance(aws.String(""))
	h.Assert(t, *netPerformance == -1, "Networking performance should parse properly when empty string")

	netPerformance = getNetworkPerformance(nil)
	h.Assert(t, *netPerformance == -1, "Networking performance should parse properly when nil")

	netPerformance = getNetworkPerformance(aws.String("abcd"))
	h.Assert(t, *netPerformance == -1, "Networking performance should parse properly when an arbitrary string is passed")
}
