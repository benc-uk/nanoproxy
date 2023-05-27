// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoProxy ingress controller - controller logic
// ----------------------------------------------------------------------------

package main

import (
	"context"
	"strconv"

	"github.com/benc-uk/nanoproxy/pkg/config"
	"github.com/go-logr/logr"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Fixed name
const ingressClassControllerName = "benc-uk/nanoproxy"

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	// A map of all ingresses we are watching
	ingressCache = make(map[string]*netv1.Ingress)
	logger       logr.Logger
)

// Reconcile is part of the main kubernetes reconciliation loop
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger = ctrllog.FromContext(ctx)

	key := req.Namespace + ":" + req.Name

	// Fetch ingress
	var ingress netv1.Ingress
	err := r.Get(ctx, req.NamespacedName, &ingress)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if ingressCache[key] != nil {
				// Handle delete
				logger.Info("Ingress deleted", "key", key)
				delete(ingressCache, key)
				buildConfig()
			}

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Check ingress class
	if ingress.Spec.IngressClassName == nil {
		// TODO: Not sure if this is the right behavior, but it makes life easier
		logger.Info("Ignoring due to missing ingressClassName", "key", key)

		// If we were previously tracking this ingress, remove it
		if ingressCache[key] != nil {
			logger.Info("Ingress deleted", "key", key)
			delete(ingressCache, key)
			buildConfig()
		}

		return ctrl.Result{}, nil
	}

	// Fetch ingress classes matching the name in the spec
	var ingressClass netv1.IngressClass
	err = r.Get(ctx, client.ObjectKey{Name: *ingress.Spec.IngressClassName}, &ingressClass)
	if err != nil {
		logger.Error(err, "Failed to get ingress class", "key", key)

		// If we were previously tracking this ingress, remove it
		if ingressCache[key] != nil {
			logger.Info("Ingress deleted", "key", key)
			delete(ingressCache, key)
			buildConfig()
		}

		return ctrl.Result{}, nil
	}

	// Finally check the controller name referenced in the IngressClass matches us
	if ingressClass.Spec.Controller != ingressClassControllerName {
		// Skip
		return ctrl.Result{}, nil
	}

	// If we got here, we are tracking this ingress and should update our cache
	logger.Info("Ingress updated or created", "key", key)
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
	err := conf.Write()
	if err != nil {
		logger.Error(err, "Failed to write config file")
	}
}

// Register controller for: networking v1 Ingress resources
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&netv1.Ingress{}).Complete(r)
}
