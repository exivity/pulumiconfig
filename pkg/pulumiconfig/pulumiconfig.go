package pulumiconfig

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Validator is an interface that wraps the Register method,
// providing a standardized way to register different types of validations.
type Validator interface {
	Register(validate *validator.Validate) error
}

// FieldValidation holds the information required for field-level validation.
type FieldValidation struct {
	Tag      string                             // The tag name used in struct fields for validation.
	Validate func(fl validator.FieldLevel) bool // The actual field validation function.
}

// StructValidation holds the information required for struct-level validation.
type StructValidation struct {
	Struct   interface{}                    // The struct type that the validation will apply to.
	Validate func(sl validator.StructLevel) // The actual struct validation function.
}

// Register adds the field validation function to the provided validator instance.
func (fv FieldValidation) Register(validate *validator.Validate) error {
	return validate.RegisterValidation(fv.Tag, fv.Validate)
}

// Register adds the struct validation function to the provided validator instance.
func (sv StructValidation) Register(validate *validator.Validate) error {
	validate.RegisterStructValidation(sv.Validate, sv.Struct)
	return nil
}

// GetConfig retrieves configuration values from the Pulumi project and populates the provided object.
// It also runs any associated validations to ensure the configuration's integrity.
func GetConfig(ctx *pulumi.Context, obj interface{}, validators ...Validator) error {
	v := reflect.ValueOf(obj)

	// Dereference if obj is a pointer to get the underlying value.
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Iterate over each field in the struct and fetch its configuration.
	for i := 0; i < v.NumField(); i++ {
		fieldType := v.Type().Field(i)
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		pulumiConfigNamespace := fieldType.Tag.Get("pulumiConfigNamespace")
		cfg := config.New(ctx, pulumiConfigNamespace)

		isRequired := fieldType.Tag.Get("validate") == "required"
		if err := getConfigValue(cfg, jsonTag, v.Field(i), isRequired); err != nil {
			return err
		}
	}

	// Initialize the validator and register custom validation rules.
	validate := validator.New()
	validators = append(validators, GetValidations(ctx)...)
	if err := registerValidations(validate, validators); err != nil {
		return err
	}

	// Validate the struct using the initialized validator.
	if err := validate.Struct(obj); err != nil {
		return fmt.Errorf("Validation error: %w", err)
	}

	return nil
}

// getConfigValue fetches the configuration value based on its type and if it's a required field.
func getConfigValue(cfg *config.Config, jsonTag string, field reflect.Value, isRequired bool) error {
	if field.Kind() == reflect.Ptr {
		return cfg.GetObject(jsonTag, field.Addr().Interface())
	} else if err := cfg.TryObject(jsonTag, field.Addr().Interface()); err != nil && isRequired {
		return fmt.Errorf("Error while reading pulumi config `%s`: %w", jsonTag, err)
	}
	return nil
}

// registerValidations registers all provided validators to the provided validator instance.
func registerValidations(validate *validator.Validate, validators []Validator) error {
	for _, v := range validators {
		if err := v.Register(validate); err != nil {
			return err
		}
	}
	return nil
}
