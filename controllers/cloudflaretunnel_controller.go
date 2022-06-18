/*
Copyright 2022.

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

package controllers

import (
	"context"
	"crypto/rand"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/cloudflare/cloudflare-go"

	cloudflaretunneloperatorv1alpha1 "github.com/beezlabs-org/cloudflare-tunnel-operator/api/v1alpha1"
)

// CloudflareTunnelReconciler reconciles a CloudflareTunnel object
type CloudflareTunnelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cloudflare-tunnel-operator.beezlabs.app,resources=cloudflaretunnels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloudflare-tunnel-operator.beezlabs.app,resources=cloudflaretunnels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloudflare-tunnel-operator.beezlabs.app,resources=cloudflaretunnels/finalizers,verbs=update

func (r *CloudflareTunnelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lfc := log.FromContext(ctx)
	lfc.Info("Starting Reconciler...")

	var cloudflareTunnel cloudflaretunneloperatorv1alpha1.CloudflareTunnel
	if err := r.Get(ctx, req.NamespacedName, &cloudflareTunnel); err != nil {
		lfc.Error(err, "could not fetch CloudflareTunnel")
		return ctrl.Result{}, err
	}

	name := cloudflareTunnel.Name
	namespace := cloudflareTunnel.Namespace
	spec := cloudflareTunnel.Spec

	secretName := spec.CredentialSecretName

	if len(secretName) == 0 {
		err := fmt.Errorf("CredentialSecretName key does not exist")
		lfc.Error(err, "CredentialSecretName not found")
		return ctrl.Result{}, err
	}

	var secret v1.Secret
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, &secret); err != nil {
		if errors.IsNotFound(err) {
			lfc.Error(err, "could not find secret with name", secretName)
		}
		return ctrl.Result{}, err
	}

	encodedCredentials, ok := secret.Data["credentials"]

	if !ok {
		err := fmt.Errorf("invalid key")
		lfc.Error(err, "key credentials not found")
		return ctrl.Result{}, err
	}

	credentials := string(encodedCredentials)

	cf, err := cloudflare.NewWithAPIToken(credentials)
	if err != nil {
		lfc.Error(err, "could not create cloudflare instance")
	}

	tunnelSecret, err := generateTunnelSecret()
	if err != nil {
		lfc.Error(err, "could not generate tunnel secret")
	}

	tunnelParams := cloudflare.TunnelCreateParams{
		AccountID: cf.AccountID,
		Name:      name,
		Secret:    tunnelSecret,
	}

	falsePointer := false

	tunnels, err := cf.Tunnels(ctx, cloudflare.TunnelListParams{AccountID: cf.AccountID, Name: name, IsDeleted: &falsePointer})
	if err != nil {
		lfc.Error(err, "could not fetch tunnel list")
	}

	var tunnel cloudflare.Tunnel

	if len(tunnels) >= 2 {
		lfc.Error(err, "2 or more tunnels already exists with the given name. Unable to choose between one of them")
	} else if len(tunnels) == 1 {
		lfc.Info("Tunnel already exists. Reconciling...")
		tunnel = tunnels[0]
	} else {
		lfc.Info("Tunnel doesn't exist. Creating...")
		tunnel, err = cf.CreateTunnel(ctx, tunnelParams)
		if err != nil {
			lfc.Error(err, "could not create the tunnel")
		}
	}

	var pod v1.Pod
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, pod); err != nil {
		if errors.IsNotFound(err) {
			lfc.Info("creating pod...")
			err = r.Create(ctx, pod)
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareTunnelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudflaretunneloperatorv1alpha1.CloudflareTunnel{}).
		Complete(r)
}

func generateTunnelSecret() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	return string(randomBytes), err
}
