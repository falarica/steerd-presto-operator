package main

import (
	"flag"
	"fmt"
	"github.com/falarica/steerd-presto-operator/pkg/apis"
	"github.com/falarica/steerd-presto-operator/pkg/controller"
	"github.com/falarica/steerd-presto-operator/pkg/controller/presto"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	"os"
	"runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = apimachineryruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func printVersion() {
	setupLog.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}
func main() {
	var cmdLineParams = presto.CommandLineParams{}
	flag.IntVar(&cmdLineParams.StatusUpdateInterval,
		"status-update-interval", 10, "Presto status update interval.")

	flag.Parse()
	printVersion()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	metricsClient, err := metrics.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		setupLog.Error(err, "Could not create metrics client")
		os.Exit(1)
	}

	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "")
		os.Exit(1)
	}

	if err = controller.AddToManager(
		mgr, metricsClient, cmdLineParams,
		ctrl.Log.WithName("controllers").WithName("Deployment")); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Deployment")
		os.Exit(1)
	}

	//if err = (&v1alpha1.Presto{}).SetupWebhookWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create webhook", "webhook", "Presto")
	//	os.Exit(1)
	//}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
