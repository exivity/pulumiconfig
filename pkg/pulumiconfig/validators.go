package pulumiconfig

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Validation struct {
	ctx *pulumi.Context
}

var (
	// ErrUnsupportedType is returned when the type of a field is not supported by the validator.
	ErrUnsupportedType = errors.New("unsupported type")
)

type ConvertType string

const (
	Int64   ConvertType = "int64"
	Uint64  ConvertType = "uint64"
	Float64 ConvertType = "float64"
)

// string2Number converts a string to a number based on the ConvertType.
func string2Number(s string, t ConvertType) (interface{}, error) {
	switch t {
	case Int64:
		return strconv.ParseInt(s, 10, 64)
	case Uint64:
		return strconv.ParseUint(s, 10, 64)
	case Float64:
		return strconv.ParseFloat(s, 64)
	default:
		return nil, ErrUnsupportedType
	}
}

// GetValidations returns a slice of Validator with all custom validators defined for Pulumi config.
func GetValidations(ctx *pulumi.Context) []Validator {
	v := &Validation{ctx: ctx}
	return []Validator{
		FieldValidation{
			Tag:      "default",
			Validate: v.defaultSetter,
		},
		FieldValidation{
			Tag:      "env",
			Validate: v.envLoader,
		},
	}
}

// defaultSetter is a validator function that sets the field to its default value if it's zero-valued.
// This function is used in conjunction with the `default` tag in struct fields.
func (v *Validation) defaultSetter(fl validator.FieldLevel) bool { //nolint:funlen,cyclop // many switch cases
	// Retrieve the default value from the struct tag.
	defaultValue := fl.Param()

	// If no default value is specified in the tag, just validate as true.
	if defaultValue == "" {
		return true
	}

	field := fl.Field()
	switch field.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Bool:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() == 0 {
			d, err := string2Number(defaultValue, Int64)
			if err != nil {
				v.ctx.Log.Error(fmt.Sprintf("failed to convert default value to int64: %s", err.Error()), nil) //nolint:errcheck // redundant error check
				return false
			}
			field.SetInt(d.(int64))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() == 0 {
			d, err := string2Number(defaultValue, Uint64)
			if err != nil {
				v.ctx.Log.Error(fmt.Sprintf("failed to convert default value to int64: %s", err.Error()), nil) //nolint:errcheck // redundant error check
				return false
			}
			field.SetUint(d.(uint64))
		}
	case reflect.Uintptr:
		return true
	case reflect.Float32, reflect.Float64:
		if field.Float() == 0 {
			d, err := string2Number(defaultValue, Float64)
			if err != nil {
				v.ctx.Log.Error(fmt.Sprintf("failed to convert default value to int64: %s", err.Error()), nil) //nolint:errcheck // redundant error check
				return false
			}
			field.SetFloat(d.(float64))
		}
	case reflect.Complex64:
		return true
	case reflect.Complex128:
		return true
	case reflect.Array:
		return true
	case reflect.Chan:
		return true
	case reflect.Func:
		return true
	case reflect.Interface:
		return true
	case reflect.Map:
		return true
	case reflect.Ptr:
		return true
	case reflect.Slice:
		return true
	case reflect.String:
		if field.String() == "" {
			field.SetString(defaultValue)
		}
	case reflect.Struct:
		return true
	case reflect.UnsafePointer:
		return true
	}

	return true
}

// envLoader is a validator function that sets the field from an environment variable only if it's zero-valued (not set by Pulumi config).
func (v *Validation) envLoader(fl validator.FieldLevel) bool {
	envVar := fl.Param()
	if envVar == "" {
		return true
	}
	field := fl.Field()
	if !field.CanSet() {
		return true
	}
	// Only set from env if the field is still zero-valued (not set by config)
	if !isZeroValue(field) {
		return true
	}
	val, ok := os.LookupEnv(envVar)
	if !ok || val == "" {
		return true
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d, err := string2Number(val, Int64)
		if err == nil {
			field.SetInt(d.(int64))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		d, err := string2Number(val, Uint64)
		if err == nil {
			field.SetUint(d.(uint64))
		}
	case reflect.Float32, reflect.Float64:
		d, err := string2Number(val, Float64)
		if err == nil {
			field.SetFloat(d.(float64))
		}
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err == nil {
			field.SetBool(b)
		}
	}
	return true
}

// isZeroValue checks if a reflect.Value is zero for its type.
func isZeroValue(v reflect.Value) bool {
	zero := reflect.Zero(v.Type())
	return reflect.DeepEqual(v.Interface(), zero.Interface())
}
