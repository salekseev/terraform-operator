package controllers

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	backendv1alpha1 "github.com/krubot/terraform-operator/pkg/apis/backend/v1alpha1"
	modulev1alpha1 "github.com/krubot/terraform-operator/pkg/apis/module/v1alpha1"
	providerv1alpha1 "github.com/krubot/terraform-operator/pkg/apis/provider/v1alpha1"
	terraform "github.com/krubot/terraform-operator/pkg/terraform"
	util "github.com/krubot/terraform-operator/pkg/util"
)

// ReconcileProvider reconciles a Backend object
type ReconcileProvider struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ReconcileProvider) deletionReconcile(providerInterface interface{}, finalizerInterfaces ...interface{}) error {
	for _, finalizerInterface := range finalizerInterfaces {
		switch provider := providerInterface.(type) {
		case *providerv1alpha1.Google:
			for _, fin := range provider.GetFinalizers() {
				instance_split_fin := strings.Split(fin, "_")
				switch finalizer := finalizerInterface.(type) {
				case *modulev1alpha1.GCS:
					if instance_split_fin[0] == "Module" && instance_split_fin[1] == "GCS" {
						if err := r.Get(context.Background(), types.NamespacedName{Name: instance_split_fin[2], Namespace: provider.ObjectMeta.Namespace}, finalizer); errors.IsNotFound(err) {
							util.RemoveFinalizer(provider, fin)
							if err := r.Update(context.Background(), provider); err != nil {
								return err
							}
						} else {
							return errors.NewBadRequest("GCS dependency is not met for deletion")
						}
					}
				case *providerv1alpha1.Google:
					if instance_split_fin[0] == "Provider" && instance_split_fin[1] == "Google" && instance_split_fin[2] != provider.ObjectMeta.Name {
						if err := r.Get(context.Background(), types.NamespacedName{Name: instance_split_fin[2], Namespace: provider.ObjectMeta.Namespace}, finalizer); errors.IsNotFound(err) {
							util.RemoveFinalizer(provider, fin)
							if err := r.Update(context.Background(), provider); err != nil {
								return err
							}
						} else {
							return errors.NewBadRequest("Google dependency is not met for deletion")
						}
					}
				case *backendv1alpha1.EtcdV3:
					if instance_split_fin[0] == "Backend" && instance_split_fin[1] == "EtcdV3" {
						if err := r.Get(context.Background(), types.NamespacedName{Name: instance_split_fin[2], Namespace: provider.ObjectMeta.Namespace}, finalizer); errors.IsNotFound(err) {
							util.RemoveFinalizer(provider, fin)
							if err := r.Update(context.Background(), provider); err != nil {
								return err
							}
						} else {
							return errors.NewBadRequest("EtcdV3 dependency is not met for deletion")
						}
					}
				}
			}
		}
	}
	return nil
}

func (r *ReconcileProvider) dependencyReconcile(providerInterface interface{}, depInterfaces ...interface{}) (bool, error) {
	// Set the initial depency state
	dependency_met := true
	// Over the list of dependencies
	for _, depInterface := range depInterfaces {
		switch provider := providerInterface.(type) {
		case *providerv1alpha1.Google:
			for _, depProvider := range provider.Dep {
				switch dep := depInterface.(type) {
				case *modulev1alpha1.GCS:
					if depProvider.Kind == "Module" {
						if err := r.Get(context.Background(), types.NamespacedName{Name: depProvider.Name, Namespace: provider.ObjectMeta.Namespace}, dep); err != nil {
							return dependency_met, err
						}
						if dep.Status.State == "Success" {
							// Add finalizer to the GoogleStorageBucket. resource
							util.AddFinalizer(dep, "Provider_"+provider.Kind+"_"+provider.ObjectMeta.Name)
							// Update the CR with finalizer
							if err := r.Update(context.Background(), dep); err != nil {
								return dependency_met, err
							}
						} else {
							dependency_met = false
						}
					}
				case *providerv1alpha1.Google:
					if depProvider.Kind == "Provider" {
						if err := r.Get(context.Background(), types.NamespacedName{Name: depProvider.Name, Namespace: provider.ObjectMeta.Namespace}, dep); err != nil {
							return dependency_met, err
						}
						if dep.Status.State == "Success" {
							// Add finalizer to the GoogleStorageBucket resource
							util.AddFinalizer(dep, "Provider_"+provider.Kind+"_"+provider.ObjectMeta.Name)
							// Update the CR with finalizer
							if err := r.Update(context.Background(), dep); err != nil {
								return dependency_met, err
							}
						} else {
							dependency_met = false
						}
					}
				case *backendv1alpha1.EtcdV3:
					if depProvider.Kind == "Backend" {
						if err := r.Get(context.Background(), types.NamespacedName{Name: depProvider.Name, Namespace: provider.ObjectMeta.Namespace}, dep); err != nil {
							return dependency_met, err
						}
						if dep.Status.State == "Success" {
							// Add finalizer to the GoogleStorageBucket resource
							util.AddFinalizer(dep, "Provider_"+provider.Kind+"_"+provider.ObjectMeta.Name)
							// Update the CR with finalizer
							if err := r.Update(context.Background(), dep); err != nil {
								return dependency_met, err
							}
						} else {
							dependency_met = false
						}
					}
				}
			}
		}
	}
	return dependency_met, nil
}

// +kubebuilder:rbac:groups=batch.my.domain,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.my.domain,resources=cronjobs/status,verbs=get;update;patch

func (r *ReconcileProvider) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	module := &modulev1alpha1.GCS{}
	provider := &providerv1alpha1.Google{}
	backend := &backendv1alpha1.EtcdV3{}

	if err := r.Get(context.Background(), req.NamespacedName, provider); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	for {
		e := "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"
		h := map[string][]string{"Metadata-Flavor": {"Google"}}
		if ret := checkURL(e, h, 200); ret == nil {
			break
		}
	}

	if util.IsBeingDeleted(provider) {
		if err := r.deletionReconcile(provider, backend, module); err != nil {
			return reconcile.Result{}, err
		}

		if err := terraform.WriteToFile([]byte("{}"), provider.ObjectMeta.Namespace, "Provider_"+provider.Kind+"_"+provider.ObjectMeta.Name); err != nil {
			return reconcile.Result{}, err
		}

		util.RemoveFinalizer(provider, "Provider_"+provider.Kind+"_"+provider.ObjectMeta.Name)

		if err := r.Update(context.Background(), provider); err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	if dependency_met, err := r.dependencyReconcile(provider, module, backend); err == nil {
		// Check if dependency is met else interate again
		if !dependency_met {
			// Set the data
			provider.Status.State = "Failure"
			provider.Status.Phase = "Dependency"
			// Update the CR with status success
			if err := r.Status().Update(context.Background(), provider); err != nil {
				return reconcile.Result{}, err
			}
			// Dependency not met, don't error but finish reconcile until next change
			return reconcile.Result{}, nil
		}

		if !reflect.DeepEqual("Dependency", provider.Status.Phase) {
			// Set the data
			provider.Status.State = "Success"
			provider.Status.Phase = "Dependency"
			// Update the CR with status ready
			if err := r.Status().Update(context.Background(), provider); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		return reconcile.Result{}, err
	}

	// Add finalizer to the module resource
	util.AddFinalizer(provider, "Provider_"+provider.Kind+"_"+provider.ObjectMeta.Name)

	// Update the CR with finalizer
	if err := r.Update(context.Background(), provider); err != nil {
		return reconcile.Result{}, err
	}

	b, err := terraform.RenderProviderToTerraform(provider.Spec, strings.ToLower(provider.Kind))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = terraform.WriteToFile(b, provider.ObjectMeta.Namespace, "Provider_"+provider.Kind+"_"+provider.ObjectMeta.Name)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Update CR with the AppStatus == Created
	if !reflect.DeepEqual("Ready", provider.Status.State) {
		// Set the data
		provider.Status.State = "Success"
		provider.Status.Phase = "Output"

		// Update the CR
		if err = r.Status().Update(context.Background(), provider); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileProvider) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&providerv1alpha1.Google{}).
		Watches(&source.Kind{Type: &providerv1alpha1.Google{}},
			&handler.EnqueueRequestForObject{}).
		WithEventFilter(util.ResourceGenerationChangedPredicate{}).
		Complete(r)
}
