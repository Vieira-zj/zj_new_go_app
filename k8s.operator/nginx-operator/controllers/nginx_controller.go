/*
Copyright 2021 ZhengJin.

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
	"fmt"
	"reflect"
	"time"

	"github.com/example/nginx-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NginxReconciler reconciles a Nginx object
type NginxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=proxy.example.com,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=proxy.example.com,resources=nginxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=proxy.example.com,resources=nginxes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;delete;get;list;update;patch;watch
//+kubebuilder:rbac:groups="",resources=deployments;services,verbs="*"

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nginx object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Nginx")

	instance := &v1alpha1.Nginx{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil && errors.IsNotFound(err) {
		logger.Info("Nginx resource not found")
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Nginx")
		return ctrl.Result{}, err
	}
	if instance.DeletionTimestamp != nil {
		return ctrl.Result{}, fmt.Errorf("CR Nginx not found")
	}

	// 如果不存在，则创建关联资源
	deploy := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deploy); err != nil && errors.IsNotFound(err) {
		logger.Info("Create CR Nginx")
		// 1. 创建 Deploy
		deploy = NewDeploy(instance)
		if err := r.Create(ctx, deploy); err != nil {
			logger.Error(err, "Failed to create Deployment")
			return ctrl.Result{}, err
		}
		// 2. 创建 Service
		service := NewService(instance)
		if err := r.Create(ctx, service); err != nil {
			logger.Error(err, "Failed to create Service")
			return ctrl.Result{}, err
		}
		// 3. 关联 Annotations
		if err := updateNginxSpec(instance); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Update(ctx, instance); err != nil {
			logger.Error(err, "Failed to update Nginx instance")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// 如果存在，判断是否需要更新
	oldspec := v1alpha1.NginxSpec{}
	if err := json.Unmarshal([]byte(instance.Annotations["spec"]), &oldspec); err != nil {
		return ctrl.Result{}, err
	}

	if !reflect.DeepEqual(instance.Spec, oldspec) {
		// 更新关联资源
		logger.Info("Update CR Nginx")
		// 1. 更新 Deploy
		newDeploy := NewDeploy(instance)
		deploy.Spec = newDeploy.Spec
		if err := r.Update(ctx, deploy); err != nil {
			return ctrl.Result{}, err
		}

		// 2. 更新 Service
		// use Delete and Create service instead of r.Update.
		// Error: spec.clusterIP: Service "nginx-app" is invalid: spec.clusterIP: Invalid value: "": field is immutable
		oldService := &corev1.Service{}
		if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, oldService); err != nil {
			logger.Error(err, "Failed to get Service")
			return ctrl.Result{}, err
		}
		if err := r.Delete(ctx, oldService); err != nil {
			logger.Error(err, "Failed to delete Service")
			return ctrl.Result{}, err
		}
		newService := NewService(instance)
		if err := r.Create(ctx, newService); err != nil {
			logger.Error(err, "Failed to create Service")
			return ctrl.Result{}, err
		}

		// 3. 更新 Annotations
		if err := updateNginxSpec(instance); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Update(ctx, instance); err != nil {
			logger.Error(err, "Failed to update Nginx instance")
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
	}

	// 更新 status
	if instance.Status.DeploymentStatus.Replicas != deploy.Status.Replicas {
		logger.Info("Update CR Nginx Status")
		instance.Status.DeploymentStatus = deploy.Status
		if err := r.Status().Update(ctx, instance); err != nil {
			logger.Error(err, "Failed to update Nginx status")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Nginx{}).
		Complete(r)
}

func updateNginxSpec(instance *v1alpha1.Nginx) error {
	data, err := json.Marshal(instance.Spec)
	if err != nil {
		return err
	}
	if instance.Annotations != nil {
		instance.Annotations["spec"] = string(data)
	} else {
		instance.Annotations = map[string]string{"spec": string(data)}
	}
	return nil
}
