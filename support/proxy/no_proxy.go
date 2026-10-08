package proxy

// Based on https://github.com/openshift/cluster-network-operator/blob/4b792c659385948e825d2ba17b8f6d2e5c3acfed/pkg/util/proxyconfig/no_proxy.go

import (
	"fmt"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	hyperv1 "github.com/openshift/hypershift/api/hypershift/v1beta1"

	"k8s.io/apimachinery/pkg/util/sets"
)

const defaultCIDR = "0.0.0.0/0"

// DefaultNoProxy builds the no-proxy list from the cluster-internal defaults and
// the cluster-wide proxy (spec.configuration.proxy) NoProxy entries.
func DefaultNoProxy(hcp *hyperv1.HostedControlPlane) string {
	var proxyConfig *configv1.ProxySpec
	if hcp.Spec.Configuration != nil {
		proxyConfig = hcp.Spec.Configuration.Proxy
	}
	return DefaultNoProxyForProxy(hcp, proxyConfig)
}

// DefaultNoProxyForProxy builds the no-proxy list from the cluster-internal
// defaults and the supplied proxy's NoProxy entries. A component that overrides
// the cluster-wide proxy passes its effective proxy here so the override's
// NoProxy replaces the cluster-wide one entirely.
func DefaultNoProxyForProxy(hcp *hyperv1.HostedControlPlane, proxyConfig *configv1.ProxySpec) string {
	set := sets.New(
		"127.0.0.1",
		"localhost",
		".svc",
		".cluster.local",
	)
	set.Insert(".hypershift.local")

	for _, mc := range hcp.Spec.Networking.MachineNetwork {
		if mc.CIDR.String() == defaultCIDR {
			continue
		}
		set.Insert(mc.CIDR.String())
	}

	for _, nss := range hcp.Spec.Networking.ServiceNetwork {
		set.Insert(nss.CIDR.String())
	}

	switch hcp.Spec.Platform.Type {
	case hyperv1.AWSPlatform, hyperv1.AzurePlatform:
		set.Insert("169.254.169.254")
	}

	// Construct the node sub domain.
	// TODO: Add support for additional cloud providers.
	switch hcp.Spec.Platform.Type {
	case hyperv1.AWSPlatform:
		region := hcp.Spec.Platform.AWS.Region
		if region == "us-east-1" {
			set.Insert(".ec2.internal")
		} else {
			set.Insert(fmt.Sprintf(".%s.compute.internal", region))
		}
	case hyperv1.AzurePlatform:
		if cloudName := hcp.Spec.Platform.Azure.Cloud; cloudName != "AzurePublicCloud" {
			// https://learn.microsoft.com/en-us/azure/virtual-network/what-is-ip-address-168-63-129-16
			set.Insert("168.63.129.16")
			// https://bugzilla.redhat.com/show_bug.cgi?id=2104997
			// TODO (cewong): determine where the ARMEndpoint is calculated
			// if cloudName == "AzureStackCloud" {
			// set.Insert(infra.Status.PlatformStatus.Azure.ARMEndpoint)
			// }
		}
	}

	for _, clusterNetwork := range hcp.Spec.Networking.ClusterNetwork {
		set.Insert(clusterNetwork.CIDR.String())
	}

	if proxyConfig != nil && len(proxyConfig.NoProxy) > 0 {
		for userValue := range strings.SplitSeq(proxyConfig.NoProxy, ",") {
			if userValue != "" {
				set.Insert(userValue)
			}
		}
	}

	return strings.Join(sets.List(set), ",")

}
