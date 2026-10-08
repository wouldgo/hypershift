package proxy

import (
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	hyperv1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

// EffectiveAuthProxy returns the proxy configuration that the External OIDC
// authentication components must use for their egress.
//
// When spec.operatorConfiguration.authentication.proxy is set, it overrides the
// cluster-wide proxy entirely. There is no field merge. The override is a
// operatorv1.AuthenticationProxyConfig. It is converted to a configv1.ProxySpec
// so the rest of the codebase handles both sources with a single type.
//
// When the override is not set, the cluster-wide spec.configuration.proxy is
// returned. When neither is set, nil is returned.
func EffectiveAuthProxy(hcp *hyperv1.HostedControlPlane) *configv1.ProxySpec {
	if hcp == nil {
		return nil
	}

	if oc := hcp.Spec.OperatorConfiguration; oc != nil {
		if authProxy := oc.Authentication.Proxy; !isZeroAuthProxy(authProxy) {
			return authProxyToProxySpec(authProxy)
		}
	}

	if cfg := hcp.Spec.Configuration; cfg != nil {
		return cfg.Proxy
	}

	return nil
}

// isZeroAuthProxy reports whether the override carries no proxy configuration.
func isZeroAuthProxy(p *operatorv1.AuthenticationProxyConfig) bool {
	return p == nil || (p.HTTPProxy == "" && p.HTTPSProxy == "")
}

// authProxyToProxySpec converts the authentication-scoped override into the
// cluster-wide proxy type. NoProxy is a slice on the override and a
// comma-separated string on configv1.ProxySpec, so it is joined here.
// ReadinessEndpoints has no source on the override and stays empty.
func authProxyToProxySpec(p *operatorv1.AuthenticationProxyConfig) *configv1.ProxySpec {
	return &configv1.ProxySpec{
		HTTPProxy:  p.HTTPProxy,
		HTTPSProxy: p.HTTPSProxy,
		NoProxy:    strings.Join(p.NoProxy, ","),
		TrustedCA:  configv1.ConfigMapNameReference{Name: p.TrustedCA.Name},
	}
}
