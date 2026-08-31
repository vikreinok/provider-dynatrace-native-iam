/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helper

import (
	"context"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

const (
	errNewClient = "cannot create new Dynatrace client"
)

// DynatraceConnector connects a managed resource to the Dynatrace API using credentials from its ProviderConfig.
type DynatraceConnector struct {
	Kube                client.Client
	NewExternalClientFn func(client dtclient.Client) managed.ExternalClient
}

// Connect implements managed.ExternalConnector.
func (c *DynatraceConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	pcr, ok := mg.(resource.ProviderConfigReferencer)
	if !ok {
		return nil, errors.New("managed resource does not implement ProviderConfigReferencer")
	}

	dt, err := dtclient.GetClientFromProviderConfig(ctx, c.Kube, pcr.GetProviderConfigReference())
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return c.NewExternalClientFn(dt), nil
}

// SetupManagedController registers and configures a managed reconciler with standard options and feature gates.
func SetupManagedController(
	mgr ctrl.Manager,
	o controller.Options,
	gvk schema.GroupVersionKind,
	groupKind string,
	forObject client.Object,
	externalConnector managed.ExternalConnector,
) error {
	name := managed.ControllerName(groupKind)

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnector(externalConnector),
		managed.WithCreationGracePeriod(1 * time.Nanosecond),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		opts = append(opts, managed.WithManagementPolicies())
	}

	if o.Features.Enabled(feature.EnableAlphaChangeLogs) {
		opts = append(opts, managed.WithChangeLogger(o.ChangeLogOptions.ChangeLogger))
	}

	if o.MetricOptions != nil {
		opts = append(opts, managed.WithMetricRecorder(o.MetricOptions.MRMetrics))
	}

	r := managed.NewReconciler(mgr, resource.ManagedKind(gvk), opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(forObject).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// SetupGatedManagedController wraps SetupManagedController with safe-start gate support.
func SetupGatedManagedController(
	mgr ctrl.Manager,
	o controller.Options,
	gvk schema.GroupVersionKind,
	groupKind string,
	forObject client.Object,
	externalConnector managed.ExternalConnector,
) error {
	if o.Gate == nil {
		return SetupManagedController(mgr, o, gvk, groupKind, forObject, externalConnector)
	}
	o.Gate.Register(func() {
		if err := SetupManagedController(mgr, o, gvk, groupKind, forObject, externalConnector); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", gvk.String())
		}
	}, gvk)
	return nil
}
