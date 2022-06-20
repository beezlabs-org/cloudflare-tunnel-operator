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
	"encoding/base64"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudflaretunneloperatorv1alpha1 "github.com/beezlabs-org/cloudflare-tunnel-operator/api/v1alpha1"
	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/models"
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
	lfc.Info("Reconciling...")

	var cloudflareTunnel cloudflaretunneloperatorv1alpha1.CloudflareTunnel
	if err := r.Get(ctx, req.NamespacedName, &cloudflareTunnel); err != nil {
		lfc.Error(err, "could not fetch CloudflareTunnel")
		return ctrl.Result{}, err
	}
	lfc.V(1).Info("Resource fetched")

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

	var secret corev1.Secret
	// try to get a secret with the given name
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, &secret); err != nil {
		if errors.IsNotFound(err) {
			// write a log only if the secret was not found and not for other errors
			lfc.Error(err, "could not find secret with name "+secretName)
		}
		return ctrl.Result{}, err
	}
	lfc.V(1).Info("Secret fetched")

	// secret found, decode the token
	encodedCredentials, okCred := secret.Data["credentials"]
	encodedAccountID, okAccount := secret.Data["accountID"]

	if !okCred {
		err := fmt.Errorf("invalid key")
		lfc.Error(err, "key credentials not found")
		return ctrl.Result{}, err
	}

	if !okAccount {
		err := fmt.Errorf("invalid key")
		lfc.Error(err, "key accountID not found")
		return ctrl.Result{}, err
	}
	lfc.V(1).Info("Secret decoded")

	cf, err := cloudflare.NewWithAPIToken(string(encodedCredentials)) // create new instance of cloudflare sdk
	if err != nil {
		lfc.Error(err, "could not create cloudflare instance")
		return ctrl.Result{}, err
	}
	lfc.V(1).Info("Cloudflare instance successfully created")

	cf.AccountID = string(encodedAccountID)

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
	lfc.V(1).Info("Existing tunnels fetched")

	var tunnel cloudflare.Tunnel

	if len(tunnels) >= 2 {
		lfc.Info("Multiple tunnels exists with same name. Attempting to match ID")
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
		tunnelSecret, err := generateTunnelSecret() // generate a random secret to be used as the tunnel secret
		if err != nil {
			lfc.Error(err, "could not generate tunnel secret")
			return ctrl.Result{}, err
		}
		lfc.V(1).Info("Cloudflare Tunnel secret generated")

		tunnelParams := cloudflare.TunnelCreateParams{
			AccountID: cf.AccountID, // account is available after the sdk authenticates with the given secret
			Name:      name,         // name of the tunnel is the same as the name of the CRD
			Secret:    tunnelSecret, // use the randomly generated secret
		}

		tunnel, err = cf.CreateTunnel(ctx, tunnelParams)
		if err != nil {
			lfc.Error(err, "could not create the tunnel")
			return ctrl.Result{}, err
		}
	}

	tunnelToken, err := cf.TunnelToken(ctx, cloudflare.TunnelTokenParams{
		AccountID: cf.AccountID,
		ID:        tunnel.ID,
	})
	if err != nil {
		lfc.Error(err, "could not fetch tunnel token")
		return ctrl.Result{}, err
	}

	// this concludes checking the remote tunnel config
	// now first we create the secret containing the creds to the tunnel

	var secretFetch corev1.Secret
	secretCreate := models.Secret(name, namespace, tunnelToken)
	// the secret needs to have an owner reference back to the controller
	if err := ctrl.SetControllerReference(&cloudflareTunnel, secretCreate, r.Scheme); err != nil {
		lfc.Error(err, "could not create controller reference in secret")
		return ctrl.Result{}, err
	}
	lfc.V(1).Info("Owner Reference for Secret created")

	// try to get an existing secret with the given name
	if err := r.Get(ctx, types.NamespacedName{Name: secretCreate.Name, Namespace: namespace}, &secretFetch); err != nil {
		if errors.IsNotFound(err) {
			// error due to secret not being present, so, create one
			lfc.Info("creating secret...")
			if err := r.Create(ctx, secretCreate); err != nil {
				lfc.Error(err, "could not create secret")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	} else {
		// secret exists, so update it to ensure it is consistent
		if err := r.Update(ctx, secretCreate); err != nil {
			lfc.Error(err, "could not update secret")
			return ctrl.Result{}, err
		}
	}

	// now we have to check the deployment status and reconcile

	var deploymentFetch appsv1.Deployment
	deploymentCreate := models.Deployment(name, namespace, replicas, tunnel.ID, secretCreate, cloudflareTunnel.Spec.URL)
	// the deployment needs to have an owner reference back to the controller
	if err := ctrl.SetControllerReference(&cloudflareTunnel, deploymentCreate, r.Scheme); err != nil {
		lfc.Error(err, "could not create controller reference in deployment")
		return ctrl.Result{}, err
	}
	lfc.V(1).Info("Owner Reference for Deployment created")

	// try to get an existing deployment with the given name
	if err := r.Get(ctx, types.NamespacedName{Name: deploymentCreate.Name, Namespace: namespace}, &deploymentFetch); err != nil {
		if errors.IsNotFound(err) {
			// error due to deployment not being present, so, create one
			lfc.Info("creating deployment...")
			if err := r.Create(ctx, deploymentCreate); err != nil {
				lfc.Error(err, "could not create deployment")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	} else {
		// deployment exists, so update it to ensure it is consistent
		if err := r.Update(ctx, deploymentCreate); err != nil {
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
		//Owns(&appsv1.Deployment{}).
		Complete(r)
}

func generateTunnelSecret() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	return base64.StdEncoding.EncodeToString(randomBytes), err
}
