package config

import (
	"testing"

	"github.com/openshift/openshift-azure/pkg/api"
)

func TestOpenShiftVersion(t *testing.T) {
	tests := map[string]struct {
		testVersion string
		testResult  string
		testError   string
	}{
		"valid version string": {
			testVersion: "311.14.20180101",
			testResult:  "v3.11.14",
		},
		"invalid version string extra periods": {
			testVersion: "3.1.0.1.4",
			testError:   `invalid imageVersion "3.1.0.1.4"`,
		},
		"invalid version string short first part": {
			testVersion: "3.14.20180101",
			testError:   `invalid imageVersion "3.14.20180101"`,
		},
	}

	for label, test := range tests {
		r, err := openShiftVersion(test.testVersion)
		if test.testError != "" {
			if err == nil || err.Error() != test.testError {
				t.Errorf("%s: got error %s, expected %s", label, err, test.testError)
			}
		} else {
			if r != test.testResult {
				t.Errorf("%s: got %s, expected %s", label, r, test.testResult)
			}
		}
	}
}

func TestNodeImageVersion(t *testing.T) {
	for _, deployOS := range []string{"", "rhel7", "centos7"} {
		cs := &api.OpenShiftManagedCluster{
			Properties: api.Properties{
				OpenShiftVersion: "v3.11",
			},
		}
		cg := simpleGenerator{
			pluginConfig: api.PluginConfig{
				TestConfig: api.TestConfig{
					DeployOS: deployOS,
				},
			},
		}
		cg.selectNodeImage(cs)
		if cs.Config.ImageVersion == "latest" {
			t.Errorf("cs.Config.ImageVersion should not equal latest")
		}
	}
}
