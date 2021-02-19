/*


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
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	migrationsv1alpha1 "flyway-operator/api/v1alpha1"
)

// MigrationReconciler reconciles a Migration object
type MigrationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=migrations.flywayoperator.io,resources=migrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=migrations.flywayoperator.io,resources=migrations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

func (r *MigrationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("migration", req.NamespacedName)

	var migration migrationsv1alpha1.Migration
	if err := r.Get(ctx, req.NamespacedName, &migration); err != nil {
		log.Info("unable to fetch migration spec " + req.NamespacedName.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if migration.ObjectMeta.DeletionTimestamp.IsZero() {
		// load db creds if provided through secret
		creds := GetCredentials(&migration)
		if creds == nil {
			return ctrl.Result{}, errors.New("unable to get db credentials for migration")
		}

		wait := waitForDB(&migration.Spec.DB, log)
		if <-wait == false {
			return ctrl.Result{}, errors.New("timeout reached after trying to connect to db")
		}
		sqlDriver := Drivers[migration.Spec.DB.Driver]
		job := batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("flyway-%s", req.NamespacedName.Name),
				Namespace: req.NamespacedName.Namespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						RestartPolicy: "Never",
						Containers: []corev1.Container{
							corev1.Container{
								Name:            "flyway-migration",
								Image:           "flyway/flyway",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Env: []corev1.EnvVar{
									corev1.EnvVar{Name: "FLYWAY_DRIVER", Value: migration.Spec.DB.Driver},
									corev1.EnvVar{Name: "FLYWAY_URL", Value: sqlDriver.ConnectionURL(&migration.Spec.DB)},
								},
								Args: []string{"migrate"},
								VolumeMounts: []corev1.VolumeMount{
									corev1.VolumeMount{Name: SQLVolumeName, MountPath: "/flyway/sql"},
								},
							},
						},
					},
				},
			},
		}

		// mutate template according to creds specs
		creds.MutateTemplate(&job.Spec.Template)
		location := GetScriptsLocation(&migration.Spec.SQL)
		if location == nil {
			return ctrl.Result{}, errors.New("unable to detect sql scripts location")
		}
		// mutate template according to sql scripts location
		location.MutateTemplate(&job.Spec.Template)

		if err := r.Client.Create(ctx, &job); err != nil {
			return ctrl.Result{}, err
		}

	} else {
		// TODO finalizer and job clean up
	}

	return ctrl.Result{}, nil
}

func waitForDB(spec *migrationsv1alpha1.DBSpec, log logr.Logger) chan bool {
	wait := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	go func() {
		select {
		case <-ctx.Done():
			log.Info("timeout (after 10 mns) while waiting for db access !")
			wait <- false
		}
	}()

	go func() {
		defer cancel()
		sqlDriver := Drivers[spec.Driver]
		for {
			_, err := sqlDriver.CheckDBAvailability(spec)
			if err != nil {
				time.Sleep(10 * time.Second)
				log.Info("waiting for database availability...")
			} else {
				wait <- true
				break
			}
		}
	}()

	return wait
}

func (r *MigrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&migrationsv1alpha1.Migration{}).
		Complete(r)
}
