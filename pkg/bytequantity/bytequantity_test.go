// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package bytequantity_test

import (
	"fmt"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/bytequantity"
	h "github.com/aws/amazon-ec2-instance-selector/v3/pkg/test"
)

func TestParseToByteQuantity(t *testing.T) {
	for _, testQuantity := range []string{"10mb", "10 mb", "10.0 mb", "10.0mb", "10m", "10mib", "10 M", "10.000 MiB"} {
		expectationVal := uint64(10)
		bq, err := bytequantity.ParseToByteQuantity(testQuantity)
		h.Ok(t, err)
		h.Assert(t, bq.Quantity == expectationVal, "quantity should have been %d, got %d instead on string %s", expectationVal, bq.Quantity, testQuantity)
	}

	for _, testQuantity := range []string{"4", "4.0", "4gb", "4 gb", "4.0 gb", "4.0gb", "4g", "4gib", "4 G", "4.000 GiB"} {
		expectationVal := uint64(4096)
		bq, err := bytequantity.ParseToByteQuantity(testQuantity)
		h.Ok(t, err)
		h.Assert(t, bq.Quantity == expectationVal, "quantity should have been %d, got %d instead on string %s", expectationVal, bq.Quantity, testQuantity)
	}

	for _, testQuantity := range []string{"109tb", "109 tb", "109.0 tb", "109.0tb", "109t", "109tib", "109 T", "109.000 TiB"} {
		expectationVal := uint64(114294784)
		bq, err := bytequantity.ParseToByteQuantity(testQuantity)
		h.Ok(t, err)
		h.Assert(t, bq.Quantity == expectationVal, "quantity should have been %d, got %d instead on string %s", expectationVal, bq.Quantity, testQuantity)
	}

	expectationVal := uint64(1025)
	testQuantity := "1.001 gb"
	bq, err := bytequantity.ParseToByteQuantity(testQuantity)
	h.Ok(t, err)
	h.Assert(t, bq.Quantity == expectationVal, "quantity should have been %d, got %d instead on string %s", expectationVal, bq.Quantity, testQuantity)

	// Only supports 3 decimal places
	bq, err = bytequantity.ParseToByteQuantity("109.0001")
	h.Nok(t, err)

	// Only support decimals on GiB and TiB
	bq, err = bytequantity.ParseToByteQuantity("109.001 mib")
	h.Nok(t, err)

	// Overflow a uint64
	overflow := "18446744073709551616"
	bq, err = bytequantity.ParseToByteQuantity(fmt.Sprintf("%s mib", overflow))
	h.Nok(t, err)

	bq, err = bytequantity.ParseToByteQuantity(fmt.Sprintf("%s gib", overflow))
	h.Nok(t, err)

	bq, err = bytequantity.ParseToByteQuantity(fmt.Sprintf("%s tib", overflow))
	h.Nok(t, err)

	// Unit not supported
	bq, err = bytequantity.ParseToByteQuantity("1 NS")
	h.Nok(t, err)
}

func TestStringGiB(t *testing.T) {
	expectedVal := "0.098 GiB"
	testVal := uint64(100)
	bq := bytequantity.ByteQuantity{Quantity: testVal}
	h.Assert(t, bq.StringGiB() == expectedVal, "%d MiB should equal %s, instead got %s", testVal, expectedVal, bq.StringGiB())

	expectedVal = "1.000 GiB"
	testVal = uint64(1024)
	bq = bytequantity.ByteQuantity{Quantity: 1024}
	h.Assert(t, bq.StringGiB() == expectedVal, "%d MiB should equal %s, instead got %s", testVal, expectedVal, bq.StringGiB())
}

func TestStringTiB(t *testing.T) {
	expectedVal := "1.000 TiB"
	testVal := uint64(1048576)
	bq := bytequantity.ByteQuantity{Quantity: testVal}
	h.Assert(t, bq.StringTiB() == expectedVal, "%d MiB should equal %s, instead got %s", testVal, expectedVal, bq.StringTiB())

	expectedVal = "0.005 TiB"
	testVal = uint64(5240)
	bq = bytequantity.ByteQuantity{Quantity: testVal}
	h.Assert(t, bq.StringTiB() == expectedVal, "%d MiB should equal %s, instead got %s", testVal, expectedVal, bq.StringTiB())
}

func TestStringMiB(t *testing.T) {
	expectedVal := "1 MiB"
	testVal := uint64(1)
	bq := bytequantity.ByteQuantity{Quantity: testVal}
	h.Assert(t, bq.StringMiB() == expectedVal, "%d MiB should equal %s, instead got %s", testVal, expectedVal, bq.StringMiB())

	expectedVal = "2 MiB"
	testVal = uint64(2)
	bq = bytequantity.ByteQuantity{Quantity: testVal}
	h.Assert(t, bq.StringMiB() == expectedVal, "%d MiB should equal %s, instead got %s", testVal, expectedVal, bq.StringMiB())
}

func TestFromMiB(t *testing.T) {
	expectedVal := uint64(1)
	bq := bytequantity.FromMiB(expectedVal)
	h.Assert(t, bq.MiB() == float64(expectedVal), "%d MiB should equal %d, instead got %s", expectedVal, expectedVal, bq.StringMiB())
}

func TestFromGiB(t *testing.T) {
	expectedVal := float64(1.0)
	testVal := uint64(1)
	bq := bytequantity.FromGiB(testVal)
	h.Assert(t, bq.GiB() == expectedVal, "%d GiB should equal %d, instead got %s", expectedVal, expectedVal, bq.StringGiB())
}

func TestFromTiB(t *testing.T) {
	expectedVal := float64(1.0)
	testVal := uint64(1)
	bq := bytequantity.FromTiB(testVal)
	h.Assert(t, bq.TiB() == expectedVal, "%d TiB should equal %d, instead got %s", expectedVal, expectedVal, bq.StringTiB())
}
