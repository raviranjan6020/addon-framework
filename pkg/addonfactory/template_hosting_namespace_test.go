package addonfactory

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	addonapiv1beta1 "open-cluster-management.io/api/addon/v1beta1"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"

	"open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/addon-framework/pkg/utils"
)

func TestTemplateHostingNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clusterv1alpha1.Install(scheme)
	_ = appsv1.AddToScheme(scheme)

	cases := []struct {
		name          string
		enabled       bool
		hosted        bool
		wantNamespace bool
	}{
		{name: "enabled and hosted", enabled: true, hosted: true, wantNamespace: true},
		{name: "enabled but default mode", enabled: true, hosted: false},
		{name: "hosted but not enabled", enabled: false, hosted: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			addon := NewFakeManagedClusterAddon("helloworld", "cluster-a", "", "")
			if c.hosted {
				addon.Annotations[addonapiv1beta1.HostingClusterNameAnnotationKey] = "hosting"
			}
			factory := NewAgentAddonFactory("helloworld", templateFS, "testmanifests/template").
				WithScheme(scheme).
				WithAgentRegistrationOption(&agent.RegistrationOption{}).
				WithAgentHostedModeEnabledOption()
			if c.enabled {
				factory = factory.WithHostingNamespaceEnabledOption()
			}
			agentAddon, err := factory.BuildTemplateAgentAddon()
			if err != nil {
				t.Fatal(err)
			}
			objects, err := agentAddon.Manifests(context.TODO(), NewFakeManagedCluster("cluster-a", "1.30.0"), addon)
			if err != nil {
				t.Fatal(err)
			}

			want := utils.HostingNamespace("cluster-a", "helloworld")
			var found *corev1.Namespace
			for _, o := range objects {
				if ns, ok := o.(*corev1.Namespace); ok && ns.Name == want {
					found = ns
				}
			}
			if !c.wantNamespace {
				if found != nil {
					t.Fatalf("expected no generated namespace, got %s", found.Name)
				}
				return
			}
			if found == nil {
				t.Fatalf("expected generated namespace %s in %d objects", want, len(objects))
			}
			if got := found.Annotations[addonapiv1beta1.HostedManifestLocationAnnotationKey]; got != "hosting" {
				t.Errorf("expected the namespace to go to the hosting cluster, got location %q", got)
			}
			if got := found.Annotations[hostingAddonAnnotationKey]; got != "cluster-a/helloworld" {
				t.Errorf("expected owner annotation cluster-a/helloworld, got %q", got)
			}
		})
	}
}
