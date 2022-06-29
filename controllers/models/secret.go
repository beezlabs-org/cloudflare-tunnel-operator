package models

import (
	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretModel struct {
	Name         string
	Namespace    string
	TunnelSecret string
	TunnelID     string
}

func Secret(model SecretModel) *SecretModel {
	return &model
}

func (s *SecretModel) GetSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name + "-" + constants.ResourceSuffix,
			Namespace: s.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       s.Name,
				"app.kubernetes.io/component":  "controller",
				"app.kubernetes.io/created-by": constants.OperatorName,
			},
		},
		StringData: map[string]string{
			s.TunnelID + ".json": s.TunnelSecret,
		},
		Type: corev1.SecretTypeOpaque,
	}
}
