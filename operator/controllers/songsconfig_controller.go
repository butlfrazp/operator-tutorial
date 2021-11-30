/*
Copyright 2021.

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
	"encoding/json"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	songsv1beta1 "github.com/butlfrazp/operator-tutorial/operator/api/v1beta1"
)

// SongsConfigReconciler reconciles a SongsConfig object
type SongsConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=songs.example.com,resources=songsconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=songs.example.com,resources=songsconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=songs.example.com,resources=songsconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SongsConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *SongsConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Memcached instance
	songsConfigs := &songsv1beta1.SongsConfig{}
	err := r.Get(ctx, req.NamespacedName, songsConfigs)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("SongConfig resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get SongConfig")
		return ctrl.Result{}, err
	}

	// checking if the deployment exists
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: songsConfigs.Name, Namespace: songsConfigs.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			dep, err := r.deploymentForSongService(songsConfigs)
			if err != nil {
				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}

			log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.Create(ctx, dep)
			if err != nil {
				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	}

	// need to update the deplpyment if it differs
	songs := &songsConfigs.Spec.Songs
	b := []byte(found.Spec.Template.Spec.Containers[0].Env[0].Value)
	deployedSongs := &[]songsv1beta1.Song{}

	err = json.Unmarshal(b, deployedSongs)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !reflect.DeepEqual(songs, deployedSongs) {
		log.Info("Updating the pod")
		bytes, err := json.Marshal(songs)
		if err != nil {
			return ctrl.Result{}, err
		}
		found.Spec.Template.Spec.Containers[0].Env[0].Value = string(bytes)

		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed updating Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	return ctrl.Result{}, nil
}

func (r *SongsConfigReconciler) deploymentForSongService(s *songsv1beta1.SongsConfig) (*appsv1.Deployment, error) {
	labels := labelsForSongsConfig(s.Name)

	bytes, err := json.Marshal(s.Spec.Songs)
	if err != nil {
		return &appsv1.Deployment{}, err
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "acr14496.azurecr.io/songs-service:latest",
						Name:  s.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
						}},
						Env: []corev1.EnvVar{{
							Name:  "SONG_DATA",
							Value: string(bytes),
						}},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(s, dep, r.Scheme)
	return dep, nil
}

func labelsForSongsConfig(name string) map[string]string {
	return map[string]string{"app": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SongsConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&songsv1beta1.SongsConfig{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
