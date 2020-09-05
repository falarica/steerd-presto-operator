package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var log = logf.Log.WithName("webhooklog")


// +kubebuilder:webhook:verbs=create;update,path=/validate-falarica-v1alpha1-presto,mutating=false,failurePolicy=fail,groups=falarica.io,resources=prestos,versions=v1alpha1,name=validatorpresto.falarica.io

func (r *Presto) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
var _ webhook.Validator = &Presto{}

func (r *Presto) ValidateCreate() error {
	log.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Presto) ValidateUpdate(old runtime.Object) error {
	log.Info("validate update", "name", r.Name)

	return r.validatePrestoUpdate(old)
}

func (r *Presto) validatePrestoUpdate(old runtime.Object) error {
	errs := r.validatePrestoSpec(old.(*Presto))
	if len(errs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "falarica.io", Kind: "Presto"},
		r.Name, errs)
}

func (r *Presto) validatePrestoSpec(old *Presto) field.ErrorList {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	var allErrs field.ErrorList
	if (old.Spec.Coordinator.CpuRequest != r.Spec.Coordinator.CpuRequest) {
		err := &field.Error{"FieldImmutable", "Field is Immutable",
			"Spec.Coordinator.CpuRequest", "Field Spec.Coordinator.CpuRequest is Immutable"}
		allErrs = append(allErrs, err)
	}
	return allErrs
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Presto) ValidateDelete() error {
	return nil
}


