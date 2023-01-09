/*
Copyright 2022 Beez Innovation Labs.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
)

type DeploymentModel struct {
	Name            string
	Namespace       string
	Replicas        int32
	TunnelID        string
	Image           string
	ImagePullPolicy corev1.PullPolicy
	Command         []string
	Args            []string
	Secret          *corev1.Secret
	ConfigMap       *corev1.ConfigMap
}

func Deployment(model DeploymentModel) *DeploymentModel {
	return &model
}

func (d *DeploymentModel) GetDeployment() *appsv1.Deployment {
	image := "cloudflare/cloudflared:latest"
	if d.Image != "" {
		image = d.Image
	}
	imagePullPolicy := corev1.PullAlways
	if d.ImagePullPolicy != "" {
		imagePullPolicy = d.ImagePullPolicy
	}
	command := []string{"cloudflared"}
	if len(d.Command) != 0 {
		command = d.Command
	}
	args := []string{"tunnel", "--metrics", "localhost:9090", "--no-autoupdate", "--config", "/config/config.yaml", "run"}
	if len(d.Args) != 0 {
		args = d.Args
	}
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
						{
							Name:            "cloudflared",
							Image:           image,
							ImagePullPolicy: imagePullPolicy,
							Command:         command,
							Args:            args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: 9090,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cloudflared-config",
									MountPath: "/config/config.yaml",
									SubPath:   "config.yaml",
								},
								{
									Name:      "cloudflared-creds",
									MountPath: "/config/" + d.TunnelID + ".json",
									SubPath:   d.TunnelID + ".json",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "cloudflared-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: d.Name + "-" + constants.ResourceSuffix},
								},
							},
						},
						{
							Name: "cloudflared-creds",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: d.Name + "-" + constants.ResourceSuffix,
								},
							},
						},
					},
				},
			},
		},
	}
}
