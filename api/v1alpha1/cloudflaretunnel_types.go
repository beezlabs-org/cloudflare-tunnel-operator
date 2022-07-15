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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudflareTunnelSpec defines the desired state of CloudflareTunnel
type CloudflareTunnelSpec struct {
	// +kubebuilder:validation:Format="url"
	Domain  string                   `json:"domain"`
	Zone    string                   `json:"zone"`
	Service *CloudflareTunnelService `json:"service"`
	// +kubebuilder:validation:Optional
	Container       *CloudflareTunnelContainer `json:"container"`
	TokenSecretName string                     `json:"tokenSecretName"`
	Replicas        int32                      `json:"replicas"`
}

type CloudflareTunnelService struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Enum=http;https
	Protocol string `json:"protocol"`
	Port     int32  `json:"port"`
}

type CloudflareTunnelContainer struct {
	// +kubebuilder:validation:Optional
	Image string `json:"image"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=IfNotPresent;Always;Never
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
	// +kubebuilder:validation:Optional
	Command []string `json:"command"`
	// +kubebuilder:validation:Optional
	Args []string `json:"args"`
}

// CloudflareTunnelStatus defines the observed state of CloudflareTunnel
type CloudflareTunnelStatus struct {
	// +kubebuilder:validation:Format="uuid"
	TunnelID    string                        `json:"tunnelID,omitempty"`
	Connections []CloudflareTunnelConnections `json:"connections"`
}

type CloudflareTunnelConnections struct {
	ConnectorID  string      `json:"connectorID,omitempty"`
	Created      metav1.Time `json:"created,omitempty"`
	Architecture string      `json:"architecture,omitempty"`
	Version      string      `json:"version,omitempty"`
	OriginIP     string      `json:"originIP,omitempty"`
	Edge         string      `json:"edge,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudflareTunnel is the Schema for the cloudflaretunnels API
type CloudflareTunnel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudflareTunnelSpec   `json:"spec,omitempty"`
	Status CloudflareTunnelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudflareTunnelList contains a list of CloudflareTunnel
type CloudflareTunnelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflareTunnel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareTunnel{}, &CloudflareTunnelList{})
}
