package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:openapi-gen=true
type CoordinatorSpec struct {
	// +kubebuilder:validation:Required
	MemoryLimit string `json:"memoryLimit"`
	// +kubebuilder:validation:Optional
	AdditionalJVMConfig string `json:"additionalJVMConfig,omitempty"`
	// +kubebuilder:validation:Optional
	AdditionalProps map[string]string `json:"additionalProps,omitempty"`
	// +kubebuilder:validation:Required
	CpuLimit string `json:"cpuLimit"`
	// +kubebuilder:validation:Optional
	CpuRequest string `json:"cpuRequest,omitempty"`
	// +kubebuilder:validation:Optional
	HttpsEnabled bool `json:"httpsEnabled,omitempty"`
	// +kubebuilder:validation:Optional
	HttpsKeyPairSecretName string `json:"httpsKeyPairSecretName,omitempty"`
	// +kubebuilder:validation:Optional
	HttpsKeyPairSecretKey string `json:"httpsKeyPairSecretKey,omitempty"`
	// +kubebuilder:validation:Optional
	HttpsKeyPairPassword string `json:"httpsKeyPairPassword,omitempty"`
}

// +k8s:openapi-gen=true
type WorkerSpec struct {
	// +kubebuilder:validation:Required
	MemoryLimit string `json:"memoryLimit"`
	// +kubebuilder:validation:Optional
	AdditionalJVMConfig string `json:"additionalJVMConfig,omitempty"`
	// +kubebuilder:validation:Optional
	AdditionalProps map[string]string `json:"additionalProps,omitempty"`
	// +kubebuilder:validation:Required
	CpuLimit string `json:"cpuLimit"`
	// +kubebuilder:validation:Optional
	CpuRequest string `json:"cpuRequest,omitempty"`
	// Optional duration in seconds the pod needs to terminate gracefully.
	// Value must be non-negative integer. The value zero indicates delete immediately.
	// If this value is nil, the default grace period will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// Defaults to 7200 seconds.
	// +optional
	// +kubebuilder:validation:Optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// +kubebuilder:validation:Maximum=10000
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Required
	Count *int32 `json:"count"`

	// +kubebuilder:validation:Optional
	Autoscaling AutoscalingSpec `json:"autoscaling,omitempty"`
}

// +k8s:openapi-gen=true
type AutoscalingSpec struct  {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Maximum=10000
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Maximum=10000
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	TargetCPUUtilizationPercentage  *int32 `json:"targetCPUUtilizationPercentage,omitempty"`
}

// +k8s:openapi-gen=true
type CatalogSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Content map[string]string `json:"content"`
}

// +k8s:openapi-gen=true
// one has to create catalogs as secret using the following command in order to use this
// kubectl create secret generic secretName --from-literal=secretKey1='connector.name=tpch' --from-literal=secretKey2='connector.name=tpcds'
type CatalogSecret struct {
	// +kubebuilder:validation:Required
	SecretName string `json:"secretName,omitempty"`
	// +kubebuilder:validation:Required
	SecretKey string `json:"secretKey,omitempty"`
}

// +k8s:openapi-gen=true
type CatalogList struct {
	// Secret names in the same namespace
	// +kubebuilder:validation:Optional
	CatalogSecrets []CatalogSecret `json:"catalogSecrets,omitempty"`
	// +kubebuilder:validation:Optional
	CatalogSpec []CatalogSpec `json:"catalogSpec,omitempty"`
}

// +k8s:openapi-gen=true
// ServiceSpec describes the attributes that a user creates on a service.
// Following is a copy of v1.ServiceSpec except that Ports is an optional field and
// Selectors field is removed.
type ServiceSpec struct {
	// +kubebuilder:validation:Optional
	NodePort *int32 `json:"nodePort,omitempty"`
	// +kubebuilder:validation:Optional
	Port *int32 `json:"port,omitempty"`

	// clusterIP is the IP address of the service and is usually assigned
	// randomly by the master. If an address is specified manually and is not in
	// use by others, it will be allocated to the service; otherwise, creation
	// of the service will fail. This field can not be changed through updates.
	// Valid values are "None", empty string (""), or a valid IP address. "None"
	// can be specified for headless services when proxying is not required.
	// Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if
	// type is ExternalName.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// +kubebuilder:validation:Optional
	ClusterIP string `json:"clusterIP,omitempty" protobuf:"bytes,3,opt,name=clusterIP"`

	// type determines how the Service is exposed. Defaults to ClusterIP. Valid
	// options are ExternalName, ClusterIP, NodePort, and LoadBalancer.
	// "ExternalName" maps to the specified externalName.
	// "ClusterIP" allocates a cluster-internal IP address for load-balancing to
	// endpoints. Endpoints are determined by the selector or if that is not
	// specified, by manual construction of an Endpoints object. If clusterIP is
	// "None", no virtual IP is allocated and the endpoints are published as a
	// set of endpoints rather than a stable IP.
	// "NodePort" builds on ClusterIP and allocates a port on every node which
	// routes to the clusterIP.
	// "LoadBalancer" builds on NodePort and creates an
	// external load-balancer (if supported in the current cloud) which routes
	// to the clusterIP.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	// +kubebuilder:validation:Optional
	Type v1.ServiceType `json:"type,omitempty" protobuf:"bytes,4,opt,name=type,casttype=ServiceType"`

	// externalIPs is a list of IP addresses for which nodes in the cluster
	// will also accept traffic for this service.  These IPs are not managed by
	// Kubernetes.  The user is responsible for ensuring that traffic arrives
	// at a node with this IP.  A common example is external load-balancers
	// that are not part of the Kubernetes system.
	// +kubebuilder:validation:Optional
	ExternalIPs []string `json:"externalIPs,omitempty" protobuf:"bytes,5,rep,name=externalIPs"`

	// Supports "ClientIP" and "None". Used to maintain session affinity.
	// Enable client IP based session affinity.
	// Must be ClientIP or None.
	// Defaults to None.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// +kubebuilder:validation:Optional
	SessionAffinity v1.ServiceAffinity `json:"sessionAffinity,omitempty" protobuf:"bytes,7,opt,name=sessionAffinity,casttype=ServiceAffinity"`

	// Only applies to Service Type: LoadBalancer
	// LoadBalancer will get created with the IP specified in this field.
	// This feature depends on whether the underlying cloud-provider supports specifying
	// the loadBalancerIP when a load balancer is created.
	// This field will be ignored if the cloud-provider does not support the feature.
	// +kubebuilder:validation:Optional
	LoadBalancerIP string `json:"loadBalancerIP,omitempty" protobuf:"bytes,8,opt,name=loadBalancerIP"`

	// If specified and supported by the platform, this will restrict traffic through the cloud-provider
	// load-balancer will be restricted to the specified client IPs. This field will be ignored if the
	// cloud-provider does not support the feature."
	// More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/
	// +kubebuilder:validation:Optional
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty" protobuf:"bytes,9,opt,name=loadBalancerSourceRanges"`

	// externalName is the external reference that kubedns or equivalent will
	// return as a CNAME record for this service. No proxying will be involved.
	// Must be a valid RFC-1123 hostname (https://tools.ietf.org/html/rfc1123)
	// and requires Type to be ExternalName.
	// +kubebuilder:validation:Optional
	ExternalName string `json:"externalName,omitempty" protobuf:"bytes,10,opt,name=externalName"`

	// externalTrafficPolicy denotes if this Service desires to route external
	// traffic to node-local or cluster-wide endpoints. "Local" preserves the
	// client source IP and avoids a second hop for LoadBalancer and Nodeport
	// type services, but risks potentially imbalanced traffic spreading.
	// "Cluster" obscures the client source IP and may cause a second hop to
	// another node, but should have good overall load-spreading.
	// +kubebuilder:validation:Optional
	ExternalTrafficPolicy v1.ServiceExternalTrafficPolicyType `json:"externalTrafficPolicy,omitempty" protobuf:"bytes,11,opt,name=externalTrafficPolicy"`

	// healthCheckNodePort specifies the healthcheck nodePort for the service.
	// If not specified, HealthCheckNodePort is created by the service api
	// backend with the allocated nodePort. Will use user-specified nodePort value
	// if specified by the client. Only effects when Type is set to LoadBalancer
	// and ExternalTrafficPolicy is set to Local.
	// +kubebuilder:validation:Optional
	HealthCheckNodePort int32 `json:"healthCheckNodePort,omitempty" protobuf:"bytes,12,opt,name=healthCheckNodePort"`

	// publishNotReadyAddresses, when set to true, indicates that DNS implementations
	// must publish the notReadyAddresses of subsets for the Endpoints associated with
	// the Service. The default value is false.
	// The primary use case for setting this field is to use a StatefulSet's Headless Service
	// to propagate SRV records for its Pods without respect to their readiness for purpose
	// of peer discovery.
	// +kubebuilder:validation:Optional
	PublishNotReadyAddresses bool `json:"publishNotReadyAddresses,omitempty" protobuf:"varint,13,opt,name=publishNotReadyAddresses"`

	// sessionAffinityConfig contains the configurations of session affinity.
	// +kubebuilder:validation:Optional
	SessionAffinityConfig *v1.SessionAffinityConfig `json:"sessionAffinityConfig,omitempty" protobuf:"bytes,14,opt,name=sessionAffinityConfig"`

	// +kubebuilder:validation:Optional
	IPFamily *v1.IPFamily `json:"ipFamily,omitempty" protobuf:"bytes,15,opt,name=ipFamily,Configcasttype=IPFamily"`
}

// +k8s:openapi-gen=true
type HMSSpec struct {
}

// +k8s:openapi-gen=true
type ImageSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	//	+kubebuilder:validation:Optional
	PrestoPath string `json:"prestoPath"`
}

// PrestoSpec defines the desired state of Presto
// +k8s:openapi-gen=true
type PrestoSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	//	+kubebuilder:validation:Required
	Coordinator CoordinatorSpec `json:"coordinator"`
	// +kubebuilder:validation:Required
	Worker WorkerSpec `json:"worker"`
	// +kubebuilder:validation:Optional
	Catalogs CatalogList `json:"catalogs,omitempty"`
	// +kubebuilder:validation:Optional
	Service ServiceSpec `json:"service,omitempty"`
	// +kubebuilder:validation:Optional
	InternalHiveMetaStore HMSSpec `json:"internalHiveMetaStore",omitempty`
	// +kubebuilder:validation:Optional
	ImageDetails ImageSpec `json:"imageDetails",omitempty`
	//additionalPrestoPropFiles:
	//   access-control.properties: |
	//    access-control.name=read-only
	//  event-listener.properties: |
	//    event-listener.name=event-logger
	//    jdbc.url=jdbc:postgresql://example.com:5432/eventlog
	//    jdbc.user=myuser
	//    jdbc.password=mypassword
	// +kubebuilder:validation:Optional
	AdditionalPrestoPropFiles map[string]string `json:"additionalPrestoPropFiles",omitempty`
	Volumes []PrestoVolumeSpec `json:"volumes,omitempty"`
}

type PrestoVolumeSpec struct {
	Name string `json:"name"`

	v1.VolumeSource `json:",inline"`

	// Mounted read-only if true, read-write otherwise (false or unspecified).
	// Defaults to false.
	// +kubebuilder:validation:Optional
	ReadOnly bool `json:"readOnly,omitempty"`
	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath"`
	// Path within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	// +kubebuilder:validation:Optional
	SubPath string `json:"subPath,omitempty"`
	// mountPropagation determines how mounts are propagated from the host
	// to container and the other way around.
	// When not set, MountPropagationNone is used.
	// This field is beta in 1.10.
	// +kubebuilder:validation:Optional
	MountPropagation *v1.MountPropagationMode `json:"mountPropagation,omitempty"`
	// Expanded path within the volume from which the container's volume should be mounted.
	// Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment.
	// Defaults to "" (volume's root).
	// SubPathExpr and SubPath are mutually exclusive.
	// This field is beta in 1.15.
	// +kubebuilder:validation:Optional
	SubPathExpr string `json:"subPathExpr,omitempty"`
}

// PrestoStatus defines the observed state of Presto
// +k8s:openapi-gen=true
type PrestoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Uuid        string `json:"uuid"`
	// +kubebuilder:validation:Optional
	DesiredWorkers int32 `json:"desiredWorkers"`
	// +kubebuilder:validation:Optional
	CurrentWorkers int32 `json:"currentWorkers"`
	// +kubebuilder:validation:Optional
	HeadlessService string `json:"headlessService"`
	// +kubebuilder:validation:Optional
	Service string `json:"service"`
	// +kubebuilder:validation:Optional
	CoordinatorAddress string `json:"coordinatorAddress"`
	// +kubebuilder:validation:Optional
	CatalogConfig string `json:"catalogConfig"`
	// +kubebuilder:validation:Optional
	CoordinatorConfig string `json:"coordinatorConfig"`
	// +kubebuilder:validation:Optional
	WorkerConfig string `json:"workerConfig"`
	// +kubebuilder:validation:Optional
	WorkerReplicaset string `json:"workerReplicaset"`
	// +kubebuilder:validation:Optional
	CoordinatorReplicaset string `json:"coordinatorReplicaset"`
	// +kubebuilder:validation:Optional
	HpaName string `json:"hpaName"`
	// +kubebuilder:validation:Optional
	ClusterState ClusterState `json:"clusterState"`
	// +kubebuilder:validation:Optional
	ErrorReason string `json:"errorReason"`
	// +kubebuilder:validation:Optional
	ModificationTime metav1.Time `json:"modificationTime,omitempty"`
	// +kubebuilder:validation:Optional
	CoordinatorCPU string `json:"coordinatorCPU,omitempty"`
	// +kubebuilder:validation:Optional
	WorkerCPU string `json:"workerCPU,omitempty"`
}

// +k8s:openapi-gen=true
type ClusterState string

const (
	ClusterFailedState ClusterState = "Failed"
	ClusterReadyState  ClusterState = "Ready"
	ClusterPending ClusterState = "Pending"
	ClusterUnknown ClusterState = "Unknown"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Presto is the Schema for the prestos API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=prestos,scope=Namespaced
// +kubebuilder:printcolumn:name="Coordinator",type="string",JSONPath=`.status.coordinatorAddress`
// +kubebuilder:printcolumn:name="ClusterState",type="string",JSONPath=`.status.clusterState`
// +kubebuilder:printcolumn:name="CoordinatorCPU",type="string",JSONPath=`.status.coordinatorCPU`
// +kubebuilder:printcolumn:name="WorkersCPU",type="string",JSONPath=`.status.workerCPU`
// +kubebuilder:printcolumn:name="DesiredWorkers",type="string",JSONPath=`.status.desiredWorkers`
// +kubebuilder:printcolumn:name="CurrentWorkers",type="string",JSONPath=`.status.currentWorkers`
type Presto struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrestoSpec   `json:"spec,omitempty"`
	Status PrestoStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrestoList contains a list of Presto
type PrestoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Presto `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Presto{}, &PrestoList{})
}
