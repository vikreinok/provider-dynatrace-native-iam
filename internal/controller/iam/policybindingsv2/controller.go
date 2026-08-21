package policybindingsv2

import (
	"context"
	"fmt"
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
	errNotBindings    = "managed resource is not a PolicyBindingsV2 custom resource"
	errTrackUsage     = "cannot track ProviderConfig usage"
	errNewClient      = "cannot create new Dynatrace client"
	errGetBindings    = "cannot get PolicyBindings from Dynatrace API"
	errSetBindings    = "cannot set PolicyBindings in Dynatrace API"
	errDeleteBindings = "cannot delete PolicyBindings in Dynatrace API"
	errMissingGroup   = "missing group UUID in spec.forProvider.group"
)

// SetupGated adds a controller that reconciles PolicyBindingsV2 managed resources.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(iamv1alpha1.PolicyBindingsV2GroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(iamv1alpha1.PolicyBindingsV2GroupVersionKind),
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
		For(&iamv1alpha1.PolicyBindingsV2{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBindingsV2)
	if !ok {
		return nil, errors.New(errNotBindings)
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

func resolveLevel(p iamv1alpha1.PolicyBindingsV2Parameters) (string, string) {
	if p.Environment != nil && *p.Environment != "" {
		return "environment", *p.Environment
	}
	if p.Account != nil && *p.Account != "" {
		return "account", *p.Account
	}
	return "account", ""
}

func resolveGroup(p iamv1alpha1.PolicyBindingsV2Parameters) (string, error) {
	if p.Group != nil && *p.Group != "" {
		return *p.Group, nil
	}
	return "", errors.New(errMissingGroup)
}

func mapBoundPolicies(bindings *dtclient.PolicyBindingsDto) map[string]dtclient.BindingItem {
	boundMap := make(map[string]dtclient.BindingItem)
	if bindings == nil {
		return boundMap
	}
	for _, b := range bindings.PolicyBindings {
		boundMap[b.ID] = b
	}
	for _, uuid := range bindings.PolicyUUIDs {
		if _, ok := boundMap[uuid]; !ok {
			boundMap[uuid] = dtclient.BindingItem{ID: uuid}
		}
	}
	return boundMap
}

func checkBindingsUpToDate(desiredPolicies []iamv1alpha1.PolicyBindingItem, bindings *dtclient.PolicyBindingsDto) (found bool, upToDate bool) {
	if bindings == nil || (len(bindings.PolicyBindings) == 0 && len(bindings.PolicyUUIDs) == 0) {
		return false, false
	}

	boundMap := mapBoundPolicies(bindings)
	upToDate = true
	for _, desired := range desiredPolicies {
		existing, ok := boundMap[desired.ID]
		if !ok {
			return false, false
		}
		if len(desired.Boundaries) > 0 && !slices.Equal(desired.Boundaries, existing.Boundaries) {
			upToDate = false
		}
	}
	return true, upToDate
}

func (e *external) handleMissingBindings(cr *iamv1alpha1.PolicyBindingsV2, groupUUID, lt, lid string) (managed.ExternalObservation, bool) {
	if cr.GetDeletionTimestamp() == nil && meta.GetExternalName(cr) != "" && meta.GetExternalName(cr) != cr.GetName() {
		externalID := fmt.Sprintf("%s#%s#%s", groupUUID, lt, lid)
		cr.Status.AtProvider.ID = externalID
		cr.Status.AtProvider.Group = groupUUID
		cr.SetConditions(xpv2.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, true
	}
	return managed.ExternalObservation{ResourceExists: false}, false
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBindingsV2)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBindings)
	}

	groupUUID, err := resolveGroup(cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	lt, lid := resolveLevel(cr.Spec.ForProvider)
	bindings, err := e.client.GetPolicyBindingsForGroup(ctx, lt, lid, groupUUID)
	if err != nil {
		if dtclient.IsNotFound(err) {
			obs, ok := e.handleMissingBindings(cr, groupUUID, lt, lid)
			if ok {
				return obs, nil
			}
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetBindings)
	}

	allFound, upToDate := checkBindingsUpToDate(cr.Spec.ForProvider.Policy, bindings)
	if !allFound {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	externalID := fmt.Sprintf("%s#%s#%s", groupUUID, lt, lid)
	meta.SetExternalName(cr, externalID)
	cr.Status.AtProvider.ID = externalID
	cr.Status.AtProvider.Group = groupUUID
	cr.Status.AtProvider.Environment = ""
	if lt == "environment" {
		cr.Status.AtProvider.Environment = lid
	}
	if lt == "account" {
		cr.Status.AtProvider.Account = lid
	}

	cr.SetConditions(xpv2.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBindingsV2)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBindings)
	}

	cr.SetConditions(xpv2.Creating())

	groupUUID, err := resolveGroup(cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	lt, lid := resolveLevel(cr.Spec.ForProvider)

	for _, pol := range cr.Spec.ForProvider.Policy {
		err := e.client.SetPolicyBinding(ctx, lt, lid, pol.ID, groupUUID, dtclient.AppendLevelPolicyBindingForGroupDto{
			Parameters: pol.Parameters,
			Metadata:   pol.Metadata,
			Boundaries: pol.Boundaries,
		})
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errSetBindings)
		}
	}

	externalID := fmt.Sprintf("%s#%s#%s", groupUUID, lt, lid)
	meta.SetExternalName(cr, externalID)
	cr.Status.AtProvider.ID = externalID
	cr.Status.AtProvider.Group = groupUUID

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBindingsV2)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBindings)
	}

	groupUUID, err := resolveGroup(cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	lt, lid := resolveLevel(cr.Spec.ForProvider)

	for _, pol := range cr.Spec.ForProvider.Policy {
		err := e.client.SetPolicyBinding(ctx, lt, lid, pol.ID, groupUUID, dtclient.AppendLevelPolicyBindingForGroupDto{
			Parameters: pol.Parameters,
			Metadata:   pol.Metadata,
			Boundaries: pol.Boundaries,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errSetBindings)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*iamv1alpha1.PolicyBindingsV2)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotBindings)
	}

	cr.SetConditions(xpv2.Deleting())

	groupUUID, err := resolveGroup(cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot resolve group for policy binding deletion")
	}

	lt, lid := resolveLevel(cr.Spec.ForProvider)

	for _, pol := range cr.Spec.ForProvider.Policy {
		_ = e.client.DeletePolicyBinding(ctx, lt, lid, pol.ID, groupUUID)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
