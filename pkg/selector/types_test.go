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

package selector_test

import (
	"regexp"
	"strings"
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/v3/pkg/test"
)

// Tests

func TestMarshalIndent(t *testing.T) {
	cpuArch := ec2types.ArchitectureTypeX8664
	allowRegex := "^abc$"
	denyRegex := "^zyx$"

	filters := selector.Filters{
		AllowList:       regexp.MustCompile(allowRegex),
		DenyList:        regexp.MustCompile(denyRegex),
		CPUArchitecture: &cpuArch,
	}
	out, err := filters.MarshalIndent("", "    ")
	outStr := string(out)
	h.Ok(t, err)
	h.Assert(t, strings.Contains(outStr, "AllowList") && strings.Contains(outStr, allowRegex), "Does not include AllowList regex string")
	h.Assert(t, strings.Contains(outStr, "DenyList") && strings.Contains(outStr, denyRegex), "Does not include DenyList regex string")
}

func TestMarshalIndent_nil(t *testing.T) {
	denyRegex := "^zyx$"

	filters := selector.Filters{
		AllowList: nil,
		DenyList:  regexp.MustCompile(denyRegex),
	}
	out, err := filters.MarshalIndent("", "    ")
	outStr := string(out)
	h.Ok(t, err)
	h.Assert(t, strings.Contains(outStr, "AllowList") && strings.Contains(outStr, "null"), "Does not include AllowList null entry")
	h.Assert(t, strings.Contains(outStr, "DenyList") && strings.Contains(outStr, denyRegex), "Does not include DenyList regex string")
}
