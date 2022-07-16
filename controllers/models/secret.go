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
	"encoding/json"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/templates"
)

type SecretModel struct {
	Name         string
	Namespace    string
	TunnelToken  string
	AccountTag   string
	TunnelSecret string
	TunnelID     string
}

type tunnelToken struct {
	A string `json:"a"`
	S string `json:"s"`
	T string `json:"t"`
}

func Secret(model SecretModel) *SecretModel {
	return &model
}

func (s *SecretModel) GetSecret() (*corev1.Secret, error) {
	secret, err := s.generateSecret()
	if err != nil {
		return nil, err
	}
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
			s.TunnelID + ".json": secret,
		},
		Type: corev1.SecretTypeOpaque,
	}, nil
}

func (s *SecretModel) generateSecret() (string, error) {
	var tokenJson tunnelToken
	if err := json.Unmarshal([]byte(s.TunnelToken), &tokenJson); err != nil {
		return "", err
	}
	s.AccountTag = tokenJson.A
	s.TunnelSecret = tokenJson.S
	s.TunnelID = tokenJson.T
	templateEngine, err := template.New("secret").Parse(templates.SECRET)
	if err != nil {
		return "", err
	}

	var dataBuffer bytes.Buffer
	err = templateEngine.Execute(&dataBuffer, &s)
	if err != nil {
		return "", err
	}

	secret := dataBuffer.String()

	return secret, nil
}
