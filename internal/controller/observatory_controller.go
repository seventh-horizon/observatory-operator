/*
Copyright 2024 The Observatory Operator Authors.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	workflowv1alpha1 "github.com/seventh-horizon/observatory-operator/api/v1alpha1"
)

// ObservatoryReconciler reconciles a Observatory object
type ObservatoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=workflow.seventh-horizon.io,resources=observatories,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workflow.seventh-horizon.io,resources=observatories/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=workflow.seventh-horizon.io,resources=observatories/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Observatory object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ObservatoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Observatory instance
	var observatory workflowv1alpha1.Observatory
	if err := r.Get(ctx, req.NamespacedName, &observatory); err != nil {
		log.Error(err, "unable to fetch Observatory")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: Implement the reconciliation logic here
	// This is where you would:
	// 1. Parse the DAG structure from observatory.Spec.Tasks
	// 2. Create pods for tasks that are ready to run (dependencies satisfied)
	// 3. Monitor task execution and update status
	// 4. Handle failures and retries
	// 5. Update metrics and emit events

	log.Info("Reconciling Observatory", "name", observatory.Name, "namespace", observatory.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObservatoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workflowv1alpha1.Observatory{}).
		Complete(r)
}
