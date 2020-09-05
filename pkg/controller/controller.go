package controller

import (
	"github.com/falarica/steerd-presto-operator/pkg/controller/presto"
	"github.com/go-logr/logr"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, *metrics.MetricsV1beta1Client,
	presto.CommandLineParams, logr.Logger) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, metricsClient *metrics.MetricsV1beta1Client,
	cmdLineParams presto.CommandLineParams, log logr.Logger) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, metricsClient, cmdLineParams, log); err != nil {
			return err
		}
	}
	return nil
}
