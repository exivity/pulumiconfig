package pulumiconfig

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

// mocks is an integer type that will implement methods required for the Pulumi mocking interface.
type mocks int

// NewResource mocks the creation of a new Pulumi resource by returning a mock resource ID and properties.
func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := args.Inputs.Mappable()
	return args.Name + "_id", resource.NewPropertyMapFromMap(outputs), nil
}

// Call mocks Pulumi external calls returning a map with mocked outputs.
func (mocks) Call(_ pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := map[string]interface{}{}
	return resource.NewPropertyMapFromMap(outputs), nil
}

type TestPulumiConfig struct {
	DigitalOcean        TestDigitalOcean         `json:"digital_ocean" overrideConfigNamespace:"pulumi_esc" validate:"required"`
	GrafanaCloud        *TestGrafanaCloud        `json:"grafana_cloud"`
	ProviderCredentials *TestProviderCredentials `json:"provider_credentials" pulumiConfigNamespace:"provider" validate:"required"`
	Enabled             bool                     `json:"enabled"`
	OrgID               int                      `json:"org_id"`
	SubscriptionID      *string                  `json:"subscription_id"`
	Name                string                   `json:"name"`
}

type TestDigitalOcean struct {
	Region string `json:"region" validate:"required,oneof=us-east-1 us-west-1 eu-west-1"`
}

type TestSize struct {
	Size int `json:"size" validate:"sizeValidation=0"`
}

type TestDefaultValue struct {
	DefaultString string  `json:"default_string" validate:"default=DefaultValue"`
	DefaultInt    int     `json:"default_int" validate:"default=100"`
	DefaultUInt   uint    `json:"default_uint" validate:"default=50"`
	DefaultFloat  float32 `json:"default_float" validate:"default=24.24"`
}

type TestGrafanaCloud struct {
	Enabled bool `json:"enabled"`
}

type TestProviderCredentials struct {
	Token        string           `json:"token"`
	GrafanaCloud TestGrafanaCloud `json:"grafana_cloud"`
}

// stringPtr is a utility function to convert a string into a pointer for easier comparison in tests.
func stringPtr(s string) *string {
	return &s
}

// sizeValidation is a custom validation function that ensures the field value is greater than or equal to 10.
func sizeValidation(fl validator.FieldLevel) bool {
	size := fl.Field().Int()
	return size >= 10
}

// nameNotEqualToToken is a struct-level validation function to ensure the Name field isn't the same as the Token field.
func nameNotEqualToToken(sl validator.StructLevel) {
	root := sl.Current().Interface().(TestPulumiConfig)
	if root.Name == root.ProviderCredentials.Token {
		sl.ReportError(root.Name, "Name", "name", "name_eq_token", "")
	}
}

func setPulumiConfig(t *testing.T, config map[string]string) {
	jsonConfig, err := json.Marshal(config)
	assert.NoError(t, err, "Error marshaling to JSON")

	err = os.Setenv(pulumi.EnvConfig, string(jsonConfig))
	assert.NoError(t, err)
}

// TestEnvValue is used to test the env validator.
type TestEnvValue struct {
	EnvString string  `json:"env_string" validate:"env=TEST_ENV_STRING"`
	EnvInt    int     `json:"env_int" validate:"env=TEST_ENV_INT"`
	EnvUInt   uint    `json:"env_uint" validate:"env=TEST_ENV_UINT"`
	EnvFloat  float32 `json:"env_float" validate:"env=TEST_ENV_FLOAT"`
	EnvBool   bool    `json:"env_bool" validate:"env=TEST_ENV_BOOL"`
}

func TestGetConfig(t *testing.T) {
	type args struct {
		obj         interface{}
		validations []Validator
	}

	tests := []struct {
		name    string
		config  map[string]string
		args    args
		envVars map[string]string
		want    interface{}
		wantErr bool
	}{
		{
			name: "all fields are present",
			config: map[string]string{
				"project:digital_ocean":         `{"region":"us-east-1"}`,
				"project:grafana_cloud":         `{"enabled":true}`,
				"provider:provider_credentials": `{"token":"token123", "grafana_cloud": {"enabled":true}}`,
				"project:enabled":               `true`,
				"project:org_id":                `123`,
				"project:subscription_id":       `"sub123"`,
				"project:name":                  `"DeploymentName"`,
			},
			args: args{
				obj: &TestPulumiConfig{},
			},
			want: &TestPulumiConfig{
				DigitalOcean: TestDigitalOcean{Region: "us-east-1"},
				GrafanaCloud: &TestGrafanaCloud{Enabled: true},
				ProviderCredentials: &TestProviderCredentials{
					Token:        "token123",
					GrafanaCloud: TestGrafanaCloud{Enabled: true},
				},
				Enabled:        true,
				OrgID:          123,
				SubscriptionID: stringPtr("sub123"),
				Name:           "DeploymentName",
			},
			wantErr: false,
		},
		{
			name: "only required fields are present",
			config: map[string]string{
				"project:digital_ocean":         `{"region":"us-east-1"}`,
				"provider:provider_credentials": `{"token":"token123", "grafana_cloud": {"enabled":true}}`,
				"project:enabled":               `true`,
				"project:org_id":                `123`,
				"project:name":                  `"DeploymentName"`,
			},
			args: args{
				obj: &TestPulumiConfig{},
			},
			want: &TestPulumiConfig{
				DigitalOcean: TestDigitalOcean{Region: "us-east-1"},
				ProviderCredentials: &TestProviderCredentials{
					Token:        "token123",
					GrafanaCloud: TestGrafanaCloud{Enabled: true},
				},
				Enabled: true,
				OrgID:   123,
				Name:    "DeploymentName",
			},
			wantErr: false,
		},
		{
			name: "required field is missing",
			config: map[string]string{
				"project:grafana_cloud": `{"enabled":true}`,
			},
			args: args{
				obj: &TestPulumiConfig{},
			},
			want:    &TestPulumiConfig{},
			wantErr: true,
		},
		{
			name: "using custom validation with valid value",
			config: map[string]string{
				"project:size": `10`,
			},
			args: args{
				obj: &TestSize{},
				validations: []Validator{
					FieldValidation{
						Tag:      "sizeValidation",
						Validate: sizeValidation,
					},
				},
			},
			want: &TestSize{
				Size: 10,
			},
			wantErr: false,
		},
		{
			name: "using custom validation with invalid value",
			config: map[string]string{
				"project:size": `5`,
			},
			args: args{
				obj: &TestSize{},
				validations: []Validator{
					FieldValidation{
						Tag:      "sizeValidation",
						Validate: sizeValidation,
					},
				},
			},
			want: &TestSize{
				Size: 5,
			},
			wantErr: true,
		},
		{
			name: "struct-level validation with invalid value",
			config: map[string]string{
				"project:digital_ocean":         `{"region":"us-east-1"}`,
				"provider:provider_credentials": `{"token":"DeploymentName", "grafana_cloud": {"enabled":true}}`,
				"project:enabled":               `true`,
				"project:org_id":                `123`,
				"project:name":                  `"DeploymentName"`,
			},
			args: args{
				obj: &TestPulumiConfig{},
				validations: []Validator{
					StructValidation{
						Struct:   &TestPulumiConfig{},
						Validate: nameNotEqualToToken,
					},
				},
			},
			want: &TestPulumiConfig{
				DigitalOcean: TestDigitalOcean{Region: "us-east-1"},
				ProviderCredentials: &TestProviderCredentials{
					Token:        "DeploymentName",
					GrafanaCloud: TestGrafanaCloud{Enabled: true},
				},
				Enabled: true,
				OrgID:   123,
				Name:    "DeploymentName",
			},
			wantErr: true,
		},
		{
			name: "default value is set",
			config: map[string]string{
				"project:default_string": `"some_value"`,
				"project:default_int":    `"10"`,
				"project:default_uint":   `66`,
				"project:default_float":  `12.13`,
			},
			args: args{
				obj: &TestDefaultValue{},
			},
			want: &TestDefaultValue{
				DefaultString: "some_value",
				DefaultInt:    100,
				DefaultUInt:   66,
				DefaultFloat:  12.13,
			},
			wantErr: false,
		},
		{
			name:   "default value is not set",
			config: map[string]string{},
			args: args{
				obj: &TestDefaultValue{},
			},
			want: &TestDefaultValue{
				DefaultString: "DefaultValue",
				DefaultInt:    100,
				DefaultUInt:   50,
				DefaultFloat:  24.24,
			},
			wantErr: false,
		},
		{
			name: "overwrite field value",
			config: map[string]string{
				"project:digital_ocean":         `{"region":"us-east-1"}`,
				"pulumi_esc:digital_ocean":      `{"region":"us-west-1"}`,
				"provider:provider_credentials": `{"token":"token123", "grafana_cloud": {"enabled":true}}`,
				"project:enabled":               `true`,
				"project:org_id":                `123`,
				"project:subscription_id":       `"sub123"`,
				"project:name":                  `"DeploymentName"`,
			},
			args: args{
				obj: &TestPulumiConfig{},
			},
			want: &TestPulumiConfig{
				DigitalOcean: TestDigitalOcean{Region: "us-west-1"},
				ProviderCredentials: &TestProviderCredentials{
					Token:        "token123",
					GrafanaCloud: TestGrafanaCloud{Enabled: true},
				},
				Enabled:        true,
				OrgID:          123,
				SubscriptionID: stringPtr("sub123"),
				Name:           "DeploymentName",
			},
			wantErr: false,
		},
		{
			name: "overwrite field value with missing original",
			config: map[string]string{
				"pulumi_esc:digital_ocean":      `{"region":"us-west-1"}`,
				"provider:provider_credentials": `{"token":"token123", "grafana_cloud": {"enabled":true}}`,
				"project:enabled":               `true`,
				"project:org_id":                `123`,
				"project:subscription_id":       `"sub123"`,
				"project:name":                  `"DeploymentName"`,
			},
			args: args{
				obj: &TestPulumiConfig{},
			},
			want: &TestPulumiConfig{
				DigitalOcean: TestDigitalOcean{Region: "us-west-1"},
				ProviderCredentials: &TestProviderCredentials{
					Token:        "token123",
					GrafanaCloud: TestGrafanaCloud{Enabled: true},
				},
				Enabled:        true,
				OrgID:          123,
				SubscriptionID: stringPtr("sub123"),
				Name:           "DeploymentName",
			},
			wantErr: false,
		},
		{
			name: "overwrite field value with invalid overwrite",
			config: map[string]string{
				"project:digital_ocean":         `{"region":"us-east-1"}`,
				"pulumi_esc:digital_ocean":      `{"region":"invalid-region"}`,
				"provider:provider_credentials": `{"token":"token123", "grafana_cloud": {"enabled":true}}`,
				"project:enabled":               `true`,
				"project:org_id":                `123`,
				"project:subscription_id":       `"sub123"`,
				"project:name":                  `"DeploymentName"`,
			},
			args: args{
				obj: &TestPulumiConfig{},
			},
			want: &TestPulumiConfig{
				DigitalOcean: TestDigitalOcean{Region: "invalid-region"},
				ProviderCredentials: &TestProviderCredentials{
					Token:        "token123",
					GrafanaCloud: TestGrafanaCloud{Enabled: true},
				},
				Enabled:        true,
				OrgID:          123,
				SubscriptionID: stringPtr("sub123"),
				Name:           "DeploymentName",
			},
			wantErr: true,
		},
		{
			name:   "env values are set from environment variables",
			config: map[string]string{},
			args: args{
				obj: &TestEnvValue{},
			},
			envVars: map[string]string{
				"TEST_ENV_STRING": "env_value",
				"TEST_ENV_INT":    "42",
				"TEST_ENV_UINT":   "99",
				"TEST_ENV_FLOAT":  "3.14",
				"TEST_ENV_BOOL":   "true",
			},
			want: &TestEnvValue{
				EnvString: "env_value",
				EnvInt:    42,
				EnvUInt:   99,
				EnvFloat:  3.14,
				EnvBool:   true,
			},
			wantErr: false,
		},
		{
			name:   "env values are not set if env vars are missing",
			config: map[string]string{},
			args: args{
				obj: &TestEnvValue{},
			},
			want:    &TestEnvValue{},
			wantErr: false,
		},
		{
			name: "config value takes precedence over env var",
			config: map[string]string{
				"project:env_string": `"config_value"`,
				"project:env_int":    `100`,
				"project:env_uint":   `123`,
				"project:env_float":  `2.71`,
				"project:env_bool":   `false`,
			},
			args: args{
				obj: &TestEnvValue{},
			},
			want: &TestEnvValue{
				EnvString: "config_value",
				EnvInt:    100,
				EnvUInt:   123,
				EnvFloat:  2.71,
				EnvBool:   false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables before setting new values
			for key := range tt.envVars {
				err := os.Unsetenv(key)
				assert.NoError(t, err, "Error clearing environment variable %s", key)
			}
			for key, value := range tt.envVars {
				err := os.Setenv(key, value)
				assert.NoError(t, err, "Error setting environment variable %s", key)
			}
			t.Cleanup(func() {
				for key := range tt.envVars {
					err := os.Unsetenv(key)
					assert.NoError(t, err, "Error unsetting environment variable %s", key)
				}
			})

			setPulumiConfig(t, tt.config)

			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				err := GetConfig(ctx, tt.args.obj, tt.args.validations...)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, tt.want, tt.args.obj, "Output object doesn't match expected")

				return nil
			},
				pulumi.WithMocks("project", "stack", mocks(0)),
			)
			assert.NoError(t, err, "ValidateConfiguration() failed")
		})
	}
}

func TestCloneStruct(t *testing.T) {
	type args struct {
		src interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "clone simple struct",
			args: args{
				src: TestDigitalOcean{Region: "us-east-1"},
			},
			want: &TestDigitalOcean{Region: "us-east-1"},
		},
		{
			name: "clone struct with pointer field",
			args: args{
				src: &TestPulumiConfig{
					DigitalOcean: TestDigitalOcean{Region: "us-east-1"},
					GrafanaCloud: &TestGrafanaCloud{Enabled: true},
				},
			},
			want: &TestPulumiConfig{
				DigitalOcean: TestDigitalOcean{Region: "us-east-1"},
				GrafanaCloud: &TestGrafanaCloud{Enabled: true},
			},
		},
		{
			name: "clone struct with nested struct",
			args: args{
				src: TestProviderCredentials{
					Token:        "token123",
					GrafanaCloud: TestGrafanaCloud{Enabled: true},
				},
			},
			want: &TestProviderCredentials{
				Token:        "token123",
				GrafanaCloud: TestGrafanaCloud{Enabled: true},
			},
		},
		{
			name: "clone struct with multiple fields",
			args: args{
				src: TestPulumiConfig{
					DigitalOcean:        TestDigitalOcean{Region: "us-east-1"},
					ProviderCredentials: &TestProviderCredentials{Token: "token123"},
					Enabled:             true,
					OrgID:               123,
					SubscriptionID:      stringPtr("sub123"),
					Name:                "DeploymentName",
				},
			},
			want: &TestPulumiConfig{
				DigitalOcean:        TestDigitalOcean{Region: "us-east-1"},
				ProviderCredentials: &TestProviderCredentials{Token: "token123"},
				Enabled:             true,
				OrgID:               123,
				SubscriptionID:      stringPtr("sub123"),
				Name:                "DeploymentName",
			},
		},
		{
			name: "clone struct with default values",
			args: args{
				src: TestDefaultValue{
					DefaultString: "DefaultValue",
					DefaultInt:    100,
					DefaultUInt:   50,
					DefaultFloat:  24.24,
				},
			},
			want: &TestDefaultValue{
				DefaultString: "DefaultValue",
				DefaultInt:    100,
				DefaultUInt:   50,
				DefaultFloat:  24.24,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CloneStruct(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CloneStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}
