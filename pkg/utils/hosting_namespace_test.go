package utils

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation"
)

func TestHostedInstanceID(t *testing.T) {
	id := HostedInstanceID("cluster-a", "helloworldhosted")
	if len(id) != 20 {
		t.Fatalf("expected 20 characters, got %q", id)
	}
	if id != HostedInstanceID("cluster-a", "helloworldhosted") {
		t.Errorf("expected a stable ID")
	}
	if id == HostedInstanceID("cluster-b", "helloworldhosted") {
		t.Errorf("expected different clusters to get different IDs")
	}
	// Length-delimited inputs: the same concatenation must not give the same ID.
	if HostedInstanceID("a", "bc") == HostedInstanceID("ab", "c") {
		t.Errorf("expected ambiguous pairs to get different IDs")
	}
}

func TestHostingNamespace(t *testing.T) {
	cases := []struct {
		name, cluster, addon, wantPrefix string
	}{
		{name: "simple", cluster: "cluster-a", addon: "helloworldhosted",
			wantPrefix: "open-cluster-management-helloworldhosted-"},
		{name: "long addon name is cut to 18 characters", cluster: "cluster-a", addon: "a-very-long-addon-name-for-testing",
			wantPrefix: "open-cluster-management-a-very-long-addon-"},
		{name: "punctuation becomes one dash", cluster: "cluster-a", addon: "My.Addon__Name",
			wantPrefix: "open-cluster-management-my-addon-name-"},
		{name: "no letters or digits", cluster: "cluster-a", addon: "...",
			wantPrefix: "open-cluster-management-addon-"},
		{name: "long cluster name", cluster: strings.Repeat("c", 250), addon: "helloworldhosted",
			wantPrefix: "open-cluster-management-helloworldhosted-"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ns := HostingNamespace(c.cluster, c.addon)
			if !strings.HasPrefix(ns, c.wantPrefix) {
				t.Errorf("expected prefix %q, got %q", c.wantPrefix, ns)
			}
			if errs := validation.IsDNS1123Label(ns); len(errs) != 0 {
				t.Errorf("%q is not a valid namespace name: %v", ns, errs)
			}
			if !strings.HasSuffix(ns, "-"+HostedInstanceID(c.cluster, c.addon)) {
				t.Errorf("expected %q to end with the instance ID", ns)
			}
		})
	}
}
