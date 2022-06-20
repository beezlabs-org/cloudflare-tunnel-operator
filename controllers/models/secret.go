package models

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
)

func Secret(name string, namespace string, tunnelSecret string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + constants.ResourceSuffix,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/component":  "controller",
				"app.kubernetes.io/created-by": constants.OperatorName,
			},
		},
		StringData: map[string]string{
			"TUNNEL_TOKEN": tunnelSecret,
		},
		Type: corev1.SecretTypeOpaque,
	}
}
