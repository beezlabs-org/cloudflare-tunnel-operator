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

	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	// get data from the CRD
	name := cloudflareTunnel.Name
	namespace := cloudflareTunnel.Namespace
	spec := cloudflareTunnel.Spec
	secretName := spec.CredentialSecretName
	replicas := spec.Replicas

	// check if a secret name is mentioned or not
	if len(secretName) == 0 {
		err := fmt.Errorf("CredentialSecretName key does not exist")
		lfc.Error(err, "CredentialSecretName not found")
		return ctrl.Result{}, err
	}

	var secret v1.Secret
	// try to get a secret with the given name
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, &secret); err != nil {
		if errors.IsNotFound(err) {
			// write a log only if the secret was not found and not for other errors
			lfc.Error(err, "could not find secret with name", secretName)
		}
		return ctrl.Result{}, err
	}

	// secret found, decode the token
	encodedCredentials, ok := secret.Data["credentials"]

	if !ok {
		err := fmt.Errorf("invalid key")
		lfc.Error(err, "key credentials not found")
		return ctrl.Result{}, err
	}

	credentials := string(encodedCredentials)

	cf, err := cloudflare.NewWithAPIToken(credentials) // create new instance of cloudflare sdk
	if err != nil {
		lfc.Error(err, "could not create cloudflare instance")
		return ctrl.Result{}, err
	}

	tunnelSecret, err := generateTunnelSecret() // generate a random secret to be used as the tunnel secret
	if err != nil {
		lfc.Error(err, "could not generate tunnel secret")
		return ctrl.Result{}, err
	}

	tunnelParams := cloudflare.TunnelCreateParams{
		AccountID: cf.AccountID, // account is available after the sdk authenticates with the given secret
		Name:      name,         // name of the tunnel is the same as the name of the CRD
		Secret:    tunnelSecret, // use the randomly generated secret
	}

	falsePointer := false // needed as the function below only accepts a *bool

	// first, we are checking if tunnels with the given name exists in the remote or not
	// if they exist, we will be getting one or more of them, since cloudflare allows duplicate named tunnels
	// if 2 or more exists, we check if the current CRD status already has the ConnectorID or not
	// if it has, we check if the returned tunnels has one with the same connector id and use it
	// else, we cannot accurately figure out which one of them to use and error out
	tunnels, err := cf.Tunnels(ctx, cloudflare.TunnelListParams{
		AccountID: cf.AccountID,
		Name:      name,
		IsDeleted: &falsePointer,
	})
	if err != nil {
		lfc.Error(err, "could not fetch tunnel list")
		return ctrl.Result{}, err
	}

	var tunnel cloudflare.Tunnel

	if len(tunnels) >= 2 {
		lfc.Info("Multiple tunnels  exists with same name. Attempting to match ID")
		existingConnectorID := cloudflareTunnel.Status.ConnectorID
		// loop through all found tunnels until one is found with a matching ID
		for _, t := range tunnels {
			if t.ID == existingConnectorID {
				lfc.Info("Tunnel found with matching ID. Reconciling...")
				tunnel = t
				break // break loop once found
			}
		}
		// if no tunnel has been assigned that means a tunnel with a matching ID was not found
		if &tunnel == nil {
			// todo: create am error instance
			lfc.Error(err, "2 or more tunnels already exists with the given name. Unable to choose between one of them")
			return ctrl.Result{}, err
		}
	} else if len(tunnels) == 1 {
		// a single tunnel found with the same name, so we use that
		lfc.Info("Tunnel already exists. Reconciling...")
		tunnel = tunnels[0]
	} else {
		lfc.Info("Tunnel doesn't exist. Creating...")
		tunnel, err = cf.CreateTunnel(ctx, tunnelParams)
		if err != nil {
			lfc.Error(err, "could not create the tunnel")
			return ctrl.Result{}, err
		}
	}

	// this concludes checking the remote tunnel config
	// now we have to check the deployment status and reconcile

	deploymentName := "cf-tunnel-" + name // the deployment name has a prefix
	var deploymentFetch v12.Deployment
	deploymentCreate := v12.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       deploymentName,
				"app.kubernetes.io/component":  "controller",
				"app.kubernetes.io/created-by": "cloudflare-tunnel-operator",
			},
		},
		Spec: v12.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": deploymentName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": deploymentName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						v1.Container{
							Name:  "cloudflared",
							Image: "ghcr.io/maggie0002/cloudflared:latest",
						},
					},
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(&cloudflareTunnel, &deploymentCreate, r.Scheme); err != nil {
		lfc.Error(err, "could not create controller reference in deployment")
		return ctrl.Result{}, err
	}
	// try to get an existing deployment with the given name
	if err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: namespace}, &deploymentFetch); err != nil {
		if errors.IsNotFound(err) {
			// error due to deployment not being present, so, create one
			lfc.Info("creating deployment...")
			if err := r.Create(ctx, &deploymentCreate); err != nil {
				lfc.Error(err, "could not create deployment")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	} else {
		// deployment exists, so update it to ensure it is consistent
		if err := r.Update(ctx, &deploymentCreate); err != nil {
			lfc.Error(err, "could not update deployment")
			return ctrl.Result{}, err
		}
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
