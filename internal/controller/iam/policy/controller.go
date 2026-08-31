package policy

import (
	"context"
	"slices"

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
	errNotPolicy    = "managed resource is not a Policy custom resource"
	errGetPolicy    = "cannot get Policy from Dynatrace API"
	errCreatePolicy = "cannot create Policy in Dynatrace API"
	errUpdatePolicy = "cannot update Policy in Dynatrace API"
	errDeletePolicy = "cannot delete Policy in Dynatrace API"
)

// SetupGated adds a controller that reconciles Policy managed resources with SafeStart support.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupGatedManagedController(
		mgr,
		o,
		iamv1alpha1.PolicyGroupVersionKind,
		iamv1alpha1.PolicyGroupKind,
		&iamv1alpha1.Policy{},
		&helper.DynatraceConnector{
			Kube: mgr.GetClient(),
			NewExternalClientFn: func(client dtclient.Client) managed.ExternalClient {
				return &external{client: client}
			},
		},
	)
}

// Setup adds a controller that reconciles Policy managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupManagedController(
		mgr,
		o,
		iamv1alpha1.PolicyGroupVersionKind,
		iamv1alpha1.PolicyGroupKind,
		&iamv1alpha1.Policy{},
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

func resolveLevel(p iamv1alpha1.PolicyParameters) (string, string) {
	if p.Environment != nil && *p.Environment != "" {
		return "environment", *p.Environment
	}
	if p.Account != nil && *p.Account != "" {
		return "account", *p.Account
	}
	return "account", ""
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*iamv1alpha1.Policy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPolicy)
	}

	lt, lid := resolveLevel(cr.Spec.ForProvider)
	uuid := meta.GetExternalName(cr)

	var pol *dtclient.PolicyDto
	var err error

	if uuid != "" && uuid != cr.GetName() {
		pol, err = e.client.GetPolicy(ctx, lt, lid, uuid)
	} else {
		list, listErr := e.client.ListPolicies(ctx, lt, lid)
		if listErr != nil {
			return managed.ExternalObservation{}, errors.Wrap(listErr, errGetPolicy)
		}
		for _, p := range list.Policies {
			if p.Name == cr.Spec.ForProvider.Name {
				pol = &p
				break
			}
		}
	}

	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPolicy)
	}

	if pol == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	meta.SetExternalName(cr, pol.UUID)
	cr.Status.AtProvider.ID = pol.UUID
	cr.Status.AtProvider.UUID = pol.UUID
	cr.Status.AtProvider.LevelType = lt
	cr.Status.AtProvider.LevelID = lid

	cr.SetConditions(xpv2.Available())

	upToDate := isPolicyUpToDate(cr.Spec.ForProvider, pol)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*iamv1alpha1.Policy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPolicy)
	}

	cr.SetConditions(xpv2.Creating())

	levelType, levelID := resolveLevel(cr.Spec.ForProvider)
	desc := ""
	if cr.Spec.ForProvider.Description != nil {
		desc = *cr.Spec.ForProvider.Description
	}

	created, err := e.client.CreatePolicy(ctx, levelType, levelID, dtclient.PolicyDto{
		Name:           cr.Spec.ForProvider.Name,
		Description:    desc,
		StatementQuery: cr.Spec.ForProvider.StatementQuery,
		Tags:           cr.Spec.ForProvider.Tags,
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePolicy)
	}

	meta.SetExternalName(cr, created.UUID)
	cr.Status.AtProvider.ID = created.UUID
	cr.Status.AtProvider.UUID = created.UUID
	cr.Status.AtProvider.LevelType = levelType
	cr.Status.AtProvider.LevelID = levelID

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*iamv1alpha1.Policy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPolicy)
	}

	uuid := meta.GetExternalName(cr)
	levelType, levelID := resolveLevel(cr.Spec.ForProvider)
	desc := ""
	if cr.Spec.ForProvider.Description != nil {
		desc = *cr.Spec.ForProvider.Description
	}

	_, err := e.client.UpdatePolicy(ctx, levelType, levelID, uuid, dtclient.PolicyDto{
		Name:           cr.Spec.ForProvider.Name,
		Description:    desc,
		StatementQuery: cr.Spec.ForProvider.StatementQuery,
		Tags:           cr.Spec.ForProvider.Tags,
	})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePolicy)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*iamv1alpha1.Policy)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotPolicy)
	}

	cr.SetConditions(xpv2.Deleting())

	uuid := meta.GetExternalName(cr)
	if uuid == "" {
		return managed.ExternalDelete{}, nil
	}

	levelType, levelID := resolveLevel(cr.Spec.ForProvider)
	err := e.client.DeletePolicy(ctx, levelType, levelID, uuid)
	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, errors.Wrap(err, errDeletePolicy)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}

func isPolicyUpToDate(p iamv1alpha1.PolicyParameters, pol *dtclient.PolicyDto) bool {
	if p.Name != pol.Name {
		return false
	}
	if p.Description != nil && *p.Description != pol.Description {
		return false
	}
	if p.StatementQuery != "" && pol.StatementQuery != "" && p.StatementQuery != pol.StatementQuery {
		return false
	}
	if !slices.Equal(p.Tags, pol.Tags) {
		return false
	}
	return true
}
