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

package cli

import (
	"os"
	"testing"

	"github.com/spf13/pflag"

	h "github.com/aws/amazon-ec2-instance-selector/v3/pkg/test"
)

// Tests

func TestRemoveIntersectingArgs(t *testing.T) {
	flagSet := pflag.NewFlagSet("test-flag-set", pflag.ContinueOnError)
	flagSet.Bool("test-bool", false, "test usage")
	os.Args = []string{"ec2-instance-selector", "--test-bool", "--this-should-stay"}
	newArgs := removeIntersectingArgs(flagSet)
	h.Assert(t, len(newArgs) == 2, "NewArgs should only include the bin name and one argument after removing intersections")
}

func TestRemoveIntersectingArgs_NextArg(t *testing.T) {
	flagSet := pflag.NewFlagSet("test-flag-set", pflag.ContinueOnError)
	flagSet.String("test-str", "", "test usage")
	os.Args = []string{"ec2-instance-selector", "--test-str", "somevalue", "--this-should-stay", "valuetostay"}
	newArgs := removeIntersectingArgs(flagSet)
	h.Assert(t, len(newArgs) == 3, "NewArgs should only include the bin name and a flag + input after removing intersections")
}

func TestRemoveIntersectingArgs_ShorthandArg(t *testing.T) {
	flagSet := pflag.NewFlagSet("test-flag-set", pflag.ContinueOnError)
	flagSet.StringP("test-str", "t", "", "test usage")
	os.Args = []string{"ec2-instance-selector", "--test-str", "somevalue", "--this-should-stay", "valuetostay", "-t", "test"}
	newArgs := removeIntersectingArgs(flagSet)
	h.Assert(t, len(newArgs) == 3, "NewArgs should only include the bin name and a flag + input after removing intersections")
}
