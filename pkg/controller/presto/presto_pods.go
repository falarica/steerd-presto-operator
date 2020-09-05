package presto

import (
	"context"
	"fmt"
	falaricav1alpha1 "github.com/falarica/steerd-presto-operator/pkg/apis/falarica/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// returns replicaSet, created, updated, error
func createUpdateReplicaSetForWorker(r *ReconcilePresto, presto *falaricav1alpha1.Presto,
	lbls map[string]string) (*v1.ReplicaSet,bool, bool, error) {
	updated := false
	created := false
	// Get the replicaSet with the name specified in PrestoCluster.spec
	replicaSet, err := getReplicaSet(r, presto, getWorkerPodLabel)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		replicaSet, err = createReplicaSetForWorker(r, presto, lbls, *presto.Spec.Worker.Count)
		if err != nil {
			r.log.Error(err,"Failed to create replicaSet object")
			return nil, created, updated, err
		}
		err = r.client.Create(context.Background(), replicaSet)
		if err != nil {
			r.log.Error(err,"Failed to create replicaSet")
			return nil, created, updated, err
		}
		created = true
	} else {
		if err != nil {
			r.log.Error(err, "Failed to get replicaSet")
			return nil, created, updated, err
		}
		// worker count shall be updated only if autoscaling is not enabled
		if presto.Spec.Worker.Autoscaling.Enabled == nil ||
			!*presto.Spec.Worker.Autoscaling.Enabled {
			// If the number of the replicas is not equal the current replicas on the ReplicaSet, update it
			if presto.Spec.Worker.Count != nil && *presto.Spec.Worker.Count != *replicaSet.Spec.Replicas {
				r.log.Info(fmt.Sprintf("PrestoCluster %s workerCount: %d, replicaSet replicas: %d",
					presto.Name, *presto.Spec.Worker.Count, *replicaSet.Spec.Replicas))
				replicaSetCopy := replicaSet.DeepCopy()
				replicaSetCopy.Spec.Replicas = presto.Spec.Worker.Count
				err = r.client.Update(context.Background(), replicaSetCopy)
				replicaSet = replicaSetCopy
				if err != nil {
					return nil, created, updated, err
				}
				updated = true
			}
		}
	}
	return replicaSet, created, updated, nil
}

func createReplicaSetForWorker(r *ReconcilePresto, presto *falaricav1alpha1.Presto,
	lbls map[string]string, workerCount int32) (*v1.ReplicaSet, error) {
	podSpec, err := getPrestoWorkerPod(r, presto)
	if err != nil {
		return nil, err
	}
	return &v1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: getWorkerReplicaSet(presto.Status.Uuid),
			Namespace:    presto.Namespace,
			OwnerReferences: []metav1.OwnerReference{*getOwnerReference(presto)},
			Labels: lbls,
		},
		Spec: v1.ReplicaSetSpec{
			Replicas: func() *int32 { i := workerCount; return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: lbls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{*getOwnerReference(presto)},
					Namespace:    presto.Namespace,
					Labels: lbls,
				},
				Spec: *podSpec,
			},
		},
	}, nil
}

// returns podCreated, podDeleted, error
func getPrestoWorkerPod(r *ReconcilePresto, presto *falaricav1alpha1.Presto) (*corev1.PodSpec, error) {
	limitResource := corev1.ResourceList{}
	requestResource := corev1.ResourceList{}
	var err error
	limitResource[corev1.ResourceCPU], err = resource.ParseQuantity(presto.Spec.Worker.CpuLimit)
	if err != nil {
		return nil, &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Worker.CpuLimit: " +
			"'%v': %v", presto.Spec.Worker.CpuLimit, err)}
	}
	limitResource[corev1.ResourceMemory], err = resource.ParseQuantity(presto.Spec.Worker.MemoryLimit)
	if err != nil {
		return nil, &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Worker.MemoryLimit: " +
			"'%v': %v", presto.Spec.Worker.MemoryLimit, err)}
	}
	if len(presto.Spec.Worker.CpuRequest) == 0 {
		// set CPURequest same as limit if not specified
		requestResource[corev1.ResourceCPU], err = resource.ParseQuantity(presto.Spec.Worker.CpuLimit)
		if err != nil {
			return nil, &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Worker.CpuLimit: " +
				"'%v': %v", presto.Spec.Worker.CpuLimit, err)}
		}
	} else {
		requestResource[corev1.ResourceCPU], err = resource.ParseQuantity(presto.Spec.Worker.CpuRequest)
		if err != nil {
			return nil, &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Worker.CpuRequest: " +
				"'%v': %v", presto.Spec.Worker.CpuRequest, err)}
		}
	}
	return createPrestoPodSpec(r, presto, false,
		limitResource, requestResource), nil
}


func getReplicaSet(r *ReconcilePresto,presto *falaricav1alpha1.Presto,
	getLabel func(string)(string, string)) (*v1.ReplicaSet, error) {
	existingReplicaSet := &v1.ReplicaSetList{}
	wk, wv := getLabel(presto.Status.Uuid)
	err := r.client.List(context.TODO(),
		existingReplicaSet,
		&client.ListOptions{
			Namespace:     presto.Namespace,
			LabelSelector: labels.SelectorFromSet(labels.Set{wk: wv}),
		})
	if errors.IsNotFound(err) {
		return nil, errors.NewNotFound(v1.Resource("replicasets"), "")
	}
	if err != nil {
		r.log.Error(err, "failed to list existing Presto replicaset")
		return nil, err
	}
	if len(existingReplicaSet.Items) == 0 {
		return nil, errors.NewNotFound(v1.Resource("replicasets"), "")
	}
	return &existingReplicaSet.Items[0], nil
}

// returns replicaSet, created, error
func createUpdateReplicaSetForCoordinator(r *ReconcilePresto, presto *falaricav1alpha1.Presto,
	lbls map[string]string) (*v1.ReplicaSet, bool, error) {
	created := false
	// Get the replicaSet with the name specified in PrestoCluster.spec
	replicaSet, err := getReplicaSet(r, presto, getCoordinatorPodLabel)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		replicaSet, err = createReplicaSetForCoordinator(r, presto, lbls)
		if err != nil {
			return nil,created, err
		}
		err = r.client.Create(context.Background(), replicaSet)
		if err != nil {
			return nil, created, err
		}
		created = true
	}
	if err != nil {
		return nil, created, err
	}
	return replicaSet, created, nil
}

func createReplicaSetForCoordinator(r *ReconcilePresto, presto *falaricav1alpha1.Presto,
	lbls map[string]string) (*v1.ReplicaSet, error) {
	podSpec, err := getPrestoCoordinatorPodSpec(r, presto, lbls)
	if err != nil {
		return nil, err
	}
	return &v1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: getCoordinatorReplicaset(presto.Status.Uuid),
			Namespace:    presto.Namespace,
			OwnerReferences: []metav1.OwnerReference{*getOwnerReference(presto)},
			Labels: lbls,
		},
		Spec: v1.ReplicaSetSpec{
			Replicas: func() *int32 { i := int32(1); return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: lbls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    presto.Namespace,
					OwnerReferences: []metav1.OwnerReference{*getOwnerReference(presto)},
					Labels: lbls,
				},
				Spec: *podSpec,
			},
		},
	}, nil
}

//  returns whether the pod has been created or not
func getPrestoCoordinatorPodSpec(r *ReconcilePresto, presto *falaricav1alpha1.Presto,
	labels map[string]string) (*corev1.PodSpec, error) {
	limitResource := corev1.ResourceList{}
	requestResource := corev1.ResourceList{}
	var err error
	limitResource[corev1.ResourceCPU], err = resource.ParseQuantity(presto.Spec.Coordinator.CpuLimit)
	if err != nil {
		return nil, &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Coordinator.CpuLimit: " +
			"'%v': %v", presto.Spec.Coordinator.CpuLimit, err)}
	}
	limitResource[corev1.ResourceMemory], err = resource.ParseQuantity(presto.Spec.Worker.MemoryLimit)
	if err != nil {
		return nil, &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Worker.MemoryLimit: " +
			"'%v': %v", presto.Spec.Worker.MemoryLimit, err)}
	}
	if len(presto.Spec.Coordinator.CpuRequest) == 0 {
		// set CPURequest same as limit if not specified
		requestResource[corev1.ResourceCPU], err = resource.ParseQuantity(presto.Spec.Coordinator.CpuLimit)
		if err != nil {
			return nil, &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Coordinator.CpuLimit: " +
				"'%v': %v", presto.Spec.Coordinator.CpuLimit, err)}
		}
	} else {
		requestResource[corev1.ResourceCPU], err = resource.ParseQuantity(presto.Spec.Coordinator.CpuRequest)
		if err != nil {
			return nil, &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Coordinator.CpuRequest: " +
				"'%v': %v", presto.Spec.Coordinator.CpuRequest, err)}
		}
	}
	return createPrestoPodSpec(r, presto, true,
		limitResource, requestResource), nil
}

// Returns podCreated, error
func createPrestoPodSpec(r *ReconcilePresto, presto *falaricav1alpha1.Presto,
	isCoordinator bool, limitResource corev1.ResourceList,
	requestResource corev1.ResourceList) *corev1.PodSpec {
	imageName := presto.Spec.ImageDetails.Name
	if len(imageName) == 0 {
		imageName = "prestosql/presto:333"
	}

	var containerPrefix string
	var lifecycle *corev1.Lifecycle
	var terminationGraceSeconds *int64 = nil

	if isCoordinator {
		lifecycle = nil
		containerPrefix = getCoordinatorContainerName(presto.Status.Uuid)
	} else {
		containerPrefix = getWorkerContainerPrefix(presto.Status.Uuid)
		if presto.Spec.Worker.TerminationGracePeriodSeconds == nil {
			defaultGraceSeconds := int64(DefaultTerminationGracePeriodSeconds)
			terminationGraceSeconds = &defaultGraceSeconds
		} else {
			terminationGraceSeconds = presto.Spec.Worker.TerminationGracePeriodSeconds
		}

		lifecycle = &corev1.Lifecycle{
			PostStart: nil,
			PreStop:   &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/sh", fmt.Sprintf("%s/%s", getPrestoPath(presto), prestoShutdownScript)},
				},
			},
		}
	}

	podSpec := &corev1.PodSpec{
		TerminationGracePeriodSeconds:  terminationGraceSeconds,
		Containers: []corev1.Container {
			{
				Name:    containerPrefix,
				Image:   imageName,
				Resources: corev1.ResourceRequirements{
					Limits:   limitResource,
					Requests: requestResource,
				},
				Lifecycle: lifecycle,
			},
		},
	}

	if isCoordinator {
		podSpec.Hostname = getCoordinatorContainerName(presto.Status.Uuid)
		podSpec.Subdomain = getPodDiscoveryServiceName(presto.Status.Uuid)
	}
	catalogMount := getCatalogVolumeMount(presto, podSpec)
	propsMount := getPropsVolumeMount(presto, podSpec, isCoordinator)
	appendAdditionalVolumes(presto, &podSpec.Volumes)
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, *propsMount)
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, *catalogMount)
	if isCoordinator && presto.Spec.Coordinator.HttpsEnabled {
		httpsMount := getHTTPSVolumeMount(presto, podSpec)
		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, *httpsMount)
	}
	appendAdditionalVolumeMounts(presto, &podSpec.Containers[0].VolumeMounts)
	return podSpec
}

func appendAdditionalVolumes(presto *falaricav1alpha1.Presto,
	vols *[]corev1.Volume) {
	for _, volSpec := range presto.Spec.Volumes {
		vol := corev1.Volume {
			Name:  volSpec.Name,
			VolumeSource: volSpec.VolumeSource,
		}
		*vols = append(*vols, vol)
	}
}

func appendAdditionalVolumeMounts(presto *falaricav1alpha1.Presto,
	volMounts *[]corev1.VolumeMount) {
	for _, volSpec := range presto.Spec.Volumes {
		volMount := corev1.VolumeMount{
			Name:             volSpec.Name,
			ReadOnly:         volSpec.ReadOnly,
			MountPath:        volSpec.MountPath,
			SubPath:          volSpec.SubPath,
			MountPropagation: volSpec.MountPropagation,
			SubPathExpr:      volSpec.SubPathExpr,
		}
		*volMounts = append(*volMounts, volMount)
	}
}

func getHTTPSVolumeMount(presto *falaricav1alpha1.Presto,
	podSpec *corev1.PodSpec) *corev1.VolumeMount {
	httpsSecretVolume := corev1.Volume{
		Name: getHTTPSSecretVolName(presto.Status.Uuid),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  presto.Spec.Coordinator.HttpsKeyPairSecretName,
			},
		},
	}

	podSpec.Volumes = append(podSpec.Volumes, httpsSecretVolume)
	return &corev1.VolumeMount{
		Name:      getHTTPSSecretVolName(presto.Status.Uuid),
		ReadOnly:  true,
		MountPath: httpsVolPath,
	}
}