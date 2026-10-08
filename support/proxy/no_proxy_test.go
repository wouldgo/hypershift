package proxy

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	configv1 "github.com/openshift/api/config/v1"
	hyperv1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

func TestDefaultNoProxyForProxy(t *testing.T) {
	hcp := &hyperv1.HostedControlPlane{
		Spec: hyperv1.HostedControlPlaneSpec{
			Configuration: &hyperv1.ClusterConfiguration{
				Proxy: &configv1.ProxySpec{NoProxy: "cluster-only.example.com"},
			},
		},
	}

	tests := []struct {
		name        string
		proxyConfig *configv1.ProxySpec
		expect      []string
		absent      []string
	}{
		{
			name:        "nil proxy still returns the cluster-internal defaults",
			proxyConfig: nil,
			expect:      []string{"127.0.0.1", "localhost", ".svc", ".cluster.local"},
			absent:      []string{"cluster-only.example.com"},
		},
		{
			name:        "supplied NoProxy is folded in alongside the defaults",
			proxyConfig: &configv1.ProxySpec{NoProxy: "idp.example.com,proxy.example.com"},
			expect:      []string{".cluster.local", "idp.example.com", "proxy.example.com"},
			absent:      []string{"cluster-only.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got := DefaultNoProxyForProxy(hcp, tt.proxyConfig)
			entries := strings.Split(got, ",")
			for _, e := range tt.expect {
				g.Expect(entries).To(ContainElement(e))
			}
			for _, e := range tt.absent {
				g.Expect(entries).NotTo(ContainElement(e))
			}
		})
	}
}

func TestDefaultNoProxyUsesClusterWide(t *testing.T) {
	g := NewWithT(t)
	hcp := &hyperv1.HostedControlPlane{
		Spec: hyperv1.HostedControlPlaneSpec{
			Configuration: &hyperv1.ClusterConfiguration{
				Proxy: &configv1.ProxySpec{NoProxy: "cluster-wide.example.com"},
			},
		},
	}
	entries := strings.Split(DefaultNoProxy(hcp), ",")
	g.Expect(entries).To(ContainElement("cluster-wide.example.com"))
	g.Expect(entries).To(ContainElement(".cluster.local"))
}
