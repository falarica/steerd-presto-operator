package presto

import (
	"context"
	"fmt"
	"github.com/falarica/steerd-presto-operator/pkg/apis/falarica/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getPodDiscoveryServiceName(clusterUUID string) string {
	return "pod-discovery-" + clusterUUID[:8]
}

func getPodDiscoveryServiceLabel(clusterUUID string) (string, string) {
	return "pod-discovery", clusterUUID
}

func getCoordinatorInternalName(clusterUUID string) string {
	return fmt.Sprintf("%s.%s",
		getCoordinatorContainerName(clusterUUID),
		getPodDiscoveryServiceName(clusterUUID))
}

func getExternalServiceName(clusterUUID string) string {
	return "external-presto-svc-" + clusterUUID[:8]
}

func getExternalServiceLabel(clusterUUID string) (string, string) {
	return "external-presto-svc", clusterUUID
}

func getCoordinatorPodName(clusterUUID string) string {
	return "coordinatorpod-" + clusterUUID[:8]
}
func getCoordinatorReplicaset(clusterUUID string) string {
	return "coordinatorreplicaset-" + clusterUUID[:8]
}

func getWorkerReplicaSet(clusterUUID string) string {
	return "workerreplicaset-" + clusterUUID[:8]
}

func getCatalogConfigMapName(clusterUUID string) string {
	return "catalogconfig-" + clusterUUID[:8]
}

func getCatalogVolName(clusterUUID string) string {
	return "catalogvol-" + clusterUUID[:8]
}

func getHTTPSSecretVolName(clusterUUID string) string {
	return "httpssecret-" + clusterUUID[:8]
}

func getHPAName(clusterUUID string) string {
	return "hpa-" + clusterUUID[:8]
}

func getCoordinatorContainerName(clusterUUID string) string {
	return "coordinatorcontainer-" + clusterUUID[:8]
}

func getCoordinatorConfigMapName(clusterUUID string) string {
	return "coordinatorconfig-" + clusterUUID[:8]
}

func getCoordinatorConfigVolumeName(clusterUUID string) string {
	return "coordinatorconfvol-" + clusterUUID[:8]
}
func getWorkerContainerPrefix(clusterUUID string) string {
	return fmt.Sprintf("workerrcontainer-%s", clusterUUID[:8])
}

func getCoordinatorPodLabel(clusterUUID string) (string, string) {
	return "coordinator", clusterUUID
}

func getWorkerPodLabel(clusterUUID string) (string, string) {
	return "worker", clusterUUID
}

func getCoordinatorPodLabels(baseLabels map[string]string, clusterUUID string) map[string]string {
	lbls := make(map[string]string)
	for key, value := range baseLabels {
		lbls[key] = value
	}
	k,v  := getCoordinatorPodLabel(clusterUUID)
	lbls[k] = v
	return lbls;
}

func getWorkerPodLabels(baseLabels map[string]string, clusterUUID string) map[string]string {
	lbls := make(map[string]string)
	for key, value := range baseLabels {
		lbls[key] = value
	}
	k,v  := getWorkerPodLabel(clusterUUID)
	lbls[k] = v
	return lbls;
}

func getWorkerConfigMapName(clusterUUID string) string {
	return "workerconfig-" + clusterUUID[:8]
}
func getWorkerConfigVolumeName(clusterUUID string) string {
	return "workerconfvol-" + clusterUUID[:8]
}

func getOwnerReference(app *v1alpha1.Presto) *metav1.OwnerReference {
	controller := true
	return &metav1.OwnerReference{
		APIVersion: app.APIVersion,
		Kind:       app.Kind,
		Name:       app.Name,
		UID:        app.UID,
		Controller: &controller,
	}
}

func createConfigMap(configMapName string, presto *v1alpha1.Presto, c client.Client,
	configMap *corev1.ConfigMap, lbls map[string]string) (bool, error) {
	oldConfigMaps := &corev1.ConfigMapList{}
	err := c.List(context.TODO(),
		oldConfigMaps,
		&client.ListOptions{
			Namespace:     presto.Namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		})
	if err != nil {
		return false, err
	}
	for _, cm := range oldConfigMaps.Items {
		if cm.Name == configMapName {
			return false, nil
		}
	}
	createErr := c.Create(context.Background(), configMap)
	if createErr != nil {
		return false, createErr
	} else {
		return true, createErr
	}
}

func getPrestoPath(presto *v1alpha1.Presto) string {
	prestoPath := presto.Spec.ImageDetails.PrestoPath
	if len(presto.Spec.ImageDetails.PrestoPath) == 0 {
		prestoPath = mountPath
	}
	return prestoPath
}
