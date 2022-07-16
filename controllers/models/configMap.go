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
	"bytes"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/templates"
)

type ConfigMapModel struct {
	Name      string
	Namespace string
	Service   string
	TunnelID  string
	Domain    string
}

func ConfigMap(model ConfigMapModel) *ConfigMapModel {
	return &model
}

func (cm *ConfigMapModel) GetConfigMap() (*corev1.ConfigMap, error) {
	configMap, err := cm.generateConfigMap()
	if err != nil {
		return nil, err
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm.Name + "-" + constants.ResourceSuffix,
			Namespace: cm.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       cm.Name,
				"app.kubernetes.io/component":  "controller",
				"app.kubernetes.io/created-by": constants.OperatorName,
			},
		},
		Data: map[string]string{
			"config.yaml": configMap,
		},
	}, nil
}

func (cm *ConfigMapModel) generateConfigMap() (string, error) {
	templateEngine, err := template.New("config").Parse(templates.CONFIG)
	if err != nil {
		return "", err
	}

	var dataBuffer bytes.Buffer
	err = templateEngine.Execute(&dataBuffer, &cm)
	if err != nil {
		return "", err
	}

	secret := dataBuffer.String()

	return secret, nil
}
