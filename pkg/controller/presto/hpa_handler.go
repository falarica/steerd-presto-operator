package presto


import (
	"context"
	"fmt"
	"github.com/falarica/steerd-presto-operator/pkg/apis/falarica/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
  There are certain global properties which we might want to change. for e.g
	--horizontal-pod-autoscaler-downscale-stabilization duration     Default: 5m0s
		The period for which autoscaler will look backwards and not scale down below
			any recommendation it made during that period.
  --horizontal-pod-autoscaler-sync-period duration     Default: 15s
		The period for syncing the number of pods in horizontal pod autoscaler.
  However, these cannot be changed for managed K8s services like GKE and EKS
  https://stackoverflow.com/questions/55815094/how-to-change-horizontal-pod-autoscaler-sync-period-field-in-kube-controller-m
  https://stackoverflow.com/questions/46317275/change-the-horizontal-pod-autoscaler-sync-period-with-gke
*/
func handleReplicaSet(
	r *ReconcilePresto,
	presto *v1alpha1.Presto,
	workerReplicaSet *v1.ReplicaSet,
	lbls map[string]string,
	ctx context.Context) (bool, bool, bool, error) {

	autoScalingEnabled := checkAutoscalingEnabled(presto)
	created := false
	updated := false
	deleted := false

	var hpa *autoscalingv1.HorizontalPodAutoscaler
	exists := true
	hpa, err := getPrestoHPA(r, presto, lbls)
	if errors.IsNotFound(err) {
		exists = false
	} else if err != nil {
		return created, updated, deleted, err
	}
	if exists {
		if autoScalingEnabled {
			if autoscaleSpecChanged(hpa, presto) {
				r.log.Info(fmt.Sprintf("HPA spec will be updated"))
				hpa, err := createHPASpec(presto, workerReplicaSet,
					lbls, hpa.ObjectMeta.ResourceVersion)
				if err != nil {
					return created, updated, deleted, err
				}
				err = r.client.Update(ctx, hpa)
				if err != nil && !errors.IsAlreadyExists(err) {
					r.log.Error(err, "Failed to update HPA spec")
					return created, updated, deleted, err
				}
				updated = true
			}
		} else {
			r.log.Info(fmt.Sprintf("Since autoscaling is disabled, deleting HPA spec"))

			err := r.client.Delete(ctx, hpa)
			if err != nil {
				r.log.Error(err, "Failed to delete HPA spec")
				return created, updated, deleted, err
			}
			deleted = true
		}
	} else if autoScalingEnabled {
		r.log.Info(fmt.Sprintf("Creating HPA Spec"))
		hpa, err := createHPASpec(presto, workerReplicaSet, lbls, "")
		if err != nil {
			return created, updated, deleted, err
		}
		err = r.client.Create(ctx, hpa)
		if err != nil && !errors.IsAlreadyExists(err) {
			r.log.Error(err, "Failed to create HPA Spec")
			return created, updated, deleted, err
		}
		created = true
	}
	return created, updated, deleted, nil
}

func autoscaleSpecChanged(hpa *autoscalingv1.HorizontalPodAutoscaler,
	presto *v1alpha1.Presto) bool {
	return hpa.Spec.MaxReplicas != *presto.Spec.Worker.Autoscaling.MaxReplicas ||
		*hpa.Spec.MinReplicas != *presto.Spec.Worker.Autoscaling.MinReplicas ||
		*hpa.Spec.TargetCPUUtilizationPercentage != *presto.Spec.Worker.Autoscaling.TargetCPUUtilizationPercentage
}

func getPrestoHPA(r *ReconcilePresto,presto *v1alpha1.Presto,
	lbls map[string]string) (*autoscalingv1.HorizontalPodAutoscaler, error) {
	prestoHPA := &autoscalingv1.HorizontalPodAutoscalerList{}
	err := r.client.List(context.TODO(),
		prestoHPA,
		&client.ListOptions{
			Namespace:     presto.Namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		})
	if errors.IsNotFound(err) {
		return nil, errors.NewNotFound(v1.Resource("HPA"), "")
	}
	if err != nil {
		r.log.Error(err, "failed to list existing Presto HPA")
		return nil, err
	}
	if len(prestoHPA.Items) == 0 {
		return nil, errors.NewNotFound(v1.Resource("HPA"), "")
	}
	return &prestoHPA.Items[0], nil
}

func checkAutoscalingEnabled(presto *v1alpha1.Presto) bool {
	if presto.Spec.Worker.Autoscaling.Enabled == nil {
		return false
	} else {
		return *presto.Spec.Worker.Autoscaling.Enabled
	}
}

func isCreatedByHpaController(hpa *autoscalingv1.HorizontalPodAutoscaler, presto *v1alpha1.Presto) bool {
	for _, ref := range hpa.OwnerReferences {
		if ref.Name == presto.Name && ref.Kind == presto.Kind {
			return true
		}
	}
	return false
}

func createHPASpec(presto *v1alpha1.Presto,
	workerReplicaSet *v1.ReplicaSet,
	lbls map[string]string, resourceVersion string) (*autoscalingv1.HorizontalPodAutoscaler, error) {

	if presto.Spec.Worker.Autoscaling.MinReplicas == nil {
		return nil, &OperatorError{errormsg: "MinReplicas cannot be null"}
	}
	minReplicas := *presto.Spec.Worker.Autoscaling.MinReplicas

	if presto.Spec.Worker.Autoscaling.MaxReplicas == nil {
		return nil, &OperatorError{errormsg: "MaxReplicas cannot be null"}
	}
	maxReplicas := *presto.Spec.Worker.Autoscaling.MaxReplicas

	if presto.Spec.Worker.Autoscaling.TargetCPUUtilizationPercentage == nil {
		return nil, &OperatorError{errormsg: "TargetCPUUtilizationPercentage cannot be null"}
	}
	targetCPUUtilization := *presto.Spec.Worker.Autoscaling.TargetCPUUtilizationPercentage

	hpa := &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name: getHPAName(presto.Status.Uuid),
			Namespace: presto.Namespace,
			Labels: lbls,
			// resource version for optimistic concurrency control
			ResourceVersion: resourceVersion,
			OwnerReferences: []metav1.OwnerReference{
				*getOwnerReference(presto),
			},
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				APIVersion: workerReplicaSet.APIVersion,
				Kind:       workerReplicaSet.Kind,
				Name:       workerReplicaSet.Name,
			},
			MinReplicas: &minReplicas,
			MaxReplicas: maxReplicas,
			TargetCPUUtilizationPercentage:  &targetCPUUtilization,
		},
	}

	return hpa, nil
}