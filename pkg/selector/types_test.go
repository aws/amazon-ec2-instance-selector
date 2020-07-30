// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package selector_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
)

// Tests

func TestMarshalIndent(t *testing.T) {
	cpuArch := "x86_64"
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
