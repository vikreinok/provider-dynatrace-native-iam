package group

import (
	"context"
	"slices"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/iam/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

const (
	errNotGroup    = "managed resource is not a Group custom resource"
	errTrackUsage  = "cannot track ProviderConfig usage"
	errGetPC       = "cannot get ProviderConfig"
	errNewClient   = "cannot create new Dynatrace client"
	errGetGroup    = "cannot get Group from Dynatrace API"
	errCreateGroup = "cannot create Group in Dynatrace API"
	errUpdateGroup = "cannot update Group in Dynatrace API"
	errDeleteGroup = "cannot delete Group in Dynatrace API"
)

// SetupGated adds a controller that reconciles Group managed resources with SafeStart support.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	if o.Gate == nil {
		return Setup(mgr, o)
	}
	o.Gate.Register(func() {
		if err := Setup(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", iamv1alpha1.GroupGroupVersionKind.String())
		}
	}, iamv1alpha1.GroupGroupVersionKind)
	return nil
}

// Setup adds a controller that reconciles Group managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(iamv1alpha1.GroupGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(iamv1alpha1.GroupGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube: mgr.GetClient(),
		}),
		managed.WithCreationGracePeriod(1*time.Nanosecond),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))), //nolint:staticcheck // Deprecated in controller-runtime but required by crossplane-runtime v2 APIRecorder
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&iamv1alpha1.Group{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*iamv1alpha1.Group)
	if !ok {
		return nil, errors.New(errNotGroup)
	}

	dt, err := dtclient.GetClientFromProviderConfig(ctx, c.kube, cr.GetProviderConfigReference())
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{client: dt}, nil
}

type external struct {
	client dtclient.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*iamv1alpha1.Group)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGroup)
	}

	uuid := meta.GetExternalName(cr)
	var grp *dtclient.GroupDto
	var err error

	if uuid != "" && uuid != cr.GetName() {
		grp, err = e.client.GetGroup(ctx, uuid)
	} else {
		// Attempt to match by name
		list, listErr := e.client.ListGroups(ctx)
		if listErr != nil {
			return managed.ExternalObservation{}, errors.Wrap(listErr, errGetGroup)
		}
		for _, g := range list.Items {
			if g.Name == cr.Spec.ForProvider.Name {
				grp = &g
				break
			}
		}
	}

	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetGroup)
	}

	if grp == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	meta.SetExternalName(cr, grp.UUID)
	cr.Status.AtProvider.ID = grp.UUID
	cr.Status.AtProvider.Owner = grp.Owner
	cr.Status.AtProvider.Hidden = grp.Hidden
	cr.Status.AtProvider.CreatedAt = grp.CreatedAt
	cr.Status.AtProvider.UpdatedAt = grp.UpdatedAt

	cr.SetConditions(xpv2.Available())

	upToDate := isGroupUpToDate(cr.Spec.ForProvider, grp)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*iamv1alpha1.Group)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGroup)
	}

	cr.SetConditions(xpv2.Creating())

	desc := ""
	if cr.Spec.ForProvider.Description != nil {
		desc = *cr.Spec.ForProvider.Description
	}

	created, err := e.client.CreateGroup(ctx, dtclient.GroupDto{
		Name:                     cr.Spec.ForProvider.Name,
		Description:              desc,
		FederatedAttributeValues: cr.Spec.ForProvider.FederatedAttributeValues,
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateGroup)
	}

	meta.SetExternalName(cr, created.UUID)
	cr.Status.AtProvider.ID = created.UUID
	cr.Status.AtProvider.Owner = created.Owner
	cr.Status.AtProvider.Hidden = created.Hidden
	cr.Status.AtProvider.CreatedAt = created.CreatedAt
	cr.Status.AtProvider.UpdatedAt = created.UpdatedAt

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*iamv1alpha1.Group)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotGroup)
	}

	uuid := meta.GetExternalName(cr)
	desc := ""
	if cr.Spec.ForProvider.Description != nil {
		desc = *cr.Spec.ForProvider.Description
	}

	err := e.client.UpdateGroup(ctx, uuid, dtclient.GroupDto{
		Name:                     cr.Spec.ForProvider.Name,
		Description:              desc,
		FederatedAttributeValues: cr.Spec.ForProvider.FederatedAttributeValues,
	})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateGroup)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*iamv1alpha1.Group)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotGroup)
	}

	cr.SetConditions(xpv2.Deleting())

	uuid := meta.GetExternalName(cr)
	if uuid == "" {
		return managed.ExternalDelete{}, nil
	}

	err := e.client.DeleteGroup(ctx, uuid)
	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteGroup)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}

func isGroupUpToDate(p iamv1alpha1.GroupParameters, g *dtclient.GroupDto) bool {
	if p.Name != g.Name {
		return false
	}
	if p.Description != nil && *p.Description != g.Description {
		return false
	}
	if !slices.Equal(p.FederatedAttributeValues, g.FederatedAttributeValues) {
		return false
	}
	return true
}
