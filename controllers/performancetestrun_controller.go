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
	"fmt"

	"github.com/go-logr/logr"
	specsv1alpha1 "github.com/performancetestinterface/performance-testing-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PerformanceTestRunReconciler reconciles a PerformanceTestRun object
type PerformanceTestRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=specs.pti-spec.io,resources=performancetestruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=specs.pti-spec.io,resources=performancetestruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=specs.pti-spec.io,resources=performancetestruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PerformanceTestRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *PerformanceTestRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("performanceTestRun", req.NamespacedName)
	// log := r.Log.WithValues("mykind", req.NamespacedName)
	// logger
	// // your logic here
	// logger.Info("fetching MyKind resource")
	performanceTestRun := specsv1alpha1.PerformanceTestRun{}
	if err := r.Client.Get(ctx, req.NamespacedName, &performanceTestRun); err != nil {
		// log.Error(err, "failed to get MyKind resource")
		// Ignore NotFound errors as they will be retried automatically if the
		// resource is created in future.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("Getting Performance Test " + performanceTestRun.Spec.PerformanceTestName)
	performanceTest := specsv1alpha1.PerformanceTest{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: "default",
		Name:      performanceTestRun.Spec.PerformanceTestName,
	}, &performanceTest); err != nil {

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("Getting Runner " + performanceTest.Spec.PerformanceTestRunner.Name)
	runner := specsv1alpha1.Runner{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: "default",
		Name:      performanceTest.Spec.PerformanceTestRunner.Name,
	}, &runner); err != nil {

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("Found runner with image " + runner.Spec.Image)
	newJob, err := CreateJob(logger, performanceTestRun.ObjectMeta.Name, performanceTestRun.ObjectMeta.Namespace, performanceTest, runner)
	if err != nil {
		logger.Error(err, "failed to get Create Job")
		return ctrl.Result{}, err
	}

	err = r.Client.Create(ctx, &newJob)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PerformanceTestRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&specsv1alpha1.PerformanceTestRun{}).
		Complete(r)
}

func CreateJob(logger logr.Logger, runName string, namespace string, test specsv1alpha1.PerformanceTest, runner specsv1alpha1.Runner) (batchv1.Job, error) {
	jobName := "job-" + runName
	logger.Info("Creating Job " + jobName)

	containerArgs := []string{"-u " + fmt.Sprintf("%v", test.Spec.PerformanceTestRunner.Users), "-i " + fmt.Sprintf("%v", test.Spec.PerformanceTestRunner.TotalIterations), "--rps " + fmt.Sprintf("%v", test.Spec.PerformanceTestRunner.QueriesPerSecondLimit)}
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: runName,
			Namespace:    namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: runName,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "runner",
							Image: runner.Spec.Image,
							Args:  containerArgs,
						},
					},
					RestartPolicy: v1.RestartPolicyOnFailure,
				},
			},
		},
	}
	return job, nil
}
