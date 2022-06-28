package models

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
)

type DeploymentModel struct {
	Name      string
	Namespace string
	Replicas  int32
	Secret    *corev1.Secret
	ConfigMap *corev1.ConfigMap
}

func Deployment(model DeploymentModel) *DeploymentModel {
	return &model
}

func (d *DeploymentModel) GetDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name + "-" + constants.ResourceSuffix,
			Namespace: d.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       d.Name,
				"app.kubernetes.io/component":  "controller",
				"app.kubernetes.io/created-by": constants.OperatorName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &d.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": d.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": d.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "cloudflared",
							Image:   "ghcr.io/maggie0002/cloudflared:latest",
							Command: []string{"./cloudflared"},
							Args:    []string{"tunnel", "run"},
						},
					},
				},
			},
		},
	}
}
