/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/config"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/iam/costcenter"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/iam/group"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/iam/policy"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/iam/policybindingsv2"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/iam/policyboundary"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/iam/serviceuser"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/iam/user"
)

// SetupGated creates all Dynatrace controllers with safe-start support and adds them to
// the supplied manager.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		config.Setup,
		group.SetupGated,
		policy.SetupGated,
		policyboundary.SetupGated,
		policybindingsv2.SetupGated,
		costcenter.SetupGated,
		user.SetupGated,
		serviceuser.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
