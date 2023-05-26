// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoProxy ingress controller - controller logic
// ----------------------------------------------------------------------------

package main

import (
	"context"
	"strconv"

	"github.com/benc-uk/nanoproxy/pkg/config"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	// A map of all ingresses we are watching
	ingressCache = make(map[string]*netv1.Ingress)
)

// Reconcile is part of the main kubernetes reconciliation loop
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	key := req.Namespace + ":" + req.Name

	var ingress netv1.Ingress
	if err := r.Get(ctx, req.NamespacedName, &ingress); err != nil {
		if apierrors.IsNotFound(err) {
			// Handle delete
			log.Info("Ingress deleted", "key", key)
			delete(ingressCache, key)
			buildConfig()

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Handle new / update
	log.Info("Ingress updated or created", "key", key)
	ingressCache[key] = &ingress
	buildConfig()

	return ctrl.Result{}, nil
}

func buildConfig() {
	conf := config.Config{}
	upstreamMap := make(map[string]config.Upstream)

	// Loop over all ingresses and build up config
	for _, i := range ingressCache {
		// Check for annotation to force https
		scheme := "http"
		annotations := i.GetAnnotations()
		if annotations != nil && annotations["nanoproxy/backend-protocol"] == "https" {
			scheme = "https"
		}

		for _, rule := range i.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				svcName := path.Backend.Service.Name
				svcPort := path.Backend.Service.Port.Number
				pathString := path.Path
				if pathString == "" {
					pathString = "/"
				}

				upstreamMap[svcName+"-"+strconv.Itoa(int(svcPort))] = config.Upstream{
					Name:   svcName + "-" + strconv.Itoa(int(svcPort)),
					Host:   svcName,
					Port:   int(svcPort),
					Scheme: scheme,
				}

				matchMode := "prefix"
				if path.PathType != nil && *path.PathType == netv1.PathTypeExact {
					matchMode = "exact"
				}

				conf.Rules = append(conf.Rules, config.Rule{
					Path:      pathString,
					Upstream:  svcName,
					MatchMode: matchMode,
					StripPath: false,
					Host:      rule.Host,
				})
			}
		}
	}

	// Convert map of upstreams array in config
	for _, u := range upstreamMap {
		conf.Upstreams = append(conf.Upstreams, u)
	}

	// We overwrite the config file each time, this is fine
	conf.Write()
}

// Register controller for: networking v1 Ingress resources
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&netv1.Ingress{}).Complete(r)
}
