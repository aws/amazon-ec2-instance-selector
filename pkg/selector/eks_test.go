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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	h "github.com/aws/amazon-ec2-instance-selector/v2/pkg/test"
)

const (
	githubStaticReleasesDir = "GithubEKSAMIRelease"
	githubReleaseVersion    = "20210125"
	githubZipFileName       = "amazon-eks-ami-20210125.zip"
)

// Tests

func TestEKSDefaultService(t *testing.T) {
	ghServer := eksGithubReleaseHTTPServer(false, false)
	defer ghServer.Close()

	registry := selector.NewRegistry()
	registry.Register("eks", &selector.EKS{
		AMIRepoURL: ghServer.URL,
	})

	eks := "eks"
	filters := selector.Filters{
		Service: &eks,
	}

	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, len(*transformedFilters.InstanceTypes) == 389, "389 instance types should be supported, but got %d", len(*transformedFilters.InstanceTypes))
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "eks should only support hvm")

	eks = "eks-v" + githubReleaseVersion
	filters.Service = &eks
	transformedFilters, err = registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, len(*transformedFilters.InstanceTypes) == 389, "389 instance types should be supported, but got %d", len(*transformedFilters.InstanceTypes))
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "eks should only support hvm")
}

func TestEKSDefaultService_FailLatestReleaseUseFallbackStaticVersion(t *testing.T) {
	ghServer := eksGithubReleaseHTTPServer(true, false)
	defer ghServer.Close()

	registry := selector.NewRegistry()
	registry.Register("eks", &selector.EKS{
		AMIRepoURL: ghServer.URL,
	})

	eks := "eks"
	filters := selector.Filters{
		Service: &eks,
	}

	transformedFilters, err := registry.ExecuteTransforms(filters)
	h.Ok(t, err)
	h.Assert(t, transformedFilters != filters, " Filters should have been modified")
	h.Assert(t, len(*transformedFilters.InstanceTypes) == 389, "389 instance types should be supported, but got %d", len(*transformedFilters.InstanceTypes))
	h.Assert(t, *transformedFilters.VirtualizationType == "hvm", "eks should only support hvm")
}

func TestEKSDefaultService_FailLatestReleaseAndFailExactVersionLookup(t *testing.T) {
	ghServer := eksGithubReleaseHTTPServer(true, true)
	defer ghServer.Close()

	registry := selector.NewRegistry()
	registry.Register("eks", &selector.EKS{
		AMIRepoURL: ghServer.URL,
	})

	eks := "eks"
	filters := selector.Filters{
		Service: &eks,
	}

	_, err := registry.ExecuteTransforms(filters)
	h.Nok(t, err)
}

// Test Helpers Functions

func eksGithubReleaseHTTPServer(failLatestRelease bool, failExactRelease bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Got Path: " + r.URL.Path)
		if r.URL.Path == "/releases/latest" {
			if failLatestRelease {
				w.WriteHeader(404)
				return
			}
			w.WriteHeader(302)
			w.Header().Add("location", "/releases/tag/v"+githubReleaseVersion)
			return
		}
		if r.URL.Path == "/archive/v"+githubReleaseVersion+".zip" {
			if failExactRelease {
				w.WriteHeader(404)
				return
			}
			ghReleaseZipPath := fmt.Sprintf("%s/%s/%s", mockFilesPath, githubStaticReleasesDir, githubZipFileName)
			eksAMIReleaseZipFile, err := ioutil.ReadFile(ghReleaseZipPath)
			if err != nil {
				panic("Could not read EKS AMI release zip file")
			}
			w.Write(eksAMIReleaseZipFile)
			return
		}
	}))
}
