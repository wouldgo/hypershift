package extoidc

import (
	"strings"

	"github.com/openshift/hypershift/support/config"
	component "github.com/openshift/hypershift/support/controlplane-component"
	"github.com/openshift/hypershift/support/podspec"
	"github.com/openshift/hypershift/support/proxy"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func adaptDeployment(cpContext component.WorkloadContext, deployment *appsv1.Deployment) error {
	profile := cpContext.HCP.Spec.Configuration.GetTLSSecurityProfile()
	tlsMinVersion, err := config.MinTLSVersion(profile)
	if err != nil {
		return err
	}
	cipherSuites, err := config.CipherSuites(profile)
	if err != nil {
		return err
	}

	// NO_PROXY keeps the in-cluster TokenReview egress direct. kube-apiserver is
	// reached in-cluster and must bypass the konnectivity HTTPS proxy sidecar.
	noProxy := strings.Join([]string{proxy.DefaultNoProxy(cpContext.HCP), "kube-apiserver"}, ",")

	podspec.UpdateContainer(ComponentName, deployment.Spec.Template.Spec.Containers, func(c *corev1.Container) {
		if tlsMinVersion != "" {
			c.Args = append(c.Args, "--tls-min-version="+tlsMinVersion)
		}
		for _, cipherSuite := range cipherSuites {
			c.Args = append(c.Args, "--tls-cipher-suites="+cipherSuite)
		}

		podspec.UpsertEnvVar(c, corev1.EnvVar{
			Name:  "NO_PROXY",
			Value: noProxy,
		})
	})

	return nil
}
