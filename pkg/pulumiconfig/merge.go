package pulumiconfig

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNilObjects       = errors.New("both objects must be non-nil")
	ErrDifferentTypes   = errors.New("objects must be of the same type")
	ErrNonPointer       = errors.New("both objects must be pointers")
	ErrMismatchedFields = errors.New("mismatched field types")
	ErrFieldNotSettable = errors.New("field is not settable")
)

func mergeObjects(obj1, obj2 interface{}) (interface{}, error) {
	if obj1 == nil || obj2 == nil {
		return nil, fmt.Errorf("%w", ErrNilObjects)
	}

	typ1 := reflect.TypeOf(obj1)
	typ2 := reflect.TypeOf(obj2)

	if typ1 != typ2 {
		return nil, fmt.Errorf("%w", ErrDifferentTypes)
	}

	if typ1.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("%w", ErrNonPointer)
	}

	// Create a new instance of the same type
	newObj := reflect.New(typ1.Elem()).Interface()

	val1 := reflect.ValueOf(obj1).Elem()
	val2 := reflect.ValueOf(obj2).Elem()
	newVal := reflect.ValueOf(newObj).Elem()

	if err := mergeFields(val1, val2, newVal); err != nil {
		return nil, err
	}

	return newObj, nil
}

func mergeFields(val1, val2, newVal reflect.Value) error {
	if val1.Type() != val2.Type() || val1.Type() != newVal.Type() {
		return fmt.Errorf("%w", ErrMismatchedFields)
	}

	for i := 0; i < val1.NumField(); i++ {
		field1 := val1.Field(i)
		field2 := val2.Field(i)
		newField := newVal.Field(i)

		if !newField.CanSet() {
			return fmt.Errorf("%w: field %d", ErrFieldNotSettable, i)
		}

		switch field1.Kind() {
		case reflect.Struct:
			// If the field is a struct, merge recursively
			nestedNewVal := reflect.New(field1.Type()).Elem()
			if err := mergeFields(field1, field2, nestedNewVal); err != nil {
				return err
			}
			newField.Set(nestedNewVal)
		default:
			// Use value from obj2 if it is non-zero, otherwise use value from obj1
			if !isZeroValue(field2) {
				newField.Set(field2)
			} else {
				newField.Set(field1)
			}
		}
	}
	return nil
}

func isZeroValue(value reflect.Value) bool {
	zero := reflect.Zero(value.Type())
	return reflect.DeepEqual(value.Interface(), zero.Interface())
}
