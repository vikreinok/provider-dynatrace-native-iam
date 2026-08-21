package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Package type metadata.
const (
	GroupName = "iam.dynatrace.crossplane.io"
	Version   = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

	// Group type metadata
	GroupKind             = reflect.TypeOf(Group{}).Name()
	GroupGroupKind        = schema.GroupKind{Group: GroupName, Kind: GroupKind}.String()
	GroupKindAPIVersion   = GroupKind + "." + SchemeGroupVersion.String()
	GroupGroupVersionKind = SchemeGroupVersion.WithKind(GroupKind)

	// Policy type metadata
	PolicyKind             = reflect.TypeOf(Policy{}).Name()
	PolicyGroupKind        = schema.GroupKind{Group: GroupName, Kind: PolicyKind}.String()
	PolicyKindAPIVersion   = PolicyKind + "." + SchemeGroupVersion.String()
	PolicyGroupVersionKind = SchemeGroupVersion.WithKind(PolicyKind)

	// PolicyBoundary type metadata
	PolicyBoundaryKind             = reflect.TypeOf(PolicyBoundary{}).Name()
	PolicyBoundaryGroupKind        = schema.GroupKind{Group: GroupName, Kind: PolicyBoundaryKind}.String()
	PolicyBoundaryKindAPIVersion   = PolicyBoundaryKind + "." + SchemeGroupVersion.String()
	PolicyBoundaryGroupVersionKind = SchemeGroupVersion.WithKind(PolicyBoundaryKind)

	// PolicyBindingsV2 type metadata
	PolicyBindingsV2Kind             = reflect.TypeOf(PolicyBindingsV2{}).Name()
	PolicyBindingsV2GroupKind        = schema.GroupKind{Group: GroupName, Kind: PolicyBindingsV2Kind}.String()
	PolicyBindingsV2KindAPIVersion   = PolicyBindingsV2Kind + "." + SchemeGroupVersion.String()
	PolicyBindingsV2GroupVersionKind = SchemeGroupVersion.WithKind(PolicyBindingsV2Kind)

	// CostCenter type metadata
	CostCenterKind             = reflect.TypeOf(CostCenter{}).Name()
	CostCenterGroupKind        = schema.GroupKind{Group: GroupName, Kind: CostCenterKind}.String()
	CostCenterKindAPIVersion   = CostCenterKind + "." + SchemeGroupVersion.String()
	CostCenterGroupVersionKind = SchemeGroupVersion.WithKind(CostCenterKind)

	// User type metadata
	UserKind             = reflect.TypeOf(User{}).Name()
	UserGroupKind        = schema.GroupKind{Group: GroupName, Kind: UserKind}.String()
	UserKindAPIVersion   = UserKind + "." + SchemeGroupVersion.String()
	UserGroupVersionKind = SchemeGroupVersion.WithKind(UserKind)

	// ServiceUser type metadata
	ServiceUserKind             = reflect.TypeOf(ServiceUser{}).Name()
	ServiceUserGroupKind        = schema.GroupKind{Group: GroupName, Kind: ServiceUserKind}.String()
	ServiceUserKindAPIVersion   = ServiceUserKind + "." + SchemeGroupVersion.String()
	ServiceUserGroupVersionKind = SchemeGroupVersion.WithKind(ServiceUserKind)
)

func init() {
	SchemeBuilder.Register(
		&Group{}, &GroupList{},
		&Policy{}, &PolicyList{},
		&PolicyBoundary{}, &PolicyBoundaryList{},
		&PolicyBindingsV2{}, &PolicyBindingsV2List{},
		&CostCenter{}, &CostCenterList{},
		&User{}, &UserList{},
		&ServiceUser{}, &ServiceUserList{},
	)
}
