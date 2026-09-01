package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Package type metadata.
const (
	GroupName = "management.dynatrace.crossplane.io"
	Version   = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

	// ZoneV2 type metadata
	ZoneV2Kind             = reflect.TypeOf(ZoneV2{}).Name()
	ZoneV2GroupKind        = schema.GroupKind{Group: GroupName, Kind: ZoneV2Kind}.String()
	ZoneV2KindAPIVersion   = ZoneV2Kind + "." + SchemeGroupVersion.String()
	ZoneV2GroupVersionKind = SchemeGroupVersion.WithKind(ZoneV2Kind)
)

func init() {
	SchemeBuilder.Register(
		&ZoneV2{}, &ZoneV2List{},
	)
}
