package console

import (
	operatorv1 "github.com/openshift/api/operator/v1"
)

// ReconcileConsoleAuthProxy sets the console-scoped auth proxy on the guest
// Console operator CR from the HostedControlPlane auth-scoped authWideproxyConfig.
//
// The value applies ONLY to console-operator's OIDC login clients (discovery,
// JWKS retrieval, code exchange, token refresh); no other guest component reads
// spec.authProxy. guestTrustedCAName is the name of the CA bundle ConfigMap
// already copied into the guest openshift-config namespace, or empty when the
// authWideproxyConfig carries no trustedCA.
//
// The authWideproxyConfig's NoProxy is a []string and is copied as-is; it must NOT be
// joined into a comma-separated string (that is the shape expected by
// configv1.ProxySpec, not by operatorv1.ConsoleAuthProxyConfig).
func ReconcileConsoleAuthProxy(console *operatorv1.Console, authWideproxyConfig *operatorv1.AuthenticationProxyConfig, guestTrustedCAName string) {
	noProxySlice := make([]string, len(authWideproxyConfig.NoProxy))
	noProxySlice = append(noProxySlice, authWideproxyConfig.NoProxy...)

	console.Spec.AuthProxy = operatorv1.ConsoleAuthProxyConfig{
		HTTPProxy:  authWideproxyConfig.HTTPProxy,
		HTTPSProxy: authWideproxyConfig.HTTPSProxy,
		NoProxy:    noProxySlice,
	}
	if guestTrustedCAName != "" {
		console.Spec.AuthProxy.TrustedCA = operatorv1.ConsoleAuthProxyTrustedCAConfigMapReference{
			Name: guestTrustedCAName,
		}
	}
}

// ClearConsoleAuthProxy removes any previously-set console auth proxy so that
// console-operator falls back to the cluster-wide proxy.
func ClearConsoleAuthProxy(console *operatorv1.Console) {
	console.Spec.AuthProxy = operatorv1.ConsoleAuthProxyConfig{}
}
