package extoidc

import (
	"encoding/json"
	"fmt"

	certs "github.com/openshift/hypershift/support/certs"
	component "github.com/openshift/hypershift/support/controlplane-component"
	externaloidc "github.com/openshift/library-go/pkg/operator/externaloidc"

	corev1 "k8s.io/api/core/v1"
)

const (
	authConfigDataKey = "auth-config.json"
)

func adaptAuthConfig(cpContext component.WorkloadContext, authCfgMap *corev1.ConfigMap) error {
	configuration := cpContext.HCP.Spec.Configuration
	if configuration == nil || configuration.Authentication == nil {
		return nil
	}

	gen := externaloidc.NewAuthenticationConfigurationGenerator(
		certs.ConfigMapCABundleResolver(cpContext, cpContext.Client, cpContext.HCP.Namespace),
	)

	authConfig, err := gen.Generate(configuration.Authentication)
	if err != nil {
		return fmt.Errorf("failed to generate authentication config: %w", err)
	}

	serializedConfig, err := json.Marshal(authConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize external-oidc-webhook authentication config: %w", err)
	}

	if authCfgMap.Data == nil {
		authCfgMap.Data = map[string]string{}
	}
	authCfgMap.Data[authConfigDataKey] = string(serializedConfig)

	return nil
}
