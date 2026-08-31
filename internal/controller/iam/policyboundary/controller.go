package policyboundary

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
	errNotBoundary    = "managed resource is not a PolicyBoundary custom resource"
	errGetBoundary    = "cannot get PolicyBoundary from Dynatrace API"
	errCreateBoundary = "cannot create PolicyBoundary in Dynatrace API"
	errUpdateBoundary = "cannot update PolicyBoundary in Dynatrace API"
	errDeleteBoundary = "cannot delete PolicyBoundary in Dynatrace API"
)

// SetupGated adds a controller that reconciles PolicyBoundary managed resources with SafeStart support.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupGatedManagedController(
		mgr,
		o,
		iamv1alpha1.PolicyBoundaryGroupVersionKind,
		iamv1alpha1.PolicyBoundaryGroupKind,
		&iamv1alpha1.PolicyBoundary{},
		&helper.DynatraceConnector{
			Kube: mgr.GetClient(),
			NewExternalClientFn: func(client dtclient.Client) managed.ExternalClient {
				return &external{client: client}
			},
		},
	)
}

// Setup adds a controller that reconciles PolicyBoundary managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupManagedController(
		mgr,
		o,
		iamv1alpha1.PolicyBoundaryGroupVersionKind,
		iamv1alpha1.PolicyBoundaryGroupKind,
		&iamv1alpha1.PolicyBoundary{},
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

func resolveLevel(p iamv1alpha1.PolicyBoundaryParameters) (string, string) {
	if p.Environment != nil && *p.Environment != "" {
		return "environment", *p.Environment
	}
	if p.Account != nil && *p.Account != "" {
		return "account", *p.Account
	}
	return "account", ""
}

func (e *external) findBoundary(ctx context.Context, lt, lid, uuid, name string) (*dtclient.PolicyBoundaryDto, error) {
	if uuid != "" {
		bnd, err := e.client.GetBoundary(ctx, lt, lid, uuid)
		if err != nil {
			if dtclient.IsNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		return bnd, nil
	}

	list, err := e.client.ListBoundaries(ctx, lt, lid)
	if err != nil {
		if dtclient.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, b := range list.Boundaries {
		if b.Name == name {
			return &b, nil
		}
	}
	return nil, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBoundary)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBoundary)
	}

	lt, lid := resolveLevel(cr.Spec.ForProvider)
	uuid := meta.GetExternalName(cr)
	if uuid == cr.GetName() {
		uuid = ""
	}

	bnd, err := e.findBoundary(ctx, lt, lid, uuid, cr.Spec.ForProvider.Name)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetBoundary)
	}
	if bnd == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	meta.SetExternalName(cr, bnd.UUID)
	cr.Status.AtProvider.ID = bnd.UUID
	cr.Status.AtProvider.UUID = bnd.UUID
	cr.Status.AtProvider.LevelType = lt
	cr.Status.AtProvider.LevelID = lid

	cr.SetConditions(xpv2.Available())

	upToDate := cr.Spec.ForProvider.Name == bnd.Name &&
		(cr.Spec.ForProvider.Query == "" || bnd.BoundaryQuery == "" || cr.Spec.ForProvider.Query == bnd.BoundaryQuery)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBoundary)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBoundary)
	}

	cr.SetConditions(xpv2.Creating())

	lt, lid := resolveLevel(cr.Spec.ForProvider)

	created, err := e.client.CreateBoundary(ctx, lt, lid, dtclient.PolicyBoundaryDto{
		Name:          cr.Spec.ForProvider.Name,
		BoundaryQuery: cr.Spec.ForProvider.Query,
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateBoundary)
	}

	meta.SetExternalName(cr, created.UUID)
	cr.Status.AtProvider.ID = created.UUID
	cr.Status.AtProvider.UUID = created.UUID
	cr.Status.AtProvider.LevelType = lt
	cr.Status.AtProvider.LevelID = lid

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBoundary)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBoundary)
	}

	lt, lid := resolveLevel(cr.Spec.ForProvider)
	uuid := meta.GetExternalName(cr)

	_, err := e.client.UpdateBoundary(ctx, lt, lid, uuid, dtclient.PolicyBoundaryDto{
		Name:          cr.Spec.ForProvider.Name,
		BoundaryQuery: cr.Spec.ForProvider.Query,
	})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateBoundary)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBoundary)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotBoundary)
	}

	cr.SetConditions(xpv2.Deleting())

	lt, lid := resolveLevel(cr.Spec.ForProvider)
	uuid := meta.GetExternalName(cr)
	if uuid == "" {
		return managed.ExternalDelete{}, nil
	}

	err := e.client.DeleteBoundary(ctx, lt, lid, uuid)
	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteBoundary)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
