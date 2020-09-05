package presto

import (
	"fmt"
	"github.com/falarica/steerd-presto-operator/pkg/apis/falarica/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func buildConfigMap(presto *v1alpha1.Presto, isCoordinator bool, configMapName string,
	labels map[string]string) (*corev1.ConfigMap, error) {
	nodeProperties := ""
	if isCoordinator {
		nodeProperties = coordinatorNodePropsMap()
	} else {
		nodeProperties = workerNodePropsMap()

	}
	configProperties := ""
	var err error
	if isCoordinator {
		configProperties, err = coordinatorConfigPropsMap(presto)
	} else {
		configProperties, err = workerConfigPropsMap(presto)
	}
	if err != nil {
		return nil, err
	}
	jvmConfig := ""
	if isCoordinator {
		jvmConfig, err = coordinatorJVMConfigMap(presto)
	} else {
		jvmConfig, err = workerJVMConfigMap(presto)
	}
	if err != nil {
		return nil, err
	}
	propertiesFiles := map[string]string{
		nodePropertiesKey:   nodeProperties,
		configPropertiesKey: configProperties,
		jvmConfigKey:        jvmConfig,
		// added the shutdown script in etc folder to avoid mounting another volume
		prestoShutdownScript: strings.ReplaceAll(shutdownScriptContent, "{MOUNT_PATH}", getPrestoPath(presto)),
	}
	for filename, content := range presto.Spec.AdditionalPrestoPropFiles {
		propertiesFiles[filename] = content
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            configMapName,
			Namespace:       presto.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*getOwnerReference(presto)},
		},
		Data: propertiesFiles,
	}, nil
}

// return createdConfig, error
func createCoordinatorConfig(presto *v1alpha1.Presto, c client.Client,
	lbls map[string]string) (bool, error) {
	configMapName := getCoordinatorConfigMapName(presto.Status.Uuid)
	configMap, err := buildConfigMap(presto, true, configMapName, lbls)
	if err != nil {
		return false, err
	}
	return createConfigMap(configMapName, presto, c, configMap, lbls)
}

func createWorkerConfig(presto *v1alpha1.Presto, c client.Client,
	lbls map[string]string) (bool, error) {
	configMapName := getWorkerConfigMapName(presto.Status.Uuid)
	configMap, err := buildConfigMap(presto, false, configMapName, lbls)
	if err != nil {
		return false, err
	}
	return createConfigMap(configMapName, presto, c, configMap, lbls)
}

func coordinatorNodePropsMap() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("node.environment=prestoproduction\n"))
	sb.WriteString(fmt.Sprintf("node.data-dir=/data/presto\n"))
	return sb.String()
}

func coordinatorJVMConfigMap(presto *v1alpha1.Presto) (string, error) {
	var sb strings.Builder
	addDefaultJVMProps(&sb)
	memlimit, err := resource.ParseQuantity(presto.Spec.Coordinator.MemoryLimit)
	if err != nil {
		return "", &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Coordinator.MemoryLimit: " +
			"'%v': %v", presto.Spec.Coordinator.MemoryLimit, err)}
	}
	memoryMb := memlimit.Value()/1024/1024
	sb.WriteString(fmt.Sprintf("-Xmx%dm\n", memoryMb))
	// adding JVM config at the end as the right most will take effect
	if len(presto.Spec.Coordinator.AdditionalJVMConfig) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", presto.Spec.Coordinator.AdditionalJVMConfig))
	}

	return sb.String(), nil
}

func coordinatorConfigPropsMap(presto *v1alpha1.Presto) (string, error) {
	coordinatorInternalName := getCoordinatorInternalName(presto.Status.Uuid)
	systemProps, err := getSystemProps(presto, coordinatorInternalName)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for key, value := range systemProps {
		sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}
	for key, value := range presto.Spec.Coordinator.AdditionalProps {
		if _, ok := systemProps[key]; ok {
			return "", &OperatorError{"%s is a system property. Cannot be specified as additional property"}
		} else {
			sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}
	}
	return sb.String(), nil
}

func getHTTPPort(presto *v1alpha1.Presto) (int32, int32) {
	port := int32(prestoPort)
	if presto.Spec.Service.Port != nil {
		port = *presto.Spec.Service.Port
	}
	var httpport int32
	if presto.Spec.Coordinator.HttpsEnabled {
		if port == int32(prestoPort) {
			// a way to generate another port
			httpport = prestoPort + 1
		} else {
			httpport = int32(prestoPort)
		}
		return httpport, port
	} else{
		httpport = port
		return httpport, -1
	}
}

func getSystemProps(presto *v1alpha1.Presto, coordinatorInternalName string) (map[string]string, error) {
	httpPort, httpsPort := getHTTPPort(presto)

	var systemProps = make(map[string]string)
	if presto.Spec.Coordinator.HttpsEnabled {
		if  len(presto.Spec.Coordinator.HttpsKeyPairPassword) == 0 {
			return nil, &OperatorError{errormsg: "HttpsKeyPairPassword has to be specified when HTTPS is enabled"}
		}
		if  len(presto.Spec.Coordinator.HttpsKeyPairSecretKey) == 0 {
			return nil, &OperatorError{errormsg: "HttpsKeyPairSecretKey has to be specified when HTTPS is enabled"}
		}
		if  len(presto.Spec.Coordinator.HttpsKeyPairSecretName) == 0 {
			return nil, &OperatorError{errormsg: "HttpsKeyPairSecretName has to be specified when HTTPS is enabled"}
		}
		systemProps = map[string]string{
			"coordinator":                        "true",
			"node.internal-address":              coordinatorInternalName,
			"discovery.uri":                      fmt.Sprintf("http://%s:%d\n", coordinatorInternalName, httpPort),
			"node-scheduler.include-coordinator": "false",
			"discovery-server.enabled":           "true",
			"http-server.http.enabled":           "true",
			"http-server.https.enabled":          "true",
			"http-server.https.port":             fmt.Sprintf("%d", httpsPort),
			"http-server.http.port":              fmt.Sprintf("%d", httpPort),
			"http-server.https.keystore.path":    httpsVolPath + "/" + presto.Spec.Coordinator.HttpsKeyPairSecretKey,
			"http-server.https.keystore.key":     presto.Spec.Coordinator.HttpsKeyPairPassword,
		}
	} else {
		systemProps = map[string]string{
			"coordinator":                        "true",
			"http-server.http.port":              fmt.Sprintf("%d", httpPort),
			"node.internal-address":              coordinatorInternalName,
			"discovery.uri":                      fmt.Sprintf("http://%s:%d\n", coordinatorInternalName, httpPort),
			"node-scheduler.include-coordinator": "false",
			"discovery-server.enabled":           "true",
			"http-server.https.enabled":          "false",
		}
	}
	return systemProps, nil
}

func workerNodePropsMap() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("node.environment=prestoproduction\n"))
	//TODO: data-dir to be made configurable as a persistent volume
	sb.WriteString(fmt.Sprintf("node.data-dir=/data/presto\n"))
	return sb.String()
}

func workerJVMConfigMap(presto *v1alpha1.Presto) (string, error) {
	var sb strings.Builder
	addDefaultJVMProps(&sb)
	memlimit, err := resource.ParseQuantity(presto.Spec.Worker.MemoryLimit)
	if err != nil {
		return "", &OperatorError{fmt.Sprintf("cannot parse presto.Spec.Worker.MemoryLimit: " +
			"'%v': %v", presto.Spec.Worker.MemoryLimit, err)}
	}
	memoryMb := memlimit.Value()/1024/1024
	sb.WriteString(fmt.Sprintf("-Xmx%dm\n", memoryMb))
	// adding JVM config at the end as the right most will take effect
	if len(presto.Spec.Worker.AdditionalJVMConfig) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", presto.Spec.Worker.AdditionalJVMConfig))
	}
	return sb.String(), nil
}

func addDefaultJVMProps(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("-server\n"))
	sb.WriteString(fmt.Sprintf("-XX:-UseBiasedLocking\n"))
	sb.WriteString(fmt.Sprintf("-XX:+UseG1GC\n"))
	sb.WriteString(fmt.Sprintf("-XX:G1HeapRegionSize=32M\n"))
	sb.WriteString(fmt.Sprintf("-XX:+ExplicitGCInvokesConcurrent\n"))
	sb.WriteString(fmt.Sprintf("-XX:+ExitOnOutOfMemoryError\n"))
	sb.WriteString(fmt.Sprintf("-XX:+UseGCOverheadLimit\n"))
	sb.WriteString(fmt.Sprintf("-XX:+HeapDumpOnOutOfMemoryError\n"))
	sb.WriteString(fmt.Sprintf("-XX:ReservedCodeCacheSize=512M\n"))
	sb.WriteString(fmt.Sprintf("-Djdk.attach.allowAttachSelf=true\n"))
	sb.WriteString(fmt.Sprintf("-Djdk.nio.maxCachedBufferSize=2000000\n"))
}

func workerConfigPropsMap(presto *v1alpha1.Presto) (string, error) {
	httpPort, _ := getHTTPPort(presto)

	var systemProps = map[string]string {
		"coordinator": "false",
		"http-server.http.port": fmt.Sprintf("%d", prestoPort),
		"discovery.uri": fmt.Sprintf("http://%s:%d", getCoordinatorInternalName(presto.Status.Uuid), httpPort),
	}
	var sb strings.Builder
	for key, value := range systemProps {
		sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}
	for key, value := range presto.Spec.Worker.AdditionalProps {
		if  _, ok := systemProps[key]; ok {
			return "", &OperatorError{"%s is a system property. Cannot be specified as additional property"}
		} else {
			sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}
	}
	return sb.String(), nil
}

func getPropsVolumeMount(presto *v1alpha1.Presto, podSpec *corev1.PodSpec,
	isCoordinator bool) *corev1.VolumeMount {
	configMap := ""
	configMapVol := ""

	if isCoordinator {
		configMap = getCoordinatorConfigMapName(presto.Status.Uuid)
		configMapVol = getCoordinatorConfigVolumeName(presto.Status.Uuid)
	} else {
		configMap = getWorkerConfigMapName(presto.Status.Uuid)
		configMapVol = getWorkerConfigVolumeName(presto.Status.Uuid)
	}

	volumeProjectionsProperties := make([]corev1.VolumeProjection, 1)
	volumeProjectionsProperties[0] = corev1.VolumeProjection{
		ConfigMap:           &corev1.ConfigMapProjection{
			LocalObjectReference: corev1.LocalObjectReference{
				Name:configMap,
			},
		},
	}

	propsVolume := corev1.Volume{
		Name: configMapVol,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: volumeProjectionsProperties,
			},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, propsVolume)
	return &corev1.VolumeMount{
		Name:      configMapVol,
		ReadOnly:  true,
		MountPath: getPrestoPath(presto),
	}
}