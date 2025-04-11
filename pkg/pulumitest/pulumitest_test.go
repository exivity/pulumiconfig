// Package pulumitest provides helper functions to test Pulumi outputs.
// These helpers make it easier to compare Pulumi outputs in unit tests.

package pulumitest

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

// TestAssertStringOutputEqual checks if two Pulumi string outputs are the same.
// It ensures that the AssertStringOutputEqual function works as expected.
func TestAssertStringOutputEqual(t *testing.T) {
	type args struct {
		expected   pulumi.Output
		actual     pulumi.Output
		msgAndArgs []interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantFailed bool
	}{
		{
			name: "nil outputs", // Both outputs are empty strings.
			args: args{
				expected: pulumi.String("").ToStringOutput(),
				actual:   pulumi.String("").ToStringOutput(),
			},
			wantFailed: false,
		},
		{
			name: "expect and actual are equal", // Both outputs have the same value.
			args: args{
				expected: pulumi.String("123").ToStringOutput(),
				actual:   pulumi.String("123").ToStringOutput(),
			},
			wantFailed: false,
		},
		{
			name: "expect and actual are not equal", // Outputs have different values.
			args: args{
				expected: pulumi.String("123").ToStringOutput(),
				actual:   pulumi.String("").ToStringOutput(),
			},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testT := &testing.T{}
			AssertStringOutputEqual(testT, tt.args.expected, tt.args.actual, tt.args.msgAndArgs...)
			assert.Equal(t, tt.wantFailed, testT.Failed())
		})
	}
}

// TestAssertMapEqual checks if two Pulumi map outputs are the same.
// It ensures that the AssertMapEqual function works expected.
func TestAssertMapEqual(t *testing.T) { //nolint:dupl // test cases are similar
	type args struct {
		expected   pulumi.MapOutput
		actual     pulumi.MapOutput
		msgAndArgs []interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantFailed bool
	}{
		{
			name: "nil outputs", // Both maps are empty.
			args: args{
				expected: pulumi.Map{}.ToMapOutput(),
				actual:   pulumi.Map{}.ToMapOutput(),
			},
			wantFailed: false,
		},
		{
			name: "expect and actual are equal", // Both maps have the same key-value pairs.
			args: args{
				expected: pulumi.Map{
					"key": pulumi.String("value"),
				}.ToMapOutput(),
				actual: pulumi.Map{
					"key": pulumi.String("value"),
				}.ToMapOutput(),
			},
			wantFailed: false,
		},
		{
			name: "expect and actual are not equal", // Maps have different values for the same key.
			args: args{
				expected: pulumi.Map{
					"key": pulumi.String("value"),
				}.ToMapOutput(),
				actual: pulumi.Map{
					"key": pulumi.String(""),
				}.ToMapOutput(),
			},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testT := &testing.T{}
			AssertMapEqual(testT, tt.args.expected, tt.args.actual, tt.args.msgAndArgs...)
			assert.Equal(t, tt.wantFailed, testT.Failed())
		})
	}
}

// TestAssertStringMapEqual checks if two Pulumi string map outputs are the same.
// It ensures that the AssertStringMapEqual function works correctly.
func TestAssertStringMapEqual(t *testing.T) { //nolint:dupl // test cases are similar
	type args struct {
		expected   pulumi.StringMapOutput
		actual     pulumi.StringMapOutput
		msgAndArgs []interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantFailed bool
	}{
		{
			name: "nil outputs", // Both string maps are empty.
			args: args{
				expected: pulumi.StringMap{}.ToStringMapOutput(),
				actual:   pulumi.StringMap{}.ToStringMapOutput(),
			},
			wantFailed: false,
		},
		{
			name: "expect and actual are equal", // Both string maps have the same key-value pairs.
			args: args{
				expected: pulumi.StringMap{
					"key": pulumi.String("value"),
				}.ToStringMapOutput(),
				actual: pulumi.StringMap{
					"key": pulumi.String("value"),
				}.ToStringMapOutput(),
			},
			wantFailed: false,
		},
		{
			name: "expect and actual are not equal", // String maps have different values for the same key.
			args: args{
				expected: pulumi.StringMap{
					"key": pulumi.String("value"),
				}.ToStringMapOutput(),
				actual: pulumi.StringMap{
					"key": pulumi.String(""),
				}.ToStringMapOutput(),
			},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testT := &testing.T{}
			AssertStringMapEqual(testT, tt.args.expected, tt.args.actual, tt.args.msgAndArgs...)
			assert.Equal(t, tt.wantFailed, testT.Failed())
		})
	}
}

// TestAssertArrayEqual checks if two Pulumi array outputs are the same.
// It ensures that the AssertArrayEqual function works correctly.
func TestAssertArrayEqual(t *testing.T) {
	type args struct {
		expected   pulumi.ArrayOutput
		actual     pulumi.ArrayOutput
		msgAndArgs []interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantFailed bool
	}{
		{
			name: "nil outputs", // Both arrays are empty.
			args: args{
				expected: pulumi.Array{}.ToArrayOutput(),
				actual:   pulumi.Array{}.ToArrayOutput(),
			},
			wantFailed: false,
		},
		{
			name: "expect and actual are equal", // Both arrays have the same values.
			args: args{
				expected: pulumi.Array{
					pulumi.String("value"),
				}.ToArrayOutput(),
				actual: pulumi.Array{
					pulumi.String("value"),
				}.ToArrayOutput(),
			},
			wantFailed: false,
		},
		{
			name: "expect and actual are not equal", // Arrays have different values.
			args: args{
				expected: pulumi.Array{
					pulumi.String("value"),
				}.ToArrayOutput(),
				actual: pulumi.Array{}.ToArrayOutput(),
			},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testT := &testing.T{}
			AssertArrayEqual(testT, tt.args.expected, tt.args.actual, tt.args.msgAndArgs...)
			assert.Equal(t, tt.wantFailed, testT.Failed())
		})
	}
}
