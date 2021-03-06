/*


Licensed under the Mozilla Public License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.mozilla.org/en-US/MPL/2.0/

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var apidefinitionlog = logf.Log.WithName("apidefinition-resource")

func (in *ApiDefinition) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-tyk-tyk-io-v1alpha1-apidefinition,mutating=true,failurePolicy=fail,groups=tyk.tyk.io,resources=apidefinitions,verbs=create;update,versions=v1alpha1,name=mapidefinition.kb.io,sideEffects=None

var _ webhook.Defaulter = &ApiDefinition{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *ApiDefinition) Default() {
	apidefinitionlog.Info("default", "name", in.Name)

	// We disable tracking by default
	if in.Spec.DoNotTrack == nil {
		in.Spec.DoNotTrack = pointer.BoolPtr(true)
	}

	if len(in.Spec.VersionData.Versions) == 0 {
		defaultVersionData := VersionData{
			NotVersioned:   true,
			DefaultVersion: "Default",
			Versions: map[string]VersionInfo{
				"Default": {
					Name:    "Default",
					Expires: "",
					Paths: VersionInfoPaths{
						Ignored:   []string{},
						WhiteList: []string{},
						BlackList: []string{},
					},
					UseExtendedPaths:            false,
					ExtendedPaths:               ExtendedPathsSet{},
					GlobalHeaders:               nil,
					GlobalHeadersRemove:         nil,
					GlobalResponseHeaders:       nil,
					GlobalResponseHeadersRemove: nil,
					IgnoreEndpointCase:          false,
					GlobalSizeLimit:             0,
				},
			},
		}

		in.Spec.VersionData = defaultVersionData
	}

	if in.Spec.UseStandardAuth {
		if in.Spec.AuthConfigs == nil {
			in.Spec.AuthConfigs = make(map[string]AuthConfig)
		}
		if _, ok := in.Spec.AuthConfigs["authToken"]; !ok {
			apidefinitionlog.Info("applying default auth_config as not set & use_standard_auth enabled")
			in.Spec.AuthConfigs["authToken"] = AuthConfig{
				AuthHeaderName: "Authorization",
			}
		}
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-tyk-tyk-io-v1alpha1-apidefinition,mutating=false,failurePolicy=fail,groups=tyk.tyk.io,resources=apidefinitions,versions=v1alpha1,name=vapidefinition.kb.io,sideEffects=None

var _ webhook.Validator = &ApiDefinition{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateCreate() error {
	apidefinitionlog.Info("validate create", "name", in.Name)

	err := validateAuth(in)
	if err != nil {
		return err
	}

	err = validateUDG(in)
	if err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateUpdate(old runtime.Object) error {
	apidefinitionlog.Info("validate update", "name", in.Name)

	err := validateAuth(in)
	if err != nil {
		return err
	}

	err = validateUDG(in)
	if err != nil {
		return err
	}

	return nil
}

func validateAuth(in *ApiDefinition) error {
	if in.Spec.UseKeylessAccess {
		if in.Spec.UseStandardAuth {
			return errors.New("conflict: cannot use_keyless_access & use_standard_auth")
		}
	}
	return nil
}

// TODO: proper udg validation required here
func validateUDG(in *ApiDefinition) error {
	if in.Spec.GraphQL == nil {
		return nil
	}

	if in.Spec.GraphQL.Enabled && in.Spec.GraphQL.ExecutionMode == "executionEngine" {
		for _, typeFieldConfig := range in.Spec.GraphQL.TypeFieldConfigurations {
			switch typeFieldConfig.DataSource.Kind {
			case "HTTPJsonDataSource":
				if typeFieldConfig.DataSource.Config.URL == "" ||
					typeFieldConfig.DataSource.Config.Method == "" {
					return errors.New("URL or Method missing for HTTPJsonDataSource")
				}
			case "GraphQLDataSource":
				if typeFieldConfig.DataSource.Config.URL == "" ||
					typeFieldConfig.DataSource.Config.Method == "" {
					return errors.New("URL or Method missing for GraphQLDataSource")
				}
			}
		}
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateDelete() error {
	apidefinitionlog.Info("validate delete", "name", in.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
