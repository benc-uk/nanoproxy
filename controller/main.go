// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoProxy ingress controller - main entrypoint and manager setup
// ----------------------------------------------------------------------------

package main

import (
	"log"
	"os"
	"strconv"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	log.Println("Starting ingress controller for nanoproxy")

	// A reasonable detection if we are running in a Kubernetes cluster
	inKube := true
	if _, err := os.Stat("/var/run/secrets/kubernetes.io"); os.IsNotExist(err) {
		inKube = false
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: !inKube,
	})))

	// Can listen on any port, but best not eh?
	portString := "9090"
	if os.Getenv("PORT") != "" {
		portString = os.Getenv("PORT")
	}

	port, _ := strconv.Atoi(portString)

	server := webhook.NewServer(webhook.Options{
		// I'm not convinced this actually works
		Port: port,
	})

	options := ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: "", // Disable health probe
		WebhookServer:          server,
		LeaderElection:         false,
		Metrics: metricsserver.Options{
			BindAddress: "0", // Disable metrics
		},
	}

	// The manager will setup the controller, handle elections
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// This is the controller that will reconcile ingress objects
	reconciler := &IngressReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	if err = reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Ingress")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
