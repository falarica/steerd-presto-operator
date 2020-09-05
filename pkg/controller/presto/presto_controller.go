package presto

import (
	"context"
	"fmt"
	falaricav1alpha1 "github.com/falarica/steerd-presto-operator/pkg/apis/falarica/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
	"time"
)

const (
	ControllerName string = "presto-controller"
)
type CommandLineParams struct {
	StatusUpdateInterval int
}


/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Presto Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, metricsClient *metrics.MetricsV1beta1Client,
	cmdLineParams CommandLineParams, log logr.Logger) error {
	return add(mgr, newReconciler(mgr, metricsClient, cmdLineParams, log))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, metricsClient *metrics.MetricsV1beta1Client,
	cmdLineParams CommandLineParams, log logr.Logger) ReconcilePresto {
	periodicPrestoEventChannel := make(chan event.GenericEvent)

	r := ReconcilePresto{
		client:               mgr.GetClient(),
		log:                  log,
		scheme:               mgr.GetScheme(),
		metricsClient:        metricsClient,
		eventRecorder:        mgr.GetEventRecorderFor(ControllerName),
		periodicPrestoEvents: periodicPrestoEventChannel,
		registeredPrestos:    new(sync.Map),
	}
	// a goroutine that sends periodic events for the registered prestos to a channel
	// that channel is being watched by the controlller.
	go func(r *ReconcilePresto) {
		ticker := time.NewTicker(time.Duration(cmdLineParams.StatusUpdateInterval) * time.Second)
		r.log.Info("Starting periodic event generation for presto clusters ")
		for {
			select {
			case  <-ticker.C:
				r.registeredPrestos.Range(func(key, value interface{}) bool {
					v := value.(event.GenericEvent)
					periodicPrestoEventChannel <- v
					return true
				})
			}
		}
	}(&r)

	return r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r ReconcilePresto) error {
	// Create a new controller. Set MaxConcurrentReconciles as 1. So there is only one thread
	c, err := controller.New(ControllerName, mgr, controller.Options{MaxConcurrentReconciles: 1, Reconciler: &r})
	if err != nil {
		return err
	}
	// watch the channel where periodic events are published for registered presto clusters
	err = c.Watch(&source.Channel{Source: r.periodicPrestoEvents}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Presto
	err = c.Watch(&source.Kind{Type: &falaricav1alpha1.Presto{}}, &handler.EnqueueRequestForObject{}, GenerationChangedPredicate{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &falaricav1alpha1.Presto{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &falaricav1alpha1.Presto{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.ReplicaSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &falaricav1alpha1.Presto{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcilePresto implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePresto{}

// ReconcilePresto reconciles a Presto object
type ReconcilePresto struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client               client.Client
	log                  logr.Logger
	scheme               *runtime.Scheme
	metricsClient        *metrics.MetricsV1beta1Client
	eventRecorder        record.EventRecorder
	periodicPrestoEvents chan event.GenericEvent
	registeredPrestos    *sync.Map
}

// Reconcile reads that state of the cluster for a Presto object and makes changes based on the state read
// and what is in the Presto.Spec
func (r *ReconcilePresto) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	_ = r.log.WithValues("deployment", request.NamespacedName)
	presto := &falaricav1alpha1.Presto{}
	err := r.client.Get(ctx, request.NamespacedName, presto)
	if err != nil {
		if errors.IsNotFound(err) {
			r.log.Info("Un-Registering for periodic events " + request.NamespacedName.String())
			r.registeredPrestos.Delete(request.NamespacedName)
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	_, ok := r.registeredPrestos.Load(request.NamespacedName)
	if !ok {
		r.log.Info("Registering for periodic events " + request.NamespacedName.String())
		r.registeredPrestos.Store(request.NamespacedName, event.GenericEvent{
			Meta:   presto.GetObjectMeta(),
			Object: presto,
		})
	}

	if len(presto.Status.Uuid) == 0 {
		clusterUUID := uuid.New().String()
		updated, err := r.updateStatus(presto,
			ctx, ClusterUpdateAction{
				clusterUUID: &clusterUUID,
			})
		if err != nil {
			r.log.Error(err, "failed to create and update cluster UUID")
			return reconcile.Result{}, err
		}
		if updated {
			// updated cluster ID. Requeue again so that cluster creation can begin
			return reconcile.Result{RequeueAfter: 0*time.Millisecond}, nil
		}
	}

	baseLabels := labels.Set{
		"clusterUUID": presto.Status.Uuid,
		"clusterName": presto.Name,
	}

	err, changesMade := r.headlessServiceConfig(presto, baseLabels, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changesMade {
		return reconcile.Result{}, nil
	}

	err, changesMade, _ = r.serviceConfig(presto, baseLabels, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changesMade {
		return reconcile.Result{}, nil
	}

	err, changesMade = r.coordinatorConfig(presto, baseLabels, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changesMade {
		return reconcile.Result{}, nil
	}

	err, changesMade = r.workerConfig(presto, baseLabels, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changesMade {
		return reconcile.Result{}, nil
	}

	err, changesMade = r.catalogConfig(presto, baseLabels, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changesMade {
		return reconcile.Result{}, nil
	}

	err, changesMade = r.coordinatorReplicaset(presto, baseLabels, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changesMade {
		return reconcile.Result{}, nil
	}

	err, changesMade, workerReplicaSet := r.workerReplicaset(presto, baseLabels, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changesMade {
		return reconcile.Result{}, nil
	}

	err, changesMade = r.hpaReplicaset(presto, baseLabels, ctx, workerReplicaSet)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changesMade {
		return reconcile.Result{}, nil
	}

	// Update the state based on coordinator pod phase
	_, coordinatorPodPhase := r.getCoordinatorPodPhase(presto, baseLabels)
	if coordinatorPodPhase == corev1.PodPending {
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			clusterState: falaricav1alpha1.ClusterPending,
			workerReplicaSet: workerReplicaSet,
		})
	} else if coordinatorPodPhase == corev1.PodFailed {
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			clusterState: falaricav1alpha1.ClusterFailedState,
			workerReplicaSet: workerReplicaSet,
		})
	}else if coordinatorPodPhase == corev1.PodRunning {
		workerCPU := fmt.Sprintf("%d%%",r.getCPUUsage(presto, false))
		coordinatorCPU := fmt.Sprintf("%d%%",r.getCPUUsage(presto, true))

		r.updateStatus(presto, ctx,ClusterUpdateAction{
			clusterState: falaricav1alpha1.ClusterReadyState,
			workerReplicaSet: workerReplicaSet,
			workerCPUUsage: &workerCPU,
			coordinatorCPUUsage: &coordinatorCPU,
		})
	} else {
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			clusterState: falaricav1alpha1.ClusterUnknown,
			workerReplicaSet: workerReplicaSet,
		})
	}
	return ctrl.Result{}, nil
}

func (r *ReconcilePresto) headlessServiceConfig(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string,
	ctx context.Context) (error, bool) {
	changesMade := false
	err, created := createHeadLessPodDiscoverySvc(presto, r, baseLabels)
	if err != nil {
		r.log.Error(err, "failed to create headless service for pods")
		r.eventRecorder.Eventf(presto, corev1.EventTypeWarning, "Failed",
			"Failed to create headless service for pods %s", err.Error())
		errorReason := fmt.Sprintf("Failed to create headless service for pods %s", err.Error())
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			errorReason: &errorReason,
			clusterState: falaricav1alpha1.ClusterFailedState,
		})
		return err, changesMade
	} else {
		if created{
			r.log.Info("created headless service ")
			r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Created",
				"Created Headless Service. %s", getPodDiscoveryServiceName(presto.Status.Uuid))
			changesMade = true
		}
		if created {
			headlessSvc := getPodDiscoveryServiceName(presto.Status.Uuid)
			r.updateStatus(presto, ctx,ClusterUpdateAction{
				headlessService: &headlessSvc,
				clusterState: falaricav1alpha1.ClusterPending,
			})
		}
	}
	return err, changesMade
}

func (r *ReconcilePresto) serviceConfig(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string,
	ctx context.Context) (error, bool, *corev1.Service) {
	changesMade := false
	err, service, created := createOrGetService(presto, r, baseLabels)
	if err != nil {
		r.log.Error(err, "failed to create service for pods")
		r.eventRecorder.Eventf(presto, corev1.EventTypeWarning, "Failed",
			"Failed to create service for pods %s", err.Error())
		errorReason := fmt.Sprintf("Failed to create service for pods %s", err.Error())
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			errorReason: &errorReason,
			clusterState: falaricav1alpha1.ClusterFailedState,
		})
		return err, changesMade, nil
	} else {
		if created{
			r.log.Info("created service ")
			r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Created",
				"Created Service. %s", getExternalServiceName(presto.Status.Uuid))
			changesMade = true
		}
		if created || len(presto.Status.CoordinatorAddress) == 0 {
			r.updateStatus(presto, ctx,ClusterUpdateAction{
				service: service,
				clusterState: falaricav1alpha1.ClusterPending,
			})
		}
	}
	return err, changesMade, service
}

func (r *ReconcilePresto) coordinatorConfig(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string,
	ctx context.Context) (error, bool) {
	created, err := createCoordinatorConfig(presto, r.client, baseLabels)
	if err != nil {
		r.log.Error(err, "failed to create coordinator config map")
		errorReason := fmt.Sprintf("Failed to create coordinator config map %s", err.Error())
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			errorReason: &errorReason,
			clusterState: falaricav1alpha1.ClusterFailedState,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeWarning, "Failed",
			"failed to create coordinator config map %s", err.Error())
		return err, false
	}
	if created {
		cm := getCoordinatorConfigMapName(presto.Status.Uuid)
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			coordinatorConfMap: &cm,
			clusterState: falaricav1alpha1.ClusterPending,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Created",
			"Created Coordinator Config. %s", cm)
		r.log.Info("created coordinator config map")
	}
	return nil, created
}
func (r *ReconcilePresto) workerConfig(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string,
	ctx context.Context) (error, bool) {
	created, err := createWorkerConfig(presto, r.client, baseLabels)
	if err != nil {
		r.log.Error(err, "failed to create worker config map")
		errorReason := fmt.Sprintf("Failed to create worker config map %s", err.Error())
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			errorReason: &errorReason,
			clusterState: falaricav1alpha1.ClusterFailedState,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeWarning, "Failed",
			"Failed to create worker config map %s", err.Error())
		return err, created
	}
	if created {
		wm := getWorkerConfigMapName(presto.Status.Uuid)
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			workerConfMap: &wm,
			clusterState: falaricav1alpha1.ClusterPending,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Created",
			"Created Worker Config. %s", wm)
		r.log.Info("created worker config map")
	}
	return nil, created
}
func (r *ReconcilePresto) catalogConfig(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string,
	ctx context.Context) (error, bool) {
	created, err := createCatalogConfig(presto, r, baseLabels)
	if err != nil {
		r.log.Error(err, "failed to create catalog config map")
		errorReason := fmt.Sprintf("Failed to create catalog config map %s", err.Error())
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			errorReason: &errorReason,
			clusterState: falaricav1alpha1.ClusterFailedState,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeWarning, "Failed",
			"Failed to create catalog config map %s", err.Error())
		return err, created
	}
	if created {
		cc := getCatalogConfigMapName(presto.Status.Uuid)
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			catalogConfMap: &cc,
			clusterState: falaricav1alpha1.ClusterPending,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Created",
			"Created Catalog Config. %s", cc)
		r.log.Info("catalog catalog config map")
		created = true
	}
	return nil, created
}
func (r *ReconcilePresto) coordinatorReplicaset(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string,
	ctx context.Context) (error, bool) {
	_, created, err := createUpdateReplicaSetForCoordinator(r, presto,
		getCoordinatorPodLabels(baseLabels, presto.Status.Uuid))
	if err != nil {
		r.log.Error(err, "failed to create/update coordinator replicaset ")
		errorReason := fmt.Sprintf("Failed to create coordinator replicaset %s", err.Error())
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			errorReason: &errorReason,
			clusterState: falaricav1alpha1.ClusterFailedState,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeWarning, "Failed",
			"Failed to create/update coordinator replicaset %s", err.Error())
		return err, created
	}
	if created {
		cr := getCoordinatorReplicaset(presto.Status.Uuid)
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			coordinatorReplicaSetName: &cr,
			clusterState: falaricav1alpha1.ClusterPending,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Created",
			"Created Coordinator Replicaset. %s", cr)
		r.log.Info("created coordinator replicaset")
	}
	return nil, created
}
func (r *ReconcilePresto) workerReplicaset(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string,
	ctx context.Context) (error, bool, *v1.ReplicaSet) {
	changesMade := false
	workerReplicaSet, created, updated, err := createUpdateReplicaSetForWorker(r, presto,
		getWorkerPodLabels(baseLabels, presto.Status.Uuid))
	if err != nil {
		errorReason := fmt.Sprintf("Failed to create worker replicaset %s", err.Error())
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			errorReason: &errorReason,
			clusterState: falaricav1alpha1.ClusterFailedState,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeWarning, "Failed",
			"Failed to create worker replicaset %s", err.Error())
		return err, changesMade, workerReplicaSet
	}
	if updated {
		r.log.Info("updated worker replicaset")
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Updated",
			"Updated Worker Replicaset. %s ", workerReplicaSet.Name)
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			workerReplicaSet: workerReplicaSet,
			clusterState: falaricav1alpha1.ClusterPending,
		})
		changesMade = true
	}
	if created {
		r.log.Info("created worker replicaset")
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			workerReplicaSet: workerReplicaSet,
			clusterState: falaricav1alpha1.ClusterPending,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Created",
			"Created Worker Replicaset. %s", workerReplicaSet.Name)
		changesMade = true
	}
	return nil, changesMade, workerReplicaSet
}

func (r *ReconcilePresto) hpaReplicaset(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string,
	ctx context.Context,
	workerReplicaSet *v1.ReplicaSet) (error, bool) {
	changesMade := false
	created, updated, deleted, err := handleReplicaSet(r, presto, workerReplicaSet, baseLabels, ctx)
	if err != nil {
		r.log.Error(err, "failed to create/update autoscale replicaset")
		errorReason := fmt.Sprintf("Failed to create autoscale config %s", err.Error())
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			errorReason: &errorReason,
			clusterState: falaricav1alpha1.ClusterFailedState,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeWarning, "Failed",
			"failed to create/update autoscale replicaset %s", err.Error())
		return  err, false
	}
	if created {
		hpaName := getHPAName(presto.Status.Uuid)
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			hpaName: &hpaName,
			clusterState: falaricav1alpha1.ClusterPending,
		})
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Created",
			"Created HPA. %s", hpaName)
		changesMade = true
	}
	if updated {
		hpaName := getHPAName(presto.Status.Uuid)
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Updated",
			"Updated HPA. %s", hpaName)
		changesMade = true
	}
	if deleted {
		hpaName := getHPAName(presto.Status.Uuid)
		hpa := ""
		r.eventRecorder.Eventf(presto, corev1.EventTypeNormal, "Deleted",
			"Deleted HPA. %s", hpaName)
		r.updateStatus(presto, ctx,ClusterUpdateAction{
			hpaName: &hpa,
			clusterState: falaricav1alpha1.ClusterPending,
		})
		changesMade = true
	}
	return nil, changesMade
}
type ClusterUpdateAction struct  {
	clusterUUID *string
	service *corev1.Service
	headlessService *string
	workerConfMap *string
	coordinatorConfMap *string
	catalogConfMap *string
	coordinatorReplicaSetName *string
	hpaName *string
	workerReplicaSet *v1.ReplicaSet
	clusterState falaricav1alpha1.ClusterState
	errorReason *string
	coordinatorCPUUsage *string
	workerCPUUsage *string
}

func (r *ReconcilePresto) updateStatus(presto *falaricav1alpha1.Presto,
	ctx context.Context, updateAction ClusterUpdateAction) (bool, error) {
	prestoCopy, err := r.getPresto(presto)
	if err != nil {
		r.log.Error(err, "failed to update the presto status")
		return false, err
	}
	if prestoCopy == nil {
		prestoCopy = presto
	}
	update := false
	// update UUID
	if updateAction.clusterUUID != nil{
		prestoCopy.Status.Uuid = *updateAction.clusterUUID
		update = true
	}
	if updateAction.service != nil{
		prestoCopy.Status.Service = updateAction.service.Name
		prestoCopy.Status.CoordinatorAddress = fmt.Sprintf("%s:%d/%d", updateAction.service.Spec.ClusterIP,
			updateAction.service.Spec.Ports[0].Port, updateAction.service.Spec.Ports[0].NodePort)
		update = true
	}
	if updateAction.headlessService != nil{
		prestoCopy.Status.HeadlessService = *updateAction.headlessService
		update = true
	}
	if updateAction.workerConfMap != nil{
		prestoCopy.Status.WorkerConfig = *updateAction.workerConfMap
		update = true
	}
	if updateAction.coordinatorConfMap != nil{
		prestoCopy.Status.CoordinatorConfig = *updateAction.coordinatorConfMap
		update = true
	}
	if updateAction.catalogConfMap != nil{
		prestoCopy.Status.CatalogConfig = *updateAction.catalogConfMap
		update = true
	}
	if updateAction.coordinatorReplicaSetName != nil{
		prestoCopy.Status.CoordinatorReplicaset = *updateAction.coordinatorReplicaSetName
		update = true
	}
	if updateAction.hpaName != nil{
		prestoCopy.Status.HpaName = *updateAction.hpaName
		update = true
	}
	if len(updateAction.clusterState) != 0 {
		prestoCopy.Status.ClusterState = updateAction.clusterState
		update = true
	}
	if updateAction.errorReason != nil{
		prestoCopy.Status.ErrorReason = *updateAction.errorReason
		update = true
	}
	if updateAction.workerCPUUsage != nil{
		prestoCopy.Status.WorkerCPU = *updateAction.workerCPUUsage
		update = true
	}
	if updateAction.coordinatorCPUUsage != nil{
		prestoCopy.Status.CoordinatorCPU = *updateAction.coordinatorCPUUsage
		update = true
	}
	// Update worker count
	if updateAction.workerReplicaSet != nil {
		prestoCopy.Status.WorkerReplicaset = updateAction.workerReplicaSet.Name
		prestoCopy.Status.DesiredWorkers = *updateAction.workerReplicaSet.Spec.Replicas
		prestoCopy.Status.CurrentWorkers = updateAction.workerReplicaSet.Status.AvailableReplicas
		update = true
	}

	if update {
		prestoCopy.Status.ModificationTime = metav1.Now()
		err = r.client.Status().Update(ctx, prestoCopy)
		if err != nil {
			r.log.Error(err, "failed to update the presto status")
			return false, err
		}
	}
	return update, nil
}

func (r *ReconcilePresto) getCPUUsage(presto *falaricav1alpha1.Presto,
	isCoordinator bool) int64 {
	var podLabels map[string]string
	var containerPrefix string
	var cpuLimit string
	if isCoordinator {
		wk, wv := getCoordinatorPodLabel(presto.Status.Uuid)
		containerPrefix = getCoordinatorContainerName(presto.Status.Uuid)
		podLabels = labels.Set{wk: wv}
		cpuLimit = presto.Spec.Coordinator.CpuLimit
	} else {
		wk, wv := getWorkerPodLabel(presto.Status.Uuid)
		containerPrefix = getWorkerContainerPrefix(presto.Status.Uuid)
		podLabels = labels.Set{wk: wv}
		cpuLimit = presto.Spec.Worker.CpuLimit
	}

	podsMetrics, err := r.metricsClient.PodMetricses(presto.Namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       presto.Kind,
			APIVersion: presto.APIVersion,
		},
		LabelSelector: labels.SelectorFromSet(podLabels).String(),
	})
	if err != nil {
		r.log.Error(err, "Failed to fetch CPU stats")
		return 0
	}
	// No need to check the error here as the CPULimit has already been parsed
	allottedQuantity, _ := resource.ParseQuantity(cpuLimit)
	allotted := allottedQuantity.MilliValue()
	var totalAllotted int64 = 0
	var totalUsed int64 = 0
	// get the cumulative CPU usage across all workers or for a coordinator
	for _, pm := range podsMetrics.Items {
		for _, cm := range pm.Containers {
			for k, used := range cm.Usage {
				if k == corev1.ResourceCPU &&
					cm.Name == containerPrefix {
					totalAllotted = totalAllotted + allotted
					totalUsed = totalUsed + used.MilliValue()
				}
			}
		}
	}
	if totalAllotted == 0 {
		return 0
	} else {
		return (totalUsed * 100) / totalAllotted
	}
}

func (r* ReconcilePresto) getCoordinatorPodPhase(presto *falaricav1alpha1.Presto,
	baseLabels map[string]string) (error, corev1.PodPhase) {
	coordinatorPodList := &corev1.PodList{}
	err := r.client.List(context.TODO(),
		coordinatorPodList,
		&client.ListOptions{
			Namespace:     presto.Namespace,
			LabelSelector: labels.SelectorFromSet(getCoordinatorPodLabels(baseLabels, presto.Status.Uuid)),
		})
	if err != nil {
		r.log.Error(err, "Failed to find the state of coordinator pod")
		return err, corev1.PodUnknown
	}
	if len(coordinatorPodList.Items) != 0 {
		return nil, coordinatorPodList.Items[0].Status.Phase
	}
	return nil, corev1.PodUnknown
}
func (r *ReconcilePresto) getPresto(oldPrestoObj *falaricav1alpha1.Presto) (*falaricav1alpha1.Presto, error) {
	prestoCluster := &falaricav1alpha1.PrestoList{}
	err := r.client.List(context.TODO(),
		prestoCluster,
		&client.ListOptions{
			Namespace:     oldPrestoObj.Namespace,
			LabelSelector: labels.SelectorFromSet(oldPrestoObj.Labels),
		})
	if err != nil {
		return nil, err
	}
	for _, pc := range prestoCluster.Items {
		if pc.Name == oldPrestoObj.Name {
			return &pc, nil
		}
	}
	return nil, nil
}