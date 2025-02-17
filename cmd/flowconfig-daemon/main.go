// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2024 Intel Corporation

package main

import (
	"flag"
	"os"

	"github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/pkg/flowconfig/flowsets"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	flowconfigv1 "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/apis/flowconfig/v1"
	flowconfigctlr "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/controllers/flowconfig"
)

const (
	defaultSysFs = "/sys"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(flowconfigv1.AddToScheme(scheme))
}

func main() {
	var sysFs string
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&sysFs, "set-sysfs", defaultSysFs, "Set alternative sysfs. Default is: /sys")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	namespace := os.Getenv("OPERATOR_NAMESPACE")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		NewCache: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
			opts.DefaultNamespaces = map[string]cache.Config{
				namespace: {},
			}
			return cache.New(config, opts)
		},
		Scheme:         scheme,
		LeaderElection: enableLeaderElection,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		setupLog.Error(err, "unable to get K8s node name from ENV var NODE_NAME")
		os.Exit(1)
	}

	fs := flowsets.NewFlowSets()
	fc := flowconfigctlr.GetDCFClient()

	setupLog.Info("checking for UFT grpc client readiness...")
	err = flowconfigctlr.CheckClientReadiness()
	if err != nil {
		setupLog.Error(err, "unable to create UFT grpc client connection")
		os.Exit(1)
	}
	setupLog.Info("UFT grpc client is ready")

	setupLog.Info("using sysfs", "sysFs", sysFs)
	flowRc := flowconfigctlr.GetNodeFlowConfigReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("NodeAclPolicy"),
		mgr.GetScheme(),
		fs,
		fc,
		nodeName,
		sysFs,
	)

	if err = flowRc.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeAclPolicy")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
