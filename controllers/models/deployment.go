package models

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
)

func Deployment(name string, namespace string, replicas int32, id string, secret *corev1.Secret, tunnelURL string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + constants.ResourceSuffix,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/component":  "controller",
				"app.kubernetes.io/created-by": constants.OperatorName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "cloudflared",
							Image:   "ghcr.io/maggie0002/cloudflared:latest",
							Command: []string{"./cloudflared"},
							Args:    []string{"tunnel", "run", id},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "TUNNEL_URL",
									Value: tunnelURL,
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								corev1.EnvFromSource{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: secret.Name,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
