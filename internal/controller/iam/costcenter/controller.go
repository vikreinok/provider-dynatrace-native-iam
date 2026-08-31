package costcenter

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	iamv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/iam/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/helper"
)

const (
	errNotCostCenter    = "managed resource is not a CostCenter custom resource"
	errGetCostCenter    = "cannot get CostCenter from Dynatrace API"
	errCreateCostCenter = "cannot create CostCenter in Dynatrace API"
	errDeleteCostCenter = "cannot delete CostCenter in Dynatrace API"
)

// SetupGated adds a controller that reconciles CostCenter managed resources with SafeStart support.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupGatedManagedController(
		mgr,
		o,
		iamv1alpha1.CostCenterGroupVersionKind,
		iamv1alpha1.CostCenterGroupKind,
		&iamv1alpha1.CostCenter{},
		&helper.DynatraceConnector{
			Kube: mgr.GetClient(),
			NewExternalClientFn: func(client dtclient.Client) managed.ExternalClient {
				return &external{client: client}
			},
		},
	)
}

// Setup adds a controller that reconciles CostCenter managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupManagedController(
		mgr,
		o,
		iamv1alpha1.CostCenterGroupVersionKind,
		iamv1alpha1.CostCenterGroupKind,
		&iamv1alpha1.CostCenter{},
		&helper.DynatraceConnector{
			Kube: mgr.GetClient(),
			NewExternalClientFn: func(client dtclient.Client) managed.ExternalClient {
				return &external{client: client}
			},
		},
	)
}

type external struct {
	client dtclient.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*iamv1alpha1.CostCenter)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCostCenter)
	}

	key := cr.Spec.ForProvider.CostCenter
	if key == "" {
		key = meta.GetExternalName(cr)
	}
	if key == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cc, err := e.client.GetCostCenter(ctx, key)
	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetCostCenter)
	}

	if cc == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	meta.SetExternalName(cr, cc.Key)
	cr.Status.AtProvider.ID = cc.Key
	cr.Status.AtProvider.CostCenter = cc.Key

	cr.SetConditions(xpv2.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*iamv1alpha1.CostCenter)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCostCenter)
	}

	cr.SetConditions(xpv2.Creating())

	key := cr.Spec.ForProvider.CostCenter
	if err := e.client.AddCostCenter(ctx, key); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateCostCenter)
	}

	meta.SetExternalName(cr, key)
	cr.Status.AtProvider.ID = key
	cr.Status.AtProvider.CostCenter = key

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*iamv1alpha1.CostCenter)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotCostCenter)
	}

	cr.SetConditions(xpv2.Deleting())

	key := meta.GetExternalName(cr)
	if key == "" {
		key = cr.Spec.ForProvider.CostCenter
	}
	if key == "" {
		return managed.ExternalDelete{}, nil
	}

	err := e.client.DeleteCostCenter(ctx, key)
	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteCostCenter)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
