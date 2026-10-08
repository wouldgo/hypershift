package proxy

import (
	"testing"

	. "github.com/onsi/gomega"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	hyperv1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

func TestEffectiveAuthProxy(t *testing.T) {
	clusterWide := &configv1.ProxySpec{
		HTTPProxy:  "http://cluster:3128",
		HTTPSProxy: "https://cluster:3128",
		NoProxy:    "svc.cluster.local",
		TrustedCA:  configv1.ConfigMapNameReference{Name: "cluster-ca"},
	}

	override := &operatorv1.AuthenticationProxyConfig{
		HTTPProxy:  "http://auth:8080",
		HTTPSProxy: "https://auth:8080",
		NoProxy:    []string{"kube-apiserver", "idp.example.com"},
		TrustedCA:  operatorv1.AuthenticationConfigMapReference{Name: "auth-ca"},
	}

	testCases := []struct {
		name     string
		hcp      *hyperv1.HostedControlPlane
		expected *configv1.ProxySpec
	}{
		{
			name:     "nil hcp returns nil",
			hcp:      nil,
			expected: nil,
		},
		{
			name:     "both nil returns nil",
			hcp:      &hyperv1.HostedControlPlane{},
			expected: nil,
		},
		{
			name: "override unset falls back to cluster-wide",
			hcp: &hyperv1.HostedControlPlane{
				Spec: hyperv1.HostedControlPlaneSpec{
					Configuration: &hyperv1.ClusterConfiguration{Proxy: clusterWide},
					OperatorConfiguration: &hyperv1.OperatorConfiguration{
						Authentication: hyperv1.AuthenticationOperatorSpec{},
					},
				},
			},
			expected: clusterWide,
		},
		{
			name: "override set wins and is converted",
			hcp: &hyperv1.HostedControlPlane{
				Spec: hyperv1.HostedControlPlaneSpec{
					Configuration: &hyperv1.ClusterConfiguration{Proxy: clusterWide},
					OperatorConfiguration: &hyperv1.OperatorConfiguration{
						Authentication: hyperv1.AuthenticationOperatorSpec{Proxy: override},
					},
				},
			},
			expected: &configv1.ProxySpec{
				HTTPProxy:  "http://auth:8080",
				HTTPSProxy: "https://auth:8080",
				NoProxy:    "kube-apiserver,idp.example.com",
				TrustedCA:  configv1.ConfigMapNameReference{Name: "auth-ca"},
			},
		},
		{
			name: "override set with no cluster-wide",
			hcp: &hyperv1.HostedControlPlane{
				Spec: hyperv1.HostedControlPlaneSpec{
					OperatorConfiguration: &hyperv1.OperatorConfiguration{
						Authentication: hyperv1.AuthenticationOperatorSpec{Proxy: override},
					},
				},
			},
			expected: &configv1.ProxySpec{
				HTTPProxy:  "http://auth:8080",
				HTTPSProxy: "https://auth:8080",
				NoProxy:    "kube-apiserver,idp.example.com",
				TrustedCA:  configv1.ConfigMapNameReference{Name: "auth-ca"},
			},
		},
		{
			name: "override with only trustedCA falls back to cluster-wide",
			hcp: &hyperv1.HostedControlPlane{
				Spec: hyperv1.HostedControlPlaneSpec{
					Configuration: &hyperv1.ClusterConfiguration{Proxy: clusterWide},
					OperatorConfiguration: &hyperv1.OperatorConfiguration{
						Authentication: hyperv1.AuthenticationOperatorSpec{
							Proxy: &operatorv1.AuthenticationProxyConfig{
								TrustedCA: operatorv1.AuthenticationConfigMapReference{Name: "auth-ca"},
							},
						},
					},
				},
			},
			expected: clusterWide,
		},
		{
			name: "override with only NoProxy falls back to cluster-wide",
			hcp: &hyperv1.HostedControlPlane{
				Spec: hyperv1.HostedControlPlaneSpec{
					Configuration: &hyperv1.ClusterConfiguration{Proxy: clusterWide},
					OperatorConfiguration: &hyperv1.OperatorConfiguration{
						Authentication: hyperv1.AuthenticationOperatorSpec{
							Proxy: &operatorv1.AuthenticationProxyConfig{
								NoProxy: []string{"example.com"},
							},
						},
					},
				},
			},
			expected: clusterWide,
		},
		{
			name: "override with NoProxy and TrustedCA falls back to cluster-wide",
			hcp: &hyperv1.HostedControlPlane{
				Spec: hyperv1.HostedControlPlaneSpec{
					Configuration: &hyperv1.ClusterConfiguration{Proxy: clusterWide},
					OperatorConfiguration: &hyperv1.OperatorConfiguration{
						Authentication: hyperv1.AuthenticationOperatorSpec{
							Proxy: &operatorv1.AuthenticationProxyConfig{
								NoProxy: []string{"example.com"},
								TrustedCA: operatorv1.AuthenticationConfigMapReference{
									Name: "auth-ca",
								},
							},
						},
					},
				},
			},
			expected: clusterWide,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(EffectiveAuthProxy(tc.hcp)).To(Equal(tc.expected))
		})
	}
}
