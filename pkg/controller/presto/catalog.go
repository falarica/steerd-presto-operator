package presto

import (
	"fmt"
	"github.com/falarica/steerd-presto-operator/pkg/apis/falarica/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func createCatalogConfig(presto *v1alpha1.Presto, r *ReconcilePresto,
	lbls map[string]string) (bool, error) {
	catalogConfigName := getCatalogConfigMapName(presto.Status.Uuid)
	configMap, err := buildCatalogConfigMap(presto, catalogConfigName, lbls)
	if err != nil {
		return false, err
	}
	return createConfigMap(catalogConfigName, presto, r.client, configMap, lbls)
}

func getCatalogVolumeMount(presto *v1alpha1.Presto, podSpec *corev1.PodSpec) *corev1.VolumeMount {
	numOfVolProjections := len(presto.Spec.Catalogs.CatalogSecrets) + 1
	volumeProjectionsCatalogs := make([]corev1.VolumeProjection, numOfVolProjections)
	volumeProjectionsCatalogs[0] = corev1.VolumeProjection{
		ConfigMap:           &corev1.ConfigMapProjection{
			LocalObjectReference: corev1.LocalObjectReference{
				Name:     getCatalogConfigMapName(presto.Status.Uuid),
			},
		},
	}
	for i ,catalogSecret := range presto.Spec.Catalogs.CatalogSecrets {
		catalogFileNames := make([]corev1.KeyToPath, 1)
		catalogFileNames[0] = corev1.KeyToPath{
			Key:  catalogSecret.SecretKey,
			Path: catalogSecret.SecretKey + catalogFileSuffix,
		}
		volumeProjectionsCatalogs[i + 1] =  corev1.VolumeProjection{
			Secret:           &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{Name: catalogSecret.SecretName},
				Items:                catalogFileNames,
			},
		}
	}

	catalogVolume := corev1.Volume{
		Name: getCatalogVolName(presto.Status.Uuid),
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: volumeProjectionsCatalogs,
			},
		},
	}

	podSpec.Volumes = append(podSpec.Volumes, catalogVolume)
	return &corev1.VolumeMount{
		Name:      getCatalogVolName(presto.Status.Uuid),
		ReadOnly:  true,
		//TODO: this depends on the docker file. Needs to be seen if it remains same
		MountPath: getPrestoPath(presto) + catalogMountPath,
	}
}

func getDefaultCatalogs(presto *v1alpha1.Presto) map[string]string{
	var defaultCatalogs = map[string]string {
		"jmx" : "connector.name=jmx\n",
		"tpch" : "connector.name=tpch\n",
		"tpcds" : "connector.name=tpcds\n",
	}
	// add default catalog if catalogs with same name are are not already added
	prunedDefaultCatalog := make(map[string]string)
	for catalogName, content := range defaultCatalogs {
		catalogPresentInSpec := false
		for _, specCatalog := range presto.Spec.Catalogs.CatalogSpec {
			if specCatalog.Name == catalogName {
				catalogPresentInSpec = true
			}
		}
		for _, specCatalogSecret := range presto.Spec.Catalogs.CatalogSecrets {
			if specCatalogSecret.SecretKey == catalogName {
				catalogPresentInSpec = true
			}
		}
		if !catalogPresentInSpec {
			prunedDefaultCatalog[catalogName] = content
		}
	}
	return prunedDefaultCatalog
}

func buildCatalogConfigMap(presto *v1alpha1.Presto,
	catalogConfigName string, labels map[string]string) (*corev1.ConfigMap, error) {
	catalogData := make(map[string]string)
	for _, catalog := range presto.Spec.Catalogs.CatalogSpec {
		var sb strings.Builder
		for key, value := range catalog.Content {
			sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}
		catalogData[catalog.Name + catalogFileSuffix] = sb.String()
	}

	defaultCatalogs := getDefaultCatalogs(presto)
	for k, v := range defaultCatalogs {
		// add .properties to the catalog name. as we are not asking that as part of catalog name
		catalogData[k + catalogFileSuffix] = v
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            catalogConfigName,
			Namespace:       presto.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*getOwnerReference(presto)},
		},
		Data: catalogData,
	}, nil
}