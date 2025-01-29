package pulumiconfig

import (
	"fmt"
	"reflect"

	"dario.cat/mergo"
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
	val := reflect.ValueOf(obj)

	// Dereference if obj is a pointer to get the underlying value.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Populate the struct fields with config values.
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		pulumiConfigNamespace := fieldType.Tag.Get("pulumiConfigNamespace")
		cfg := config.New(ctx, pulumiConfigNamespace)

		isRequired := fieldType.Tag.Get("validate") == "required"
		err := populateFieldFromConfig(cfg, jsonTag, val.Field(i))

		overwritePulumiConfigNamespace := fieldType.Tag.Get("overrideConfigNamespace")
		var errOverwrite error
		if overwritePulumiConfigNamespace != "" {
			errOverwrite = overwriteFieldFromOverwriteCfg(ctx, val.Field(i), jsonTag, overwritePulumiConfigNamespace)
		}

		// If this field is required and both attempts (main + overwrite) failed, return error.
		if isRequired && err != nil && errOverwrite != nil {
			return fmt.Errorf("error while reading pulumi config '%s': %w", jsonTag, err)
		}
	}

	// Initialize the validator and register custom validation rules.
	validate := validator.New()
	validators = append(validators, GetValidations(ctx)...) // Assuming GetValidations is defined elsewhere.

	if err := registerValidations(validate, validators); err != nil {
		return err
	}

	// Validate the struct using the initialized validator.
	if err := validate.Struct(obj); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	return nil
}

// populateFieldFromConfig fetches and assigns the configuration value for a given field.
func populateFieldFromConfig(cfg *config.Config, key string, field reflect.Value) error {
	if field.Kind() == reflect.Ptr {
		return cfg.GetObject(key, field.Addr().Interface())
	}

	if err := cfg.TryObject(key, field.Addr().Interface()); err != nil {
		return fmt.Errorf("error while reading pulumi config '%s': %w", key, err)
	}

	return nil
}

// overwriteFieldFromOverwriteCfg handles the overwrite logic, reading from another config namespace
// and merging the result back into the original field.
func overwriteFieldFromOverwriteCfg(ctx *pulumi.Context, field reflect.Value, jsonTag, overwriteNamespace string) error {
	overwriteCfg := config.New(ctx, overwriteNamespace)

	// Create a fresh copy of the field named overwriteV.
	clone := CloneStruct(field.Interface())
	overwriteVal := reflect.ValueOf(clone)
	if overwriteVal.Kind() == reflect.Ptr {
		overwriteVal = overwriteVal.Elem()
	}

	// Now read config values into overwriteVal.
	err := populateFieldFromConfig(overwriteCfg, jsonTag, overwriteVal)
	if err != nil {
		return err
	}

	// Merge the overwritten values back to the original object.
	if mergeErr := mergo.Merge(field.Addr().Interface(), overwriteVal.Addr().Interface(), mergo.WithOverride); mergeErr != nil {
		return mergeErr
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

// CloneStruct uses reflection to create a new instance of the same type
// and copy each exported field's value from src to the new instance.
func CloneStruct(src interface{}) interface{} {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}

	srcType := srcVal.Type()
	dst := reflect.New(srcType).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		if dst.Field(i).CanSet() {
			dst.Field(i).Set(srcVal.Field(i))
		}
	}

	return dst.Addr().Interface()
}
