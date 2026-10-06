package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Package type metadata.
const (
	GroupName = "host.dynatrace.crossplane.io"
	Version   = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

	// HostEntity type metadata
	HostEntityKind             = reflect.TypeOf(HostEntity{}).Name()
	HostEntityGroupKind        = schema.GroupKind{Group: GroupName, Kind: HostEntityKind}.String()
	HostEntityKindAPIVersion   = HostEntityKind + "." + SchemeGroupVersion.String()
	HostEntityGroupVersionKind = SchemeGroupVersion.WithKind(HostEntityKind)
)

func init() {
	SchemeBuilder.Register(
		&HostEntity{}, &HostEntityList{},
	)
}
