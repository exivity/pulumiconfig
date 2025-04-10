// Package pulumitest provides helper functions to facilitate testing
// of Pulumi resources and outputs. This package utilizes testify to
// offer assertions and comparisons tailored to Pulumi's constructs.
//
// Usage:
//
// In your tests, import this package and use its provided functions
// to compare expected and actual Pulumi resources or outputs. The
// functions in this package abstract away the repetitive tasks and
// boilerplate, allowing you to focus on writing meaningful tests
// for your Pulumi programs.
package pulumitest

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

// AssertStringOutputEqual compares two pulumi.StringOutput values
// and uses testify to report if they are not equal.
//
// Usage:
//
//	pulumitest.AssertStringOutputEqual(t, expectedOutput, actualOutput)
func AssertStringOutputEqual(t *testing.T, expected, actual pulumi.Output, msgAndArgs ...interface{}) {
	wg := &sync.WaitGroup{}
	var expectedValue, actualValue interface{}
	wg.Add(2) //nolint:mnd // We need to wait for two goroutines.

	applyFunc := func(output pulumi.Output, target *interface{}) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic while applying output: %v", r)
			}
		}()
		output.ApplyT(func(v interface{}) interface{} {
			defer wg.Done()
			*target = getPointerValue(v)
			return nil
		})
	}

	go applyFunc(expected, &expectedValue)
	go applyFunc(actual, &actualValue)

	wg.Wait()

	msgAndArgs = append(msgAndArgs, "Pulumi string outputs are not equal")
	assert.Equal(t, expectedValue, actualValue, msgAndArgs...)
}

// AssertMapEqual compares two pulumi.Map values and uses testify to report if they are not equal.
//
// Usage:
//
//	pulumitest.AssertMapEqual(t, expectedMap, actualMap)
func AssertMapEqual(t *testing.T, expected, actual pulumi.MapOutput, msgAndArgs ...interface{}) {
	wg := &sync.WaitGroup{}
	var expectedValue, actualValue map[string]interface{}
	wg.Add(2) //nolint:mnd // We need to wait for two goroutines.

	applyFunc := func(output pulumi.MapOutput, target *map[string]interface{}) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic while applying MapOutput: %v", r)
			}
		}()
		output.ApplyT(func(v map[string]interface{}) interface{} {
			defer wg.Done()
			*target = v
			return nil
		})
	}

	go applyFunc(expected, &expectedValue)
	go applyFunc(actual, &actualValue)

	wg.Wait()

	msgAndArgs = append(msgAndArgs, "Pulumi Map outputs are not equal")
	assert.Equal(t, expectedValue, actualValue, msgAndArgs...)
}

// AssertStringMapEqual compares two pulumi.StringMap values and uses testify to report if they are not equal.
//
// Usage:
//
//	pulumitest.AssertStringMapEqual(t, expectedStringMap, actualStringMap)
func AssertStringMapEqual(t *testing.T, expected, actual pulumi.StringMapOutput, msgAndArgs ...interface{}) {
	wg := &sync.WaitGroup{}
	var expectedValue, actualValue map[string]string
	wg.Add(2) //nolint:mnd // We need to wait for two goroutines.

	applyFunc := func(output pulumi.StringMapOutput, target *map[string]string) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic while applying MapOutput: %v", r)
			}
		}()
		output.ApplyT(func(v map[string]string) interface{} {
			defer wg.Done()
			*target = v
			return nil
		})
	}

	go applyFunc(expected, &expectedValue)
	go applyFunc(actual, &actualValue)

	wg.Wait()

	msgAndArgs = append(msgAndArgs, "Pulumi StringMap outputs are not equal")
	assert.Equal(t, expectedValue, actualValue, msgAndArgs...)
}

// AssertArrayEqual compares two pulumi.Array values and uses testify to report if they are not equal.
//
// Usage:
//
//	pulumitest.AssertArrayEqual(t, expectedArray, actualArray)
func AssertArrayEqual(t *testing.T, expected, actual pulumi.ArrayOutput, msgAndArgs ...interface{}) {
	wg := &sync.WaitGroup{}
	var expectedValue, actualValue []interface{}
	wg.Add(2) //nolint:mnd // We need to wait for two goroutines.

	applyFunc := func(output pulumi.ArrayOutput, target *[]interface{}) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic while applying ArrayOutput: %v", r)
			}
		}()
		output.ApplyT(func(v []interface{}) interface{} {
			defer wg.Done()
			*target = v
			return nil
		})
	}

	go applyFunc(expected, &expectedValue)
	go applyFunc(actual, &actualValue)

	wg.Wait()

	msgAndArgs = append(msgAndArgs, "Pulumi Array outputs are not equal")
	assert.Equal(t, expectedValue, actualValue, msgAndArgs...)
}

// getPointerValue dereferences a pointer value until it reaches the base value.
// If the input is not a pointer, the original value is returned.
// For nil pointers, the function will return nil.
func getPointerValue(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			// Return a clear indication for nil pointers.
			return nil
		}
		rv = rv.Elem()
	}
	return rv.Interface()
}

// AssertResourceEqual compares two Pulumi resources and reports any differences using testify.
// It handles Pulumi specific types like pulumi.Output by delegating to specific assert functions.
// Other fields are compared using standard testify assert methods.
func AssertResourceEqual(t *testing.T, expected, actual interface{}, fields []string, msgAndArgs ...interface{}) {
	expectedValue := getPointerValue(expected)
	actualValue := getPointerValue(actual)

	expectedType := reflect.TypeOf(expectedValue)
	actualType := reflect.TypeOf(actualValue)

	assert.Equal(t, expectedType, actualType, "Types of resources are not the same.")

	if expectedType != actualType {
		return
	}

	expectedVal := reflect.ValueOf(expectedValue)
	actualVal := reflect.ValueOf(actualValue)

	compareAll := len(fields) == 0

	// Loop through fields of the struct.
	for i := 0; i < expectedVal.NumField(); i++ {
		fieldName := expectedType.Field(i).Name // get the field's name

		if !compareAll && !slices.Contains(fields, fieldName) {
			continue
		}

		if fieldName == "CustomResourceState" {
			continue
		}

		expectedField := expectedVal.Field(i)
		actualField := actualVal.Field(i)

		// Get the type of each field.
		expectedFieldType := expectedField.Type()
		actualFieldType := actualField.Type()

		// Check if the field is of type pulumi.Output.
		if expectedFieldType.Implements(reflect.TypeOf((*pulumi.Output)(nil)).Elem()) && actualFieldType.Implements(reflect.TypeOf((*pulumi.Output)(nil)).Elem()) {
			expectedOutput := expectedField.Interface().(pulumi.Output)
			actualOutput := actualField.Interface().(pulumi.Output)
			// AssertStringOutputEqual(t, expectedOutput, actualOutput)
			AssertStringOutputEqual(t, expectedOutput, actualOutput, append(msgAndArgs, fmt.Sprintf("Field '%s' mismatch.", fieldName))...)
		} else {
			// If it's not a pulumi.CustomResourceState, use standard testify assertion.
			assert.Equal(t, expectedField.Interface(), actualField.Interface(), fmt.Sprintf("Field '%s' mismatch.", fieldName), msgAndArgs)
		}
	}
}

// SetPulumiConfig sets the Pulumi config for the test.
func SetPulumiConfig(t *testing.T, config map[string]string) {
	jsonConfig, err := json.Marshal(config)
	assert.NoError(t, err, "Error marshaling to JSON")

	err = os.Setenv(pulumi.EnvConfig, string(jsonConfig))
	assert.NoError(t, err)
}
