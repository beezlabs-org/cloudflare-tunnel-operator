package models

import (
	"bytes"
	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"text/template"

	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
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
