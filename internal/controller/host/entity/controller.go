package entity

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	hostv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/host/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/helper"
)

const (
	errNotHostEntity = "managed resource is not a HostEntity custom resource"
	errLookup        = "cannot lookup HostEntity from Dynatrace API"
)

// SetupGated adds a controller that reconciles HostEntity managed resources with SafeStart support.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupGatedManagedController(
		mgr,
		o,
		hostv1alpha1.HostEntityGroupVersionKind,
		hostv1alpha1.HostEntityGroupKind,
		&hostv1alpha1.HostEntity{},
		&hostv1alpha1.HostEntityList{},
		&helper.DynatraceConnector{
			Kube:                mgr.GetClient(),
			NewExternalClientFn: NewExternalClient,
		},
	)
}

// Setup adds a controller that reconciles HostEntity managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupManagedController(
		mgr,
		o,
		hostv1alpha1.HostEntityGroupVersionKind,
		hostv1alpha1.HostEntityGroupKind,
		&hostv1alpha1.HostEntity{},
		&hostv1alpha1.HostEntityList{},
		&helper.DynatraceConnector{
			Kube:                mgr.GetClient(),
			NewExternalClientFn: NewExternalClient,
		},
	)
}

// NewExternalClient creates a new ExternalClient for HostEntity.
func NewExternalClient(client dtclient.Client) managed.ExternalClient {
	return &external{client: client}
}

type external struct {
	client dtclient.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*hostv1alpha1.HostEntity)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotHostEntity)
	}

	if meta.WasDeleted(cr) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	entityType := "HOST"
	if cr.Spec.ForProvider.Type != nil && *cr.Spec.ForProvider.Type != "" {
		entityType = *cr.Spec.ForProvider.Type
	}
	entityName := ""
	if cr.Spec.ForProvider.Name != nil {
		entityName = *cr.Spec.ForProvider.Name
	}

	if entityName == "" {
		cr.SetConditions(xpv2.ReconcileError(errors.New("spec.forProvider.name is required")))
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	id, tags, count, err := e.client.LookupHostEntity(ctx, entityType, entityName)
	if err != nil {
		cr.SetConditions(xpv2.ReconcileError(err))
		return managed.ExternalObservation{}, errors.Wrap(err, errLookup)
	}

	if id != "" {
		cr.Status.AtProvider.EntityID = &id
		cr.Status.AtProvider.Tags = toHostEntityTags(tags)
		meta.SetExternalName(cr, id)
		if count > 1 {
			cr.SetConditions(xpv2.Available().WithMessage(fmt.Sprintf("Warning: Multiple entities (%d) matched query. Using first match: %s", count, id)))
		} else {
			cr.SetConditions(xpv2.Available())
		}
	} else {
		cr.Status.AtProvider.EntityID = nil
		cr.Status.AtProvider.Tags = nil
		cr.SetConditions(xpv2.Unavailable().WithMessage(fmt.Sprintf("Entity %s not found in Dynatrace yet", entityName)))
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}

func toHostEntityTags(tags []dtclient.HostEntityTagDto) []hostv1alpha1.HostEntityTag {
	if tags == nil {
		return nil
	}
	res := make([]hostv1alpha1.HostEntityTag, len(tags))
	for i, t := range tags {
		res[i] = hostv1alpha1.HostEntityTag{
			Context:              t.Context,
			Key:                  t.Key,
			Value:                t.Value,
			StringRepresentation: t.StringRepresentation,
		}
	}
	return res
}
