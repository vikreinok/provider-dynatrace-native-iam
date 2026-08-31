package serviceuser

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/iam/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/helper"
)

const (
	errNotServiceUser    = "managed resource is not a ServiceUser custom resource"
	errGetServiceUser    = "cannot get ServiceUser from Dynatrace API"
	errCreateServiceUser = "cannot create ServiceUser in Dynatrace API"
	errDeleteServiceUser = "cannot delete ServiceUser in Dynatrace API"
)

// SetupGated adds a controller that reconciles ServiceUser managed resources with SafeStart support.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupGatedManagedController(
		mgr,
		o,
		iamv1alpha1.ServiceUserGroupVersionKind,
		iamv1alpha1.ServiceUserGroupKind,
		&iamv1alpha1.ServiceUser{},
		&helper.DynatraceConnector{
			Kube: mgr.GetClient(),
			NewExternalClientFn: func(client dtclient.Client) managed.ExternalClient {
				return &external{kube: mgr.GetClient(), client: client}
			},
		},
	)
}

// Setup adds a controller that reconciles ServiceUser managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupManagedController(
		mgr,
		o,
		iamv1alpha1.ServiceUserGroupVersionKind,
		iamv1alpha1.ServiceUserGroupKind,
		&iamv1alpha1.ServiceUser{},
		&helper.DynatraceConnector{
			Kube: mgr.GetClient(),
			NewExternalClientFn: func(client dtclient.Client) managed.ExternalClient {
				return &external{kube: mgr.GetClient(), client: client}
			},
		},
	)
}

type external struct {
	kube   client.Client
	client dtclient.Client
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

func (e *external) resolveGroups(ctx context.Context, groupIDs []string) ([]string, error) {
	if len(groupIDs) == 0 {
		return nil, nil
	}
	resolved := make([]string, len(groupIDs))
	for i, g := range groupIDs {
		if isUUID(g) {
			resolved[i] = g
			continue
		}
		if e.kube != nil {
			grp := &iamv1alpha1.Group{}
			if err := e.kube.Get(ctx, types.NamespacedName{Name: g}, grp); err != nil {
				return nil, errors.Wrapf(err, "cannot get referenced Group %q", g)
			}
			uuid := grp.Status.AtProvider.ID
			if uuid == "" {
				uuid = meta.GetExternalName(grp)
			}
			if uuid == "" || uuid == grp.GetName() || !isUUID(uuid) {
				return nil, fmt.Errorf("referenced Group %q is not ready yet (waiting for status.atProvider.id)", g)
			}
			resolved[i] = uuid
		} else {
			return nil, fmt.Errorf("group ID %q is not a valid UUID", g)
		}
	}
	return resolved, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*iamv1alpha1.ServiceUser)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotServiceUser)
	}

	uuid := meta.GetExternalName(cr)
	var su *dtclient.ServiceUserDto
	var err error

	if uuid != "" && uuid != cr.GetName() {
		su, err = e.client.GetServiceUser(ctx, uuid)
	} else {
		list, listErr := e.client.ListServiceUsers(ctx)
		if listErr != nil {
			return managed.ExternalObservation{}, errors.Wrap(listErr, errGetServiceUser)
		}
		for _, u := range list.Items {
			if u.Name == cr.Spec.ForProvider.Name {
				su = &u
				break
			}
		}
	}

	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetServiceUser)
	}

	if su == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	meta.SetExternalName(cr, su.UID)
	cr.Status.AtProvider.UID = su.UID
	cr.Status.AtProvider.Email = su.Email

	cr.SetConditions(xpv2.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*iamv1alpha1.ServiceUser)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotServiceUser)
	}

	cr.SetConditions(xpv2.Creating())

	desc := ""
	if cr.Spec.ForProvider.Description != nil {
		desc = *cr.Spec.ForProvider.Description
	}

	groups, err := e.resolveGroups(ctx, cr.Spec.ForProvider.Groups)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	created, err := e.client.CreateServiceUser(ctx, dtclient.ServiceUserDto{
		Name:        cr.Spec.ForProvider.Name,
		Description: desc,
		Groups:      groups,
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateServiceUser)
	}

	meta.SetExternalName(cr, created.UID)
	cr.Status.AtProvider.UID = created.UID
	cr.Status.AtProvider.Email = created.Email

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*iamv1alpha1.ServiceUser)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotServiceUser)
	}

	cr.SetConditions(xpv2.Deleting())

	uuid := meta.GetExternalName(cr)
	if uuid == "" {
		return managed.ExternalDelete{}, nil
	}

	err := e.client.DeleteServiceUser(ctx, uuid)
	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteServiceUser)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
