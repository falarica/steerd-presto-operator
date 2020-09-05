package presto

import (
	"context"
	"fmt"
	"github.com/falarica/steerd-presto-operator/pkg/apis/falarica/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// This function is used to create headless service
// Headless service can allow communication using coordiator name rather than service IP
func createHeadLessPodDiscoverySvc(presto *v1alpha1.Presto, r *ReconcilePresto,
	lbls map[string]string) (error, bool) {
	created := false
	svcKey, svcLabelVal := getPodDiscoveryServiceLabel(presto.Status.Uuid)
		lbls[svcKey] = svcLabelVal
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getPodDiscoveryServiceName(presto.Status.Uuid),
			Namespace: presto.Namespace,
			Labels:    lbls,
			OwnerReferences: []metav1.OwnerReference{*getOwnerReference(presto)},
		},
		Spec:       corev1.ServiceSpec{
			Selector:                 lbls,
			ClusterIP:                "None",
		},
	}

	oldService := &corev1.ServiceList{}
	err := r.client.List(context.Background(),
		oldService,
		&client.ListOptions{
			Namespace:     presto.Namespace,
			LabelSelector: labels.SelectorFromSet(labels.Set{svcKey: svcLabelVal}),
		})
	if err != nil {
		return err, created
	}
	if len(oldService.Items) == 0 {
		createErr := r.client.Create(context.Background(), service)
		if createErr != nil {
			return createErr, created
		}
		created = true
	}
	return nil, created
}

func createOrGetService(presto *v1alpha1.Presto, r *ReconcilePresto,
	lbls map[string]string) (error, *corev1.Service, bool) {
	created := false
	svcKey, svcLabelVal := getExternalServiceLabel(presto.Status.Uuid)
	lbls[svcKey] = svcLabelVal
	wk, wv := getCoordinatorPodLabel(presto.Status.Uuid)
	if presto.Spec.Service.Type == corev1.ServiceTypeExternalName {
		return &OperatorError{fmt.Sprintf("Service of the following type" +
			" not supported: %s", corev1.ServiceTypeExternalName) }, nil, created
	}
	servicePort := getServicePort(presto)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            getExternalServiceName(presto.Status.Uuid),
			Namespace:       presto.Namespace,
			Labels:          lbls,
			OwnerReferences: []metav1.OwnerReference{*getOwnerReference(presto)},
		},
		Spec: corev1.ServiceSpec{
			Selector:                 labels.Set{wk: wv},
			ClusterIP:                presto.Spec.Service.ClusterIP,
			Type:                     presto.Spec.Service.Type,
			ExternalIPs:              presto.Spec.Service.ExternalIPs,
			SessionAffinity:          presto.Spec.Service.SessionAffinity,
			LoadBalancerIP:           presto.Spec.Service.LoadBalancerIP,
			LoadBalancerSourceRanges: presto.Spec.Service.LoadBalancerSourceRanges,
			ExternalName:             presto.Spec.Service.ExternalName,
			ExternalTrafficPolicy:    presto.Spec.Service.ExternalTrafficPolicy,
			HealthCheckNodePort:      presto.Spec.Service.HealthCheckNodePort,
			PublishNotReadyAddresses: presto.Spec.Service.PublishNotReadyAddresses,
			SessionAffinityConfig:    presto.Spec.Service.SessionAffinityConfig,
			IPFamily:                 presto.Spec.Service.IPFamily,
			Ports:                    servicePort,
		},
	}
	err, retService := getService(r, presto, labels.Set{svcKey: svcLabelVal})
	if err != nil {
		return err, nil, created
	}
	if retService == nil {
		createErr := r.client.Create(context.TODO(), service)
		if createErr != nil {
			return createErr, nil, created
		}
		created = true
		err, retService = getService(r, presto, lbls)
		if err != nil {
			return err, nil, created
		}
	}
	return nil, retService, created
}

func getService(r *ReconcilePresto, presto *v1alpha1.Presto,
	lbls map[string]string) (error, *corev1.Service){
	services := &corev1.ServiceList{}
	err := r.client.List(context.TODO(),
		services,
		&client.ListOptions{
			Namespace:     presto.Namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		})
	if err != nil {
		return err, nil
	}
	if len(services.Items) == 0 {
		return nil, nil
	} else {
		return nil, &services.Items[0]
	}

}

func getServicePort(presto *v1alpha1.Presto) []corev1.ServicePort {
	port := int32(prestoPort)
	if presto.Spec.Service.Port != nil {
		port = *presto.Spec.Service.Port
	}
	if presto.Spec.Service.Type != "ClusterIP" && presto.Spec.Service.NodePort != nil {
		return []corev1.ServicePort{
			{
				Name:     "presto-coordinator-port",
				Port:     port,
				NodePort: *presto.Spec.Service.NodePort,
				TargetPort: intstr.IntOrString{
					IntVal: port,
				},
			},
		}
	}
	return []corev1.ServicePort{
		{
			Name: "presto-coordinator-port",
			Port: port,
			TargetPort: intstr.IntOrString{
				IntVal: port,
			},
		},
	}
}
